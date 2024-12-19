package obirefidx

import (
	"fmt"
	"math"
	"os"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obichunk"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

func MakeStartClusterSliceWorker(clusterslot string, threshold float64) obiseq.SeqSliceWorker {

	clusteridslot := fmt.Sprintf("%s_clusterid", clusterslot)
	clusterheadslot := fmt.Sprintf("%s_clusterhead", clusterslot)
	cludteridentityslot := fmt.Sprintf("%s_clusteridentity", clusterslot)
	clusternclusterslot := fmt.Sprintf("%s_cluster_n", clusterslot)
	StartClusterSliceWorkers := func(sequences obiseq.BioSequenceSlice) (obiseq.BioSequenceSlice, error) {
		sequences.SortOnCount(true)
		n := 0
		for i, seqA := range sequences {
			if !seqA.HasAttribute(clusteridslot) {
				n++
				seqA.SetAttribute(clusteridslot, seqA.Id())
				seqA.SetAttribute(clusterheadslot, true)
				seqA.SetAttribute(cludteridentityslot, 1.0)
				for j := i + 1; j < len(sequences); j++ {
					seqB := sequences[j]
					if !seqB.HasAttribute(clusteridslot) {
						errmax := int(math.Ceil(float64(max(seqA.Len(), seqB.Len())) * (1.0 - threshold) * 2))
						lca, alilen := obialign.FastLCSScore(seqA, seqB, errmax, nil)
						id := float64(lca) / float64(alilen)

						if lca >= 0 && id >= threshold {
							seqB.SetAttribute(clusteridslot, seqA.Id())
							seqB.SetAttribute(clusterheadslot, false)
							seqB.SetAttribute(cludteridentityslot, id)
						}
					}
				}
			}

		}

		log.Debugf("Clustered %d sequences into %d clusters", len(sequences), n)

		for _, seq := range sequences {
			seq.SetAttribute(clusternclusterslot, n)
		}

		return sequences, nil
	}

	return StartClusterSliceWorkers
}

func MakeIndexingSliceWorker(indexslot, idslot string,
	kmers *[]*obikmer.Table4mer,
	taxonomy *obitax.Taxonomy,
) obiseq.SeqSliceWorker {
	IndexingSliceWorkers := func(sequences obiseq.BioSequenceSlice) (obiseq.BioSequenceSlice, error) {

		kmercounts := make(
			[]*obikmer.Table4mer,
			len(sequences))

		taxa := taxonomy.NewTaxonSlice(sequences.Len(), sequences.Len())

		for i, seq := range sequences {
			j, ok := seq.GetIntAttribute(idslot)

			if !ok {
				return nil, fmt.Errorf("sequence %s has no %s attribute", seq.Id(), idslot)
			}

			kmercounts[i] = (*kmers)[j]
			taxon := seq.Taxon(taxonomy)
			if taxon == nil {
				log.Panicf("%s: Cannot get the taxon %s in %s", seq.Id(), seq.Taxid(), taxonomy.Name())
			}
			taxa.Set(i, taxon)

		}

		limits := make(chan [2]int)
		waiting := sync.WaitGroup{}

		go func() {
			for i := 0; i < len(sequences); i += 10 {
				limits <- [2]int{i, min(i+10, len(sequences))}
			}
			close(limits)
		}()

		f := func() {
			for l := range limits {
				for i := l[0]; i < l[1]; i++ {
					idx := IndexSequence(i, sequences, &kmercounts, taxa, taxonomy)
					sequences[i].SetAttribute(indexslot, idx)
				}
			}

			waiting.Done()
		}

		nworkers := max(min(obioptions.CLIParallelWorkers(), len(sequences)/10), 1)

		waiting.Add(nworkers)

		for w := 0; w < nworkers; w++ {
			go f()
		}

		waiting.Wait()

		return sequences, nil
	}

	return IndexingSliceWorkers
}

func IndexFamilyDB(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	log.Infoln("Family level reference database indexing...")
	log.Infoln("Loading database...")
	source, references := iterator.Load()
	nref := len(references)
	log.Infof("Done. Database contains %d sequences", nref)

	taxonomy, error := obioptions.CLILoadSelectedTaxonomy()
	if error != nil {
		log.Panicln(error)
	}

	log.Infoln("Indexing database kmers...")

	refcounts := make(
		[]*obikmer.Table4mer,
		len(references))

	buffer := make([]byte, 0, 1000)

	for i, seq := range references {
		seq.SetAttribute("reffamidx_id", i)
		refcounts[i] = obikmer.Count4Mer(seq, &buffer, nil)
	}

	log.Info("done")

	partof := obiiter.IBatchOver(source, references,
		obioptions.CLIBatchSize()).MakeIWorker(obiseq.MakeSetSpeciesWorker(taxonomy),
		false,
		obioptions.CLIParallelWorkers(),
	).MakeIWorker(obiseq.MakeSetGenusWorker(taxonomy),
		false,
		obioptions.CLIParallelWorkers(),
	).MakeIWorker(obiseq.MakeSetFamilyWorker(taxonomy),
		false,
		obioptions.CLIParallelWorkers(),
	)

	family_iterator, err := obichunk.ISequenceChunk(
		partof,
		obiseq.AnnotationClassifier("family_taxid", "NA"),
	)

	if err != nil {
		log.Fatal(err)
	}

	family_iterator.MakeISliceWorker(
		MakeStartClusterSliceWorker("reffamidx", 0.9),
		false,
		obioptions.CLIParallelWorkers(),
	).MakeISliceWorker(
		MakeIndexingSliceWorker("reffamidx_in", "reffamidx_id", &refcounts, taxonomy),
		false,
		obioptions.CLIParallelWorkers(),
	).Speed("Family Indexing", nref).Consume()

	clusters := obiseq.MakeBioSequenceSlice(0)
	kcluster := make([]*obikmer.Table4mer, 0)
	taxa := taxonomy.NewTaxonSlice(references.Len(), references.Len())

	j := 0
	for i, seq := range references {
		if is_centrer, _ := seq.GetBoolAttribute("reffamidx_clusterhead"); is_centrer {
			clusters = append(clusters, seq)
			kcluster = append(kcluster, refcounts[i])

			taxon := seq.Taxon(taxonomy)
			if taxon == nil {
				log.Panicf("%s: Cannot get the taxon %s in %s", seq.Id(), seq.Taxid(), taxonomy.Name())
			}
			taxa.Set(j, taxon)

			j++
		}
	}

	log.Infof("Done. Found %d clusters", clusters.Len())

	pbopt := make([]progressbar.Option, 0, 5)
	pbopt = append(pbopt,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetDescription("Cluster indexing"),
	)

	bar := progressbar.NewOptions(len(clusters), pbopt...)

	limits := make(chan [2]int)
	waiting := sync.WaitGroup{}

	go func() {
		for i := 0; i < len(clusters); i += 10 {
			limits <- [2]int{i, min(i+10, len(clusters))}
		}
		close(limits)
	}()

	f := func() {
		for l := range limits {
			for i := l[0]; i < l[1]; i++ {
				idx := IndexSequence(i, clusters, &kcluster, taxa, taxonomy)
				clusters[i].SetOBITagRefIndex(idx)
				bar.Add(1)
			}
		}

		waiting.Done()
	}

	nworkers := obioptions.CLIParallelWorkers()
	waiting.Add(nworkers)

	for w := 0; w < nworkers; w++ {
		go f()
	}

	waiting.Wait()

	results := obiiter.IBatchOver(source, references,
		obioptions.CLIBatchSize()).Speed("Writing db", nref)

	return results
}

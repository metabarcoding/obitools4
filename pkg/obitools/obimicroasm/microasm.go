package obimicroasm

import (
	"fmt"
	"os"
	"path"
	"slices"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obisuffix"
)

func BuildFilterOnPatternReadPairWorker(
	forward, reverse string,
	errormax int,
	cutReads bool,
) obiseq.SeqWorker {
	forwardPatternDir, err := obiapat.MakeApatPattern(forward, errormax, false)

	if err != nil {
		log.Fatalf("Cannot compile forward primer %s : %v", forward, err)
	}

	reverse_rev := obiseq.NewBioSequence("fp", []byte(reverse), "").ReverseComplement(true).String()
	reveresePatternRev, err := obiapat.MakeApatPattern(reverse_rev, errormax, false)

	if err != nil {
		log.Fatalf("Cannot compile reverse complement reverse primer %s : %v", reverse, err)
	}

	matchRead := func(sequence *obiseq.BioSequence) *obiseq.BioSequence {
		var aseq obiapat.ApatSequence
		var err error
		var read, match *obiseq.BioSequence

		aseq, err = obiapat.MakeApatSequence(sequence, false)

		if err != nil {
			log.Fatalf("Cannot prepare apat sequence from %s : %v", sequence.Id(), err)
		}

		start, end, nerr, matched := forwardPatternDir.BestMatch(aseq, 0, aseq.Len())

		if matched {
			read = sequence

			if cutReads {
				read, err = sequence.Subsequence(start, sequence.Len(), false)

				if err != nil {
					log.Fatalf("Cannot cut, on forward, forward read %s [%d,%d] : %v",
						sequence.Id(), start, sequence.Len(), err)
				}
			}

			read.SetAttribute("forward_primer", forward)
			match, _ = sequence.Subsequence(start, end, false)
			read.SetAttribute("forward_match", match.String())
			read.SetAttribute("forward_error", nerr)

			aseq, err = obiapat.MakeApatSequence(read, false, aseq)

			if err != nil {
				log.Fatalf("Cannot prepare apat sequence from %s : %v", sequence.Id(), err)
			}

			start, end, nerr, matched = reveresePatternRev.BestMatch(aseq, 0, aseq.Len())

			if matched {

				frread := read

				if cutReads {
					frread, err = read.Subsequence(0, end, false)

					if err != nil {
						log.Fatalf("Cannot xxx cut, on reverse, forward read %s [%d,%d] : %v",
							sequence.Id(), start, read.Len(), err)
					}
				}

				frread.SetAttribute("reverse_primer", reverse)
				match, _ = read.Subsequence(start, end, false)
				frread.SetAttribute("reverse_match", match.ReverseComplement(true).String())
				frread.SetAttribute("reverse_error", nerr)

				read = frread
				//				log.Warnf("Forward-Reverse primer matched on %s : %d\n%s", read.Id(), read.Len(),
				//					obiformats.FormatFasta(read, obiformats.FormatFastSeqJsonHeader))
			}

		} else {
			start, end, nerr, matched = reveresePatternRev.BestMatch(aseq, 0, aseq.Len())

			if matched {
				read = sequence
				if cutReads {
					read, err = sequence.Subsequence(0, end, false)

					if err != nil {
						log.Fatalf("Cannot yyy cut, on reverse, forward read %s [%d,%d] : %v",
							sequence.Id(), 0, end, err)
					}

				}

				read.SetAttribute("reverse_primer", reverse)
				match, _ = read.Subsequence(start, end, false)
				read.SetAttribute("reverse_match", match.ReverseComplement(true).String())
				read.SetAttribute("reverse_error", nerr)
			} else {
				read = nil
			}

		}

		return read
	}

	w := func(sequence *obiseq.BioSequence) (result obiseq.BioSequenceSlice, err error) {
		result = obiseq.MakeBioSequenceSlice()

		paired := sequence.PairedWith()
		sequence.UnPair()

		read := matchRead(sequence)

		if read == nil {
			sequence = sequence.ReverseComplement(true)
			read = matchRead(sequence)
		}

		if read != nil {
			result = append(result, read)
		}

		if paired != nil {
			read = matchRead(paired)

			if read == nil {
				read = matchRead(paired.ReverseComplement(true))
			}

			if read != nil {
				result = append(result, read)
			}
		}

		return
	}

	return w
}

func ExtractOnPatterns(iter obiiter.IBioSequence,
	forward, reverse string,
	errormax int,
	cutReads bool,
) obiseq.BioSequenceSlice {

	matched := iter.MakeIWorker(
		BuildFilterOnPatternReadPairWorker(forward, reverse, errormax, cutReads),
		false,
	)

	rep := obiseq.MakeBioSequenceSlice()

	for matched.Next() {
		frgs := matched.Get()
		rep = append(rep, frgs.Slice()...)
	}

	return rep
}

func BuildPCRProduct(seqs obiseq.BioSequenceSlice,
	consensus_id string,
	kmer_size int,
	forward, reverse string,
	backtrack bool,
	save_graph bool, dirname string) (*obiseq.BioSequence, error) {

	from := obiseq.NewBioSequence("forward", []byte(forward), "")
	to := obiseq.NewBioSequence("reverse", []byte(CLIReversePrimer()), "").ReverseComplement(true)

	if backtrack {
		from, to = to, from
	}

	if seqs.Len() == 0 {
		return nil, fmt.Errorf("no sequence provided")
	}

	if save_graph {
		if dirname == "" {
			dirname = "."
		}

		if stat, err := os.Stat(dirname); err != nil || !stat.IsDir() {
			// path does not exist or is not directory
			os.RemoveAll(dirname)
			err := os.Mkdir(dirname, 0755)

			if err != nil {
				log.Panicf("Cannot create directory %s for saving graphs", dirname)
			}
		}

		fasta, err := os.Create(path.Join(dirname, fmt.Sprintf("%s_consensus.fasta", consensus_id)))

		if err == nil {
			defer fasta.Close()
			fasta.Write(obiformats.FormatFastaBatch(obiiter.MakeBioSequenceBatch(
				fmt.Sprintf("%s_consensus", consensus_id),
				0,
				seqs,
			),
				obiformats.FormatFastSeqJsonHeader, false).Bytes())
			fasta.Close()
		}

	}

	log.Debugf("Number of reads : %d\n", len(seqs))

	if kmer_size < 0 {
		longest := make([]int, len(seqs))

		for i, seq := range seqs {
			s := obiseq.BioSequenceSlice{seq}
			sa := obisuffix.BuildSuffixArray(&s)
			longest[i] = slices.Max(sa.CommonSuffix())
		}

		// spectrum := map[int]int{}
		// for _, s := range longest {
		// 	spectrum[s]++
		// }

		// log.Warnf("spectum kmer size : %v", spectrum)

		kmer_size = slices.Max(longest) + 1
		log.Infof("estimated kmer size : %d", kmer_size)
	}

	var graph *obikmer.DeBruijnGraph

	var hp []uint64
	var err error
	var starts []uint64
	var stops []uint64

	for {
		graph = obikmer.MakeDeBruijnGraph(kmer_size)

		for _, s := range seqs {
			graph.Push(s)
		}

		if !backtrack {
			starts = graph.Search(from, CLIAllowedMismatch())
			stops = graph.BackSearch(to, CLIAllowedMismatch())
		} else {
			starts = graph.BackSearch(from, CLIAllowedMismatch())
			stops = graph.Search(to, CLIAllowedMismatch())
		}

		log.Infof("Found %d starts", len(starts))
		pweight := map[int]int{}
		for _, s := range starts {
			w := graph.Weight(s)
			pweight[w]++
			log.Warnf("Starts : %s (%d)\n", graph.DecodeNode(s), w)
		}

		log.Infof("Found %d stops", len(stops))
		for _, s := range stops {
			w := graph.Weight(s)
			pweight[w]++
			log.Warnf("Stop : %s (%d)\n", graph.DecodeNode(s), w)
		}

		log.Infof("Weight spectrum : %v", pweight)

		wmax := 0
		sw := 0
		for w := range pweight {
			sw += w
			if w > wmax {
				wmax = w
			}
		}

		graph.FilterMinWeight(int(sw / len(pweight)))
		graph.FilterMaxWeight(int(wmax * 2))

		log.Infof("Minimum coverage : %d", int(sw/len(pweight)))
		log.Infof("Maximum coverage : %d", int(wmax*2))

		if !graph.HasCycleInDegree() {
			break
		}

		kmer_size++

		if kmer_size > 31 {
			break
		}

		SetKmerSize(kmer_size)
		log.Warnf("Cycle detected, increasing kmer size to %d\n", kmer_size)
	}

	if !backtrack {
		starts = graph.Search(from, CLIAllowedMismatch())
		stops = graph.BackSearch(to, CLIAllowedMismatch())
	} else {
		starts = graph.BackSearch(from, CLIAllowedMismatch())
		stops = graph.Search(to, CLIAllowedMismatch())
	}

	hp, err = graph.HaviestPath(starts, stops, backtrack)

	log.Debugf("Graph size : %d\n", graph.Len())

	maxw := graph.MaxWeight()
	modew := graph.WeightMode()
	meanw := graph.WeightMean()
	specw := graph.WeightSpectrum()
	kmer := graph.KmerSize()

	log.Warnf("Weigh mode: %d Weigth mean : %4.1f Weigth max : %d, kmer = %d", modew, meanw, maxw, kmer)
	log.Warn(specw)

	if save_graph {

		file, err := os.Create(path.Join(dirname,
			fmt.Sprintf("%s_consensus.gml", consensus_id)))

		if err != nil {
			fmt.Println(err)
		} else {
			file.WriteString(graph.Gml())
			file.Close()
		}
	}

	if err == nil {
		s := graph.DecodePath(hp)

		seq := obiseq.NewBioSequence(consensus_id, []byte(s), "")

		total_kmer := graph.Len()
		sumCount := 0

		if seq != nil {
			for _, s := range seqs {
				sumCount += s.Count()
			}
			seq.SetAttribute("obiconsensus_consensus", true)
			seq.SetAttribute("obiconsensus_weight", sumCount)
			seq.SetAttribute("obiconsensus_seq_length", seq.Len())
			seq.SetAttribute("obiconsensus_kmer_size", kmer_size)
			seq.SetAttribute("obiconsensus_kmer_max_occur", graph.MaxWeight())
			seq.SetAttribute("obiconsensus_filtered_graph_size", graph.Len())
			seq.SetAttribute("obiconsensus_full_graph_size", total_kmer)
		}

		log.Warnf("Consensus sequence : \n%s", obiformats.FormatFasta(seq, obiformats.FormatFastSeqJsonHeader))

		return seq, nil

	}

	return nil, err
}

func CLIAssemblePCR() *obiseq.BioSequence {

	pairs, err := CLIPairedSequence()

	if err != nil {
		log.Errorf("Cannot open file (%v)", err)
		os.Exit(1)
	}

	matched := ExtractOnPatterns(pairs,
		CLIForwardPrimer(),
		CLIReversePrimer(),
		CLIAllowedMismatch(),
		true,
	)

	seq, err := BuildPCRProduct(
		matched,
		CLIGraphFilesDirectory(),
		CLIKmerSize(),
		CLIForwardPrimer(),
		CLIReversePrimer(),
		false,
		CLISaveGraphToFiles(),
		CLIGraphFilesDirectory())

	if err != nil {
		log.Fatalf("Cannot build the consensus sequence : %v", err)

	}

	forwardPatternDir, err := obiapat.MakeApatPattern(
		CLIForwardPrimer(),
		CLIAllowedMismatch(),
		false)

	if err != nil {
		log.Fatalf("Cannot compile forward primer %s : %v", CLIForwardPrimer(), err)
	}

	reverse_rev := obiseq.NewBioSequence("fp", []byte(CLIReversePrimer()), "").ReverseComplement(true).String()
	reveresePatternRev, err := obiapat.MakeApatPattern(reverse_rev, CLIAllowedMismatch(), false)

	if err != nil {
		log.Fatalf("Cannot compile reverse complement reverse primer %s : %v", CLIReversePrimer(), err)
	}

	aseq, err := obiapat.MakeApatSequence(seq, false)

	if err != nil {
		log.Fatalf("Cannot build apat sequence: %v", err)
	}

	fstart, fend, fnerr, hasfw := forwardPatternDir.BestMatch(aseq, 0, aseq.Len())
	rstart, rend, rnerr, hasrev := reveresePatternRev.BestMatch(aseq, 0, aseq.Len())

	for hasfw && !hasrev {
		var rseq *obiseq.BioSequence
		rseq, err = BuildPCRProduct(
			matched,
			CLIGraphFilesDirectory(),
			CLIKmerSize(),
			CLIForwardPrimer(),
			CLIReversePrimer(),
			true,
			CLISaveGraphToFiles(),
			CLIGraphFilesDirectory())

		if err != nil {
			log.Fatalf("Cannot build Reverse PCR sequence: %v", err)
		}

		kmerSize, _ := seq.GetIntAttribute("obiconsensus_kmer_size")
		fp, _ := seq.Subsequence(seq.Len()-kmerSize, seq.Len(), false)
		rp, _ := rseq.Subsequence(0, kmerSize, false)
		rp = rp.ReverseComplement(true)

		pairs, err := CLIPairedSequence()

		if err != nil {
			log.Errorf("Cannot open file (%v)", err)
			os.Exit(1)
		}

		nmatched := ExtractOnPatterns(pairs,
			fp.String(),
			rp.String(),
			CLIAllowedMismatch(),
			true,
		)

		in := map[string]bool{}

		for _, s := range matched {
			in[s.String()] = true
		}

		for _, s := range nmatched {
			if !in[s.String()] {
				matched = append(matched, s)
			}
		}

		seq, err = BuildPCRProduct(
			matched,
			CLIGraphFilesDirectory(),
			CLIKmerSize(),
			CLIForwardPrimer(),
			CLIReversePrimer(),
			false,
			CLISaveGraphToFiles(),
			CLIGraphFilesDirectory())

		aseq, err := obiapat.MakeApatSequence(seq, false)

		if err != nil {
			log.Fatalf("Cannot build apat sequence: %v", err)
		}
		fstart, fend, fnerr, hasfw = forwardPatternDir.BestMatch(aseq, 0, aseq.Len())
		rstart, rend, rnerr, hasrev = reveresePatternRev.BestMatch(aseq, 0, aseq.Len())

	}

	marker, _ := seq.Subsequence(fstart, rend, false)

	marker.SetAttribute("forward_primer", CLIForwardPrimer())
	match, _ := seq.Subsequence(fstart, fend, false)
	marker.SetAttribute("forward_match", match.String())
	marker.SetAttribute("forward_error", fnerr)

	marker.SetAttribute("reverse_primer", CLIReversePrimer())
	match, _ = seq.Subsequence(rstart, rend, false)
	marker.SetAttribute("reverse_match", match.ReverseComplement(true).String())
	marker.SetAttribute("reverse_error", rnerr)

	return marker
}

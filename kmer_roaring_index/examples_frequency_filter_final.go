package main

import (
	"fmt"
	"obikmer"
)

func main() {
	// ==========================================
	// EXEMPLE 1 : Utilisation basique
	// ==========================================
	fmt.Println("=== EXEMPLE 1 : Utilisation basique ===\n")

	k := 31
	minFreq := 3 // Garder les k-mers vus ≥3 fois

	// Créer le filtre
	filter := obikmer.NewFrequencyFilter(k, minFreq)

	// Simuler des séquences avec différentes fréquences
	sequences := [][]byte{
		[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"), // Kmer X
		[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"), // Kmer X (freq=2)
		[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"), // Kmer X (freq=3) ✓
		[]byte("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT"), // Kmer Y
		[]byte("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT"), // Kmer Y (freq=2) ✗
		[]byte("GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG"), // Kmer Z (freq=1) ✗
	}

	fmt.Printf("Traitement de %d séquences...\n", len(sequences))
	for _, seq := range sequences {
		filter.AddSequence(seq)
	}

	// Récupérer les k-mers filtrés
	filtered := filter.GetFilteredSet("filtered")
	fmt.Printf("\nK-mers avec freq ≥ %d: %d\n", minFreq, filtered.Cardinality())

	// Statistiques
	stats := filter.Stats()
	fmt.Println("\n" + stats.String())

	// ==========================================
	// EXEMPLE 2 : Vérifier les niveaux
	// ==========================================
	fmt.Println("\n=== EXEMPLE 2 : Inspection des niveaux ===\n")

	// Vérifier chaque niveau
	for level := 0; level < minFreq; level++ {
		levelSet := filter.GetKmersAtLevel(level)
		fmt.Printf("Niveau %d (freq≥%d): %d k-mers\n",
			level+1, level+1, levelSet.Cardinality())
	}

	// ==========================================
	// EXEMPLE 3 : Données réalistes
	// ==========================================
	fmt.Println("\n=== EXEMPLE 3 : Simulation données séquençage ===\n")

	filter2 := obikmer.NewFrequencyFilter(31, 3)

	// Simuler un dataset réaliste :
	// - 1000 reads
	// - 80% contiennent des erreurs (singletons)
	// - 15% vrais k-mers à basse fréquence
	// - 5% vrais k-mers à haute fréquence

	// Vraie séquence répétée
	trueSeq := []byte("ACGTACGTACGTACGTACGTACGTACGTACG")
	for i := 0; i < 50; i++ {
		filter2.AddSequence(trueSeq)
	}

	// Séquence à fréquence moyenne
	mediumSeq := []byte("CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC")
	for i := 0; i < 5; i++ {
		filter2.AddSequence(mediumSeq)
	}

	// Erreurs de séquençage (singletons)
	for i := 0; i < 100; i++ {
		errorSeq := []byte(fmt.Sprintf("TTTTTTTTTTTTTTTTTTTTTTTTTTTT%03d", i))
		filter2.AddSequence(errorSeq)
	}

	stats2 := filter2.Stats()
	fmt.Println(stats2.String())

	fmt.Println("Distribution attendue:")
	fmt.Println("  - Beaucoup de singletons (erreurs)")
	fmt.Println("  - Peu de k-mers à haute fréquence (signal)")
	fmt.Println("  → Filtrage efficace !")

	// ==========================================
	// EXEMPLE 4 : Tester différents seuils
	// ==========================================
	fmt.Println("\n=== EXEMPLE 4 : Comparaison de seuils ===\n")

	testSeqs := [][]byte{
		[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
		[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
		[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
		[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
		[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"), // freq=5
		[]byte("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT"),
		[]byte("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT"),
		[]byte("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT"), // freq=3
		[]byte("GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG"), // freq=1
	}

	for _, minFreq := range []int{2, 3, 5} {
		f := obikmer.NewFrequencyFilter(31, minFreq)
		f.AddSequences(testSeqs)

		fmt.Printf("minFreq=%d: %d k-mers retenus (%.2f MB)\n",
			minFreq,
			f.Cardinality(),
			float64(f.MemoryUsage())/1024/1024)
	}

	// ==========================================
	// EXEMPLE 5 : Comparaison mémoire
	// ==========================================
	fmt.Println("\n=== EXEMPLE 5 : Comparaison mémoire ===\n")

	filter3 := obikmer.NewFrequencyFilter(31, 3)

	// Simuler 10000 séquences
	for i := 0; i < 10000; i++ {
		seq := make([]byte, 100)
		for j := range seq {
			seq[j] = "ACGT"[(i+j)%4]
		}
		filter3.AddSequence(seq)
	}

	fmt.Println(filter3.CompareWithSimpleMap())

	// ==========================================
	// EXEMPLE 6 : Workflow complet
	// ==========================================
	fmt.Println("\n=== EXEMPLE 6 : Workflow complet ===\n")

	fmt.Println("1. Créer le filtre")
	finalFilter := obikmer.NewFrequencyFilter(31, 3)

	fmt.Println("2. Traiter les données (simulation)")
	// En pratique : lire depuis FASTQ
	// for read := range ReadFastq("data.fastq") {
	//     finalFilter.AddSequence(read)
	// }

	// Simulation
	for i := 0; i < 1000; i++ {
		seq := []byte("ACGTACGTACGTACGTACGTACGTACGTACG")
		finalFilter.AddSequence(seq)
	}

	fmt.Println("3. Récupérer les k-mers filtrés")
	result := finalFilter.GetFilteredSet("final")

	fmt.Println("4. Utiliser le résultat")
	fmt.Printf("   K-mers de qualité: %d\n", result.Cardinality())
	fmt.Printf("   Mémoire utilisée: %.2f MB\n", float64(finalFilter.MemoryUsage())/1024/1024)

	fmt.Println("5. Sauvegarder (optionnel)")
	// result.Save("filtered_kmers.bin")

	// ==========================================
	// EXEMPLE 7 : Vérification individuelle
	// ==========================================
	fmt.Println("\n=== EXEMPLE 7 : Vérification de k-mers spécifiques ===\n")

	checkFilter := obikmer.NewFrequencyFilter(31, 3)

	testSeq := []byte("ACGTACGTACGTACGTACGTACGTACGTACG")
	for i := 0; i < 5; i++ {
		checkFilter.AddSequence(testSeq)
	}

	var kmers []uint64
	kmers = obikmer.EncodeKmers(testSeq, 31, &kmers)

	if len(kmers) > 0 {
		testKmer := kmers[0]

		fmt.Printf("K-mer test: 0x%016X\n", testKmer)
		fmt.Printf("  Présent dans filtre: %v\n", checkFilter.Contains(testKmer))
		fmt.Printf("  Fréquence approx: %d\n", checkFilter.GetFrequency(testKmer))
	}

	// ==========================================
	// EXEMPLE 8 : Intégration avec collection
	// ==========================================
	fmt.Println("\n=== EXEMPLE 8 : Intégration avec KmerSetCollection ===\n")

	// Créer une collection de génomes filtrés
	collection := obikmer.NewKmerSetCollection(31)

	genomes := map[string][][]byte{
		"Genome1": {
			[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
			[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
			[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
			[]byte("TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT"), // Erreur
		},
		"Genome2": {
			[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
			[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
			[]byte("ACGTACGTACGTACGTACGTACGTACGTACG"),
			[]byte("GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG"), // Erreur
		},
	}

	for id, sequences := range genomes {
		// Filtrer chaque génome
		genomeFilter := obikmer.NewFrequencyFilter(31, 3)
		genomeFilter.AddSequences(sequences)

		// Ajouter à la collection
		filteredSet := genomeFilter.GetFilteredSet(id)
		collection.Add(filteredSet)

		fmt.Printf("%s: %d k-mers de qualité\n", id, filteredSet.Cardinality())
	}

	// Analyser la collection
	fmt.Println("\nAnalyse comparative:")
	collectionStats := collection.ComputeStats()
	fmt.Printf("  Core genome: %d k-mers\n", collectionStats.CoreSize)
	fmt.Printf("  Pan genome: %d k-mers\n", collectionStats.PanGenomeSize)

	// ==========================================
	// RÉSUMÉ
	// ==========================================
	fmt.Println("\n=== RÉSUMÉ ===\n")
	fmt.Println("Le FrequencyFilter permet de:")
	fmt.Println("  ✓ Filtrer les k-mers par fréquence minimale")
	fmt.Println("  ✓ Utiliser une mémoire optimale avec Roaring bitmaps")
	fmt.Println("  ✓ Une seule passe sur les données")
	fmt.Println("  ✓ Éliminer efficacement les erreurs de séquençage")
	fmt.Println("")
	fmt.Println("Workflow typique:")
	fmt.Println("  1. filter := NewFrequencyFilter(k, minFreq)")
	fmt.Println("  2. for each sequence: filter.AddSequence(seq)")
	fmt.Println("  3. filtered := filter.GetFilteredSet(id)")
	fmt.Println("  4. Utiliser filtered dans vos analyses")
}

// ==================================
// FONCTION HELPER POUR BENCHMARKS
// ==================================

func BenchmarkFrequencyFilter() {
	k := 31
	minFreq := 3

	// Test avec différentes tailles
	sizes := []int{1000, 10000, 100000}

	fmt.Println("\n=== BENCHMARK ===\n")

	for _, size := range sizes {
		filter := obikmer.NewFrequencyFilter(k, minFreq)

		// Générer des séquences
		for i := 0; i < size; i++ {
			seq := make([]byte, 100)
			for j := range seq {
				seq[j] = "ACGT"[(i+j)%4]
			}
			filter.AddSequence(seq)
		}

		fmt.Printf("Size=%d reads:\n", size)
		fmt.Printf("  Filtered k-mers: %d\n", filter.Cardinality())
		fmt.Printf("  Memory: %.2f MB\n", float64(filter.MemoryUsage())/1024/1024)
		fmt.Println()
	}
}

// ==================================
// FONCTION POUR DONNÉES RÉELLES
// ==================================

func ProcessRealData() {
	// Exemple pour traiter de vraies données FASTQ

	k := 31
	minFreq := 3

	filter := obikmer.NewFrequencyFilter(k, minFreq)

	// Pseudo-code pour lire un FASTQ
	/*
	fastqFile := "sample.fastq"
	reader := NewFastqReader(fastqFile)

	for reader.HasNext() {
		read := reader.Next()
		filter.AddSequence(read.Sequence)
	}

	// Récupérer le résultat
	filtered := filter.GetFilteredSet("sample_filtered")
	filtered.Save("sample_filtered_kmers.bin")

	// Stats
	stats := filter.Stats()
	fmt.Println(stats.String())
	*/

	fmt.Println("Workflow pour données réelles:")
	fmt.Println("  1. Créer le filtre avec minFreq approprié (2-5 typique)")
	fmt.Println("  2. Stream les reads depuis FASTQ")
	fmt.Println("  3. Récupérer les k-mers filtrés")
	fmt.Println("  4. Utiliser pour assemblage/comparaison/etc.")

	_ = filter // unused
}

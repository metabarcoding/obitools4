package obiseq

import "iter"

func (seq *BioSequence) Kmers(k int) iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		// Gérer les cas où k est invalide ou la séquence trop courte
		if k <= 0 || k > len(seq.sequence) {
			return
		}

		// Itérer sur tous les k-mers possibles
		for i := 0; i <= len(seq.sequence)-k; i++ {
			// Extraire le k-mer actuel
			kmer := seq.sequence[i : i+k]

			// Passer au consommateur et arrêter si demandé
			if !yield(kmer) {
				return
			}
		}
	}
}

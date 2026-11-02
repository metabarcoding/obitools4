//go:build !goexperiment.simd

package obikmer

func kmerLCSBatch(query uint64, partners []uint64, k int, out []KmerLCSResult) {
	kmerLCSBatchScalar(query, partners, k, out)
}

//go:build !goexperiment.simd

package obikmer

func kmerLCSBatch(query uint64, partners []uint64, k int, out []KmerLCSResult) {
	kmerLCSBatchScalar(query, partners, k, out)
}

// kmerLCSPrepared has nothing to precompute without SIMD.
type kmerLCSPrepared struct {
	k        int
	partners []uint64
}

func prepareKmerLCS(partners []uint64, k int) kmerLCSPrepared {
	return kmerLCSPrepared{k: k, partners: partners}
}

func (p *kmerLCSPrepared) align(query uint64, out []KmerLCSResult) {
	kmerLCSBatchScalar(query, p.partners, p.k, out)
}

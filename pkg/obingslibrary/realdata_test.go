package obingslibrary_test

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// BenchmarkExtractMultiBarcode runs ExtractMultiBarcode over the reads of the FASTQ
// file OBIDNAPAT_FASTQ (OBIDNAPAT_READS of them, 20000 by default) with the sample
// description OBIDNAPAT_NGSFILTER; skipped without them. Meant for -cpuprofile.
func BenchmarkExtractMultiBarcode(b *testing.B) {
	path, ngs := os.Getenv("OBIDNAPAT_FASTQ"), os.Getenv("OBIDNAPAT_NGSFILTER")
	if path == "" || ngs == "" {
		b.Skip("OBIDNAPAT_FASTQ and OBIDNAPAT_NGSFILTER not set")
	}
	n := 20000
	if v, err := strconv.Atoi(os.Getenv("OBIDNAPAT_READS")); err == nil && v > 0 {
		n = v
	}
	f, err := os.Open(ngs)
	if err != nil {
		b.Fatal(err)
	}
	library, err := obiformats.ReadNGSFilter(f)
	f.Close()
	if err != nil {
		b.Fatal(err)
	}
	if err := library.Compile2(); err != nil {
		b.Fatal(err)
	}

	r, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer r.Close()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	var reads []*obiseq.BioSequence
	for line, id := 0, ""; sc.Scan() && len(reads) < n; line++ {
		switch line % 4 {
		case 0:
			id = strings.TrimPrefix(strings.Fields(sc.Text())[0], "@")
		case 1:
			reads = append(reads, obiseq.NewBioSequence(id, append([]byte(nil), sc.Bytes()...), ""))
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range reads {
			library.ExtractMultiBarcode(s.Copy())
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(len(reads)), "ns/read")
}

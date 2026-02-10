package obik

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"github.com/DavidGamba/go-getoptions"
	"gopkg.in/yaml.v3"
)

type setSummary struct {
	Index    int                    `json:"index" yaml:"index"`
	ID       string                 `json:"id" yaml:"id"`
	Count    uint64                 `json:"count" yaml:"count"`
	DiskSize int64                  `json:"disk_bytes" yaml:"disk_bytes"`
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type groupSummary struct {
	Path       string                 `json:"path" yaml:"path"`
	ID         string                 `json:"id,omitempty" yaml:"id,omitempty"`
	K          int                    `json:"k" yaml:"k"`
	M          int                    `json:"m" yaml:"m"`
	Partitions int                    `json:"partitions" yaml:"partitions"`
	TotalSets  int                    `json:"total_sets" yaml:"total_sets"`
	TotalKmers uint64                 `json:"total_kmers" yaml:"total_kmers"`
	TotalDisk  int64                  `json:"total_disk_bytes" yaml:"total_disk_bytes"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Sets       []setSummary           `json:"sets" yaml:"sets"`
	Jaccard    [][]float64            `json:"jaccard,omitempty" yaml:"jaccard,omitempty"`
}

func runSummary(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: obik summary [options] <index_directory>")
	}

	ksg, err := obikmer.OpenKmerSetGroup(args[0])
	if err != nil {
		return fmt.Errorf("failed to open kmer index: %w", err)
	}

	summary := groupSummary{
		Path:       ksg.Path(),
		ID:         ksg.Id(),
		K:          ksg.K(),
		M:          ksg.M(),
		Partitions: ksg.Partitions(),
		TotalSets:  ksg.Size(),
		TotalKmers: ksg.Len(),
		Metadata:   ksg.Metadata,
		Sets:       make([]setSummary, ksg.Size()),
	}

	var totalDisk int64
	for i := 0; i < ksg.Size(); i++ {
		diskSize := computeSetDiskSize(ksg, i)
		totalDisk += diskSize
		summary.Sets[i] = setSummary{
			Index:    i,
			ID:       ksg.SetIDOf(i),
			Count:    ksg.Len(i),
			DiskSize: diskSize,
			Metadata: ksg.AllSetMetadata(i),
		}
	}
	summary.TotalDisk = totalDisk

	// Jaccard matrix
	if _jaccard && ksg.Size() > 1 {
		dm := ksg.JaccardDistanceMatrix()
		n := ksg.Size()
		matrix := make([][]float64, n)
		for i := 0; i < n; i++ {
			matrix[i] = make([]float64, n)
			for j := 0; j < n; j++ {
				if i == j {
					matrix[i][j] = 0
				} else {
					matrix[i][j] = dm.Get(i, j)
				}
			}
		}
		summary.Jaccard = matrix
	}

	format := CLIOutFormat()
	switch format {
	case "json":
		return outputSummaryJSON(summary)
	case "yaml":
		return outputSummaryYAML(summary)
	case "csv":
		return outputSummaryCSV(summary)
	default:
		return outputSummaryJSON(summary)
	}
}

func computeSetDiskSize(ksg *obikmer.KmerSetGroup, setIndex int) int64 {
	var total int64
	for p := 0; p < ksg.Partitions(); p++ {
		path := ksg.PartitionPath(setIndex, p)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		total += info.Size()
	}
	// Also count the set directory entry itself
	setDir := filepath.Join(ksg.Path(), fmt.Sprintf("set_%d", setIndex))
	entries, err := os.ReadDir(setDir)
	if err == nil {
		// We already counted .kdi files above; this is just for completeness
		_ = entries
	}
	return total
}

func outputSummaryJSON(summary groupSummary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputSummaryYAML(summary groupSummary) error {
	data, err := yaml.Marshal(summary)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func outputSummaryCSV(summary groupSummary) error {
	fmt.Println("index,id,count,disk_bytes")
	for _, s := range summary.Sets {
		fmt.Printf("%d,%s,%d,%d\n", s.Index, s.ID, s.Count, s.DiskSize)
	}
	return nil
}

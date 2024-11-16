package obifind

import (
	"bytes"
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

func IFilterRankRestriction() func(*obitax.ITaxon) *obitax.ITaxon {
	f := func(s *obitax.ITaxon) *obitax.ITaxon {
		return s
	}

	if __restrict_rank__ != "" {
		f = func(s *obitax.ITaxon) *obitax.ITaxon {
			return s.IFilterOnTaxRank(__restrict_rank__)
		}
	}

	return f
}

func ITaxonNameMatcher() (func(string) *obitax.ITaxon, error) {
	taxonomy, err := CLILoadSelectedTaxonomy()

	if err != nil {
		return nil, err
	}

	fun := func(name string) *obitax.ITaxon {
		return taxonomy.IFilterOnName(name, __fixed_pattern__)
	}

	return fun, nil
}

func ITaxonRestrictions() (func(*obitax.ITaxon) *obitax.ITaxon, error) {

	clades, err := CLITaxonomicalRestrictions()

	if err != nil {
		return nil, err
	}

	rankfilter := IFilterRankRestriction()

	fun := func(iterator *obitax.ITaxon) *obitax.ITaxon {
		return rankfilter(iterator).IFilterBelongingSubclades(clades)
	}

	return fun, nil
}

func TaxonAsString(taxon *obitax.Taxon, pattern string) string {
	text := taxon.ScientificName()

	if __with_path__ {
		var bf bytes.Buffer
		path := taxon.Path()

		bf.WriteString(path.Get(path.Len() - 1).ScientificName())

		for i := path.Len() - 2; i >= 0; i-- {
			fmt.Fprintf(&bf, ":%s", path.Get(i).ScientificName())
		}

		text = bf.String()
	}

	return fmt.Sprintf("%-20s | %10s | %10s | %-20s | %s",
		pattern,
		taxon.String(),
		taxon.Parent().String(),
		taxon.Rank(),
		text)
}

func TaxonWriter(itaxa *obitax.ITaxon, pattern string) {

	for itaxa.Next() {
		fmt.Println(TaxonAsString(itaxa.Get(), pattern))
	}
}

package obifind

import (
	"bytes"
	"fmt"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitax"
)

func IFilterRankRestriction() func(*obitax.ITaxonSet) *obitax.ITaxonSet {
	f := func(s *obitax.ITaxonSet) *obitax.ITaxonSet {
		return s
	}

	if __restrict_rank__ != "" {
		f = func(s *obitax.ITaxonSet) *obitax.ITaxonSet {
			return s.IFilterOnTaxRank(__restrict_rank__)
		}
	}

	return f
}

func ITaxonNameMatcher() (func(string) *obitax.ITaxonSet, error) {
	taxonomy, err := CLILoadSelectedTaxonomy()

	if err != nil {
		return nil, err
	}

	fun := func(name string) *obitax.ITaxonSet {
		return taxonomy.IFilterOnName(name, __fixed_pattern__)
	}

	return fun, nil
}

func ITaxonRestrictions() (func(*obitax.ITaxonSet) *obitax.ITaxonSet, error) {

	clades, err := CLITaxonomicalRestrictions()

	if err != nil {
		return nil, err
	}

	rankfilter := IFilterRankRestriction()

	fun := func(iterator *obitax.ITaxonSet) *obitax.ITaxonSet {
		return rankfilter(iterator).IFilterBelongingSubclades(clades)
	}

	return fun, nil
}

func TaxonAsString(taxon *obitax.TaxNode, pattern string) string {
	text := taxon.ScientificName()

	if __with_path__ {
		var bf bytes.Buffer
		path, err := taxon.Path()

		if err != nil {
			fmt.Printf("%+v", err)
		}

		bf.WriteString(path.Get(path.Length() - 1).ScientificName())

		for i := path.Length() - 2; i >= 0; i-- {
			fmt.Fprintf(&bf, ":%s", path.Get(i).ScientificName())
		}

		text = bf.String()
	}

	return fmt.Sprintf("%-20s | %10d | %10d | %-20s | %s",
		pattern,
		taxon.Taxid(),
		taxon.Parent().Taxid(),
		taxon.Rank(),
		text)
}

func TaxonWriter(itaxa *obitax.ITaxonSet, pattern string) {
	for itaxa.Next() {
		fmt.Println(TaxonAsString(itaxa.Get(), pattern))
	}
}

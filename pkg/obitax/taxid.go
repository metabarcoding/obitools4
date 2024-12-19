package obitax

import (
	"fmt"
	"strconv"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

// Taxid represents a taxonomic identifier as a pointer to a string.
type Taxid *string

// TaxidFactory is a factory for creating Taxid instances from strings and integers.
type TaxidFactory struct {
	inner    *InnerString
	code     string
	alphabet obiutils.AsciiSet
}

// NewTaxidFactory creates and returns a new instance of TaxidFactory.
func NewTaxidFactory(code string, alphabet obiutils.AsciiSet) *TaxidFactory {
	return &TaxidFactory{
		inner:    NewInnerString(),
		code:     code,
		alphabet: alphabet,
	}
	// Initialize and return a new TaxidFactory.
}

// FromString converts a string representation of a taxonomic identifier into a Taxid.
// It extracts the relevant part of the string after the first colon (':') if present.
func (f *TaxidFactory) FromString(taxid string) (Taxid, error) {
	taxid = obiutils.AsciiSpaceSet.TrimLeft(taxid)
	part1, part2 := obiutils.SplitInTwo(taxid, ':')
	if len(part2) == 0 {
		taxid = part1
	} else {
		//log.Warnf("TaxidFactory.FromString: taxid %s -> -%s- -%s- ", taxid, part1, part2)
		if part1 != f.code {
			return nil, fmt.Errorf("taxid %s string does not start with taxonomy code %s", taxid, f.code)
		}
		taxid = part2
	}

	taxid, err := f.alphabet.FirstWord(taxid) // Get the first word from the input string.

	if err != nil {
		return nil, err
	}

	// Return a new Taxid by innerizing the extracted taxid string.
	rep := Taxid(f.inner.Innerize(taxid))
	return rep, nil
}

// FromInt converts an integer taxonomic identifier into a Taxid.
// It first converts the integer to a string and then innerizes it.
func (f *TaxidFactory) FromInt(taxid int) (Taxid, error) {
	s := strconv.Itoa(taxid)        // Convert the integer to a string.
	return f.inner.Innerize(s), nil // Return a new Taxid by innerizing the string.
}

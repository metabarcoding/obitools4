package obitax

import "log"

var __defaut_taxonomy__ *Taxonomy

func (taxonomy *Taxonomy) SetAsDefault() {
	__defaut_taxonomy__ = taxonomy
}

func (taxonomy *Taxonomy) OrDefault(panicOnNil bool) *Taxonomy {
	if taxonomy == nil {
		return __defaut_taxonomy__
	}

	if panicOnNil && taxonomy == nil {
		log.Panic("Cannot deal with nil taxonomy")
	}

	return taxonomy
}

func IsDefaultTaxonomyDefined() bool {
	return __defaut_taxonomy__ != nil
}

func DefaultTaxonomy() *Taxonomy {
	return __defaut_taxonomy__
}

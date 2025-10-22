package obitax

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

var __defaut_taxonomy__ *Taxonomy
var __defaut_taxonomy_mutex__ sync.Mutex

func (taxonomy *Taxonomy) SetAsDefault() {
	log.Infof("Set as default taxonomy %s", taxonomy.Name())
	__defaut_taxonomy__ = taxonomy
}

func (taxonomy *Taxonomy) OrDefault(panicOnNil bool) *Taxonomy {
	if taxonomy == nil {
		taxonomy = __defaut_taxonomy__
	}

	if panicOnNil && taxonomy == nil {
		log.Panic("Cannot deal with nil taxonomy")
	}

	return taxonomy
}

func HasDefaultTaxonomyDefined() bool {
	return __defaut_taxonomy__ != nil
}

func DefaultTaxonomy() *Taxonomy {
	return __defaut_taxonomy__
}

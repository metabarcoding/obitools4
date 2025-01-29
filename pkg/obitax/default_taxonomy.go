package obitax

import (
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
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

func IsDefaultTaxonomyDefined() bool {
	return __defaut_taxonomy__ != nil
}

func DefaultTaxonomy() *Taxonomy {
	var err error
	if __defaut_taxonomy__ == nil {
		if obidefault.HasSelectedTaxonomy() {
			__defaut_taxonomy_mutex__.Lock()
			defer __defaut_taxonomy_mutex__.Unlock()
			if __defaut_taxonomy__ == nil {
				__defaut_taxonomy__, err = LoadTaxonomy(
					obidefault.SelectedTaxonomy(),
					!obidefault.AreAlternativeNamesSelected(),
				)

				if err != nil {
					log.Fatalf("Cannot load default taxonomy: %v", err)

				}
			}
		}
	}

	return __defaut_taxonomy__
}

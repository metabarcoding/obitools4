package obitax

import (
	log "github.com/sirupsen/logrus"
)

func (t1 *Taxon) LCA(t2 *Taxon) (*Taxon, error) {
	if t1 == nil {
		log.Panicf("Try to get LCA of nil taxon")
	}

	if t2 == nil {
		log.Panicf("Try to get LCA of nil taxon")
	}

	p1 := t1.Path()
	p2 := t2.Path()

	i1 := p1.Len() - 1
	i2 := p2.Len() - 1

	for i1 >= 0 && i2 >= 0 && p1.slice[i1].id == p2.slice[i2].id {
		i1--
		i2--
	}

	return &Taxon{
		Taxonomy: t1.Taxonomy,
		Node:     p1.slice[i1+1],
	}, nil
}

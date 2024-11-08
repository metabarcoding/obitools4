package obitax

import (
	log "github.com/sirupsen/logrus"
)

func (t1 *TaxNode) LCA(t2 *TaxNode) (*TaxNode, error) {
	if t1 == nil {
		log.Panicf("Try to get LCA of nil taxon")
	}

	if t2 == nil {
		log.Panicf("Try to get LCA of nil taxon")
	}

	p1, err1 := t1.Path()

	if err1 != nil {
		return nil, err1
	}

	p2, err2 := t2.Path()

	if err2 != nil {
		return nil, err2
	}

	i1 := len(*p1) - 1
	i2 := len(*p2) - 1

	for i1 >= 0 && i2 >= 0 && (*p1)[i1].taxid == (*p2)[i2].taxid {
		i1--
		i2--
	}

	return (*p1)[i1+1], nil
}

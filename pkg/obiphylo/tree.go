package obiphylo

import (
	"fmt"
	"math"
	"strings"
)

type PhyloNode struct {
	Name       string
	Children   map[*PhyloNode]float64
	Attributes map[string]any
}

func NewPhyloNode() *PhyloNode {
	return &PhyloNode{}
}

func (n *PhyloNode) AddChild(child *PhyloNode, distance float64) {
	if n.Children == nil {
		n.Children = map[*PhyloNode]float64{}
	}
	n.Children[child] = distance
}

func (n *PhyloNode) SetAttribute(key string, value any) {
	if n.Attributes == nil {
		n.Attributes = make(map[string]any)
	}
	n.Attributes[key] = value
}

func (n *PhyloNode) GetDistanceToChild(child *PhyloNode) float64 {
	return n.Children[child]
}

func (n *PhyloNode) GetAttribute(key string) any {
	return n.Attributes[key]
}

func (n *PhyloNode) Newick(level int) string {
	nc := len(n.Children)
	result := strings.Builder{}
	result.WriteString(strings.Repeat(" ", level))
	if nc > 0 {
		result.WriteString("(\n")
		i := 0
		for child, distance := range n.Children {
			result.WriteString(child.Newick(level + 1))
			if !math.IsNaN(distance) {
				result.WriteString(fmt.Sprintf(":%.5f", distance))
			}
			i++
			if i < nc {
				result.WriteByte(',')
			}
			result.WriteString("\n")
		}
		result.WriteString(strings.Repeat("  ", level))
		result.WriteByte(')')
	}
	if n.Name != "" {
		result.WriteString(n.Name)
	}

	if level == 0 {
		result.WriteString(";\n")
	}

	return result.String()
}

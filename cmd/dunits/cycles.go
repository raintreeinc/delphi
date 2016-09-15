package main

import (
	"strings"

	"github.com/gonum/graph/simple"
	"github.com/gonum/graph/topo"
)

func FindCycles(index *Index) [][]string {
	refid := map[string]simple.Node{}
	refunit := map[simple.Node]string{}
	graph := simple.NewDirectedGraph(1.0, 0.0)

	for _, use := range index.Uses {
		id := graph.NewNodeID()
		node := simple.Node(id)
		refid[strings.ToLower(use.Unit)] = node
		refunit[node] = use.Unit

		graph.AddNode(node)
	}

	for _, use := range index.Uses {
		cuse := strings.ToLower(use.Unit)
		for _, dep := range use.Interface {
			cdep := strings.ToLower(dep)

			from, to := refid[cuse], refid[cdep]
			graph.SetEdge(simple.Edge{from, to, 1.0})
		}
	}

	var cycles [][]string

	for _, component := range topo.TarjanSCC(graph) {
		if len(component) <= 1 {
			continue
		}
		var cycle []string
		for _, node := range component {
			cycle = append(cycle, refunit[node.(simple.Node)])
		}

		cycles = append(cycles, cycle)
	}

	return cycles
}

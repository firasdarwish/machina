package machina

import (
	"fmt"
	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
	"io"
)

func (m *machine[TState, TTrigger]) GenerateDotGraph(w io.Writer) error {
	g := graph.New(func(t TState) string {
		return fmt.Sprintf("%v", t)
	}, graph.Directed())

	states := []TState{}

	for t, c := range m.statesConfig {
		if InList(states, t) {
			continue
		}
		states = append(states, t)
		err := g.AddVertex(t)
		if err != nil {
			panic(err)
		}

		for _, t3 := range c.transitions {
			for _, t4 := range t3 {
				if InList(states, t4.toState) {
					continue
				}
				states = append(states, t4.toState)
				err = g.AddVertex(t4.toState)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	for t, s := range m.statesConfig {
		for trig, i := range s.transitions {
			for _, t3 := range i {
				err := g.AddEdge(fmt.Sprintf("%v", t), fmt.Sprintf("%v", t3.toState), func(properties *graph.EdgeProperties) {
					properties.Attributes["label"] = fmt.Sprintf("%v", trig)
				})
				if err != nil {
					panic(err)
				}
			}
		}
	}

	return draw.DOT(g, w)
}

func InList[T comparable](list []T, el T) bool {
	for _, t := range list {
		if t == el {
			return true
		}
	}

	return false
}

// Use graphviz to show a rete network.

package rete

import "fmt"
import "os"
import "github.com/awalterschulze/gographviz"

func MakeGraph(root Node) (*gographviz.Escape, error) {
	graphName := "rete"
	graph := gographviz.NewEscape()
	graph.SetDir(true)
	graph.SetName(graphName)
	Walk(root, func(n Node) {
		nodeattrs := map[string]string{}
		if _, ok := n.(AbstractBufferNode); ok {
			nodeattrs["shape"] = "box"
		}
		graph.AddNode(graphName, n.Label(), nodeattrs)
	})
	Walk(root, func(n Node) {
		for _, o := range n.Outputs() {
			graph.AddEdge(n.Label(), o.Label(), true, nil)
		}
	})
	return graph, nil
}

func WriteGraphvizFile(graph *gographviz.Escape, filename string) error {
	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Can't open %s: %s", filename, err)
	}
	defer out.Close()
	out.WriteString(graph.String())
	return nil
}

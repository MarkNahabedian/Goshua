// Generate GraphViz files to debug the case where one expected a
// conclusion that didn;t occur.

package rete

import "fmt"
import "os"
import "path/filepath"
import "reflect"
import "github.com/awalterschulze/gographviz"


type MissingConclusionGrapher struct {
	// Inputs
	OutputPath string
	Root Node
	Expected reflect.Type
	// Intermediate data:
	
	// Outputs
	Err error
	// FromType is the type that the caller expected the rete to
	// conclude.
	FromType reflect.Type
	// BufferCount is the total number of TypedBufferNodes in the rete.
	BufferCount int
	// InstalledRulesCount is the total number of RuleNodes
	// installed in the rete.
	InstalledRulesCount int
	// TypeCount is the number of type nodes added to the graph.
	TypeCount int
	// RuleCount returns the number of rule nodes added to the graph.
	RuleCount int

	Success bool
}


// ruleGraphId returns the GraphViz dot file id for the Rule.
func (g *MissingConclusionGrapher) ruleGraphId(rule Rule) string {
	return "rule " + rule.Name()
}

// typeGraphId returns the GraphViz dot file id for the type.
func (g *MissingConclusionGrapher) typeGraphId(t reflect.Type) string {
	// gographviz is failing to quote identifiers that contain a period
	// return t.String()
	return t.Name()
}


type graphvizAttrs map[string]string

var defaultBufferNodeAttributes = graphvizAttrs(map[string]string {
	"shape": "box",
})

var defaultRuleNodeAttributes = graphvizAttrs(map[string]string {
	"shape": "oval",
})

var defaultEdgeAttributes = graphvizAttrs(map[string]string {
})

func (a graphvizAttrs) copy() graphvizAttrs {
	copy := map[string]string{}
	for k, v := range a {
		copy[k] = v
	}
	return graphvizAttrs(copy)
}

func (a graphvizAttrs) set (attribute, value string) graphvizAttrs {
	a[attribute] = value
	return a
}


// GraphMissingConclusion writes a GraphViz dot file to output_path
// that graphs the rules and data that failed to conclude expected.
func GraphMissingConclusion(output_path string, root Node, expected reflect.Type) *MissingConclusionGrapher {
	g := &MissingConclusionGrapher {
		Root: root,
		Expected: expected,
	}
	installed_rules := []Rule{}
	buffers := map[reflect.Type] TypedBufferNode{}
	Walk(root, func(n Node) {
		switch n1 := n.(type) {
		case *RuleNode:
			installed_rules = append(installed_rules, n1.RuleSpec)
		case TypedBufferNode:
			t := n1.Type()
			if t != nil {
				buffers[t] = n1
			}
		}
	})
	g.BufferCount = len(buffers)
	g.InstalledRulesCount = len(installed_rules)
	graphName := fmt.Sprintf("missing-%s", expected.Name())
	graph := gographviz.NewEscape()
	graph.SetDir(true)
	graph.SetName(graphName)
	graphed_types := map[reflect.Type]bool{}
	graphed_rules := map[Rule]bool{}
	// Add a type node to the graph if not already present:
	graph_type := func (t reflect.Type) {
		if graphed_types[t] {
			return
		}
		bt := buffers[t]
		if bt == nil {
			panic(fmt.Sprintf("No buffer for %s", t))
		}
		graph.AddNode(graphName, g.typeGraphId(t),
			defaultBufferNodeAttributes.
				copy().set(
				"label",
				fmt.Sprintf("%s: %d items", t.String(), bt.Count())))
		g.TypeCount += 1
		graphed_types[t] = true
	}
	// Add a rule node to the graph if not already present: 
	graph_rule := func (r Rule) bool {
		if graphed_rules[r] {
			return false
		}
		graph.AddNode(graphName, g.ruleGraphId(r), defaultRuleNodeAttributes)	
		graphed_rules[r] = true
		for _, t := range r.ParamTypes() {
			graph.AddEdge(g.typeGraphId(t), g.ruleGraphId(r),
				true, defaultEdgeAttributes)
		}
		for _, t := range r.EmitTypes() {
			graph.AddEdge(g.ruleGraphId(r), g.typeGraphId(t),
				true, defaultEdgeAttributes)
		}
		return true
	}
	want_types := []reflect.Type { expected }
	for len(want_types) > 0 {
		typ := want_types[0]
		graph_type(typ)
		want_types = want_types[1:]
		for _, rule := range installed_rules {
			if rule.Emits(typ) {
				if graph_rule(rule) {
					want_types = append(want_types, rule.ParamTypes()...)
				}
			}
		}
	}
	g.OutputPath, g.Err = filepath.Abs(output_path)
	out, err := os.Create(g.OutputPath)
	if err != nil {
		g.Err = fmt.Errorf("Can't create %s: %s", g.OutputPath, err)
		return g
	}
	defer out.Close()
	out.WriteString(graph.String())
	g.Success = true
	return g
}


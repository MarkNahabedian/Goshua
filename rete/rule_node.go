package rete

import "fmt"
import "goshua/rete/rule_compiler/runtime"


// RuleNode implements the application of a rule.
type RuleNode struct {
	BasicNode
	RuleSpec runtime.Rule
}

// IsValid is part of the node interface.
func (n *RuleNode) IsValid() bool {
	for i, input := range n.Inputs() {
		if _, ok := input.(AbstractBufferNode); !ok {
			return false
		}
		if n.RuleSpec.ParamTypes()[i] != input.Inputs()[0].(*TypeTestNode).Type {
			return false
		}
	}
	return true
}

func (n *RuleNode) Label() string {
	return fmt.Sprintf("rule %s", n.RuleSpec.Name())
}

func InstallRule(root Node, rule runtime.Rule) {
	rule_node := &RuleNode {
		RuleSpec: rule,
	}
	for _, param_type := range rule.ParamTypes() {
		ttn := GetTypeTestNode(root, param_type)
		if ttn == nil {
			panic(fmt.Sprintf("GetTypeTestNode returned nil for %v", param_type))
		}
		rpn := GetRuleParameterNode(ttn)
		if rpn == nil {
			panic("GetRuleParameterNode returned nil for %v")
		}
		Connect(rpn, rule_node)
	}
	Connect(rule_node, root)
}


type RuleParameterNode struct {
	BufferNode
}

func (n *RuleParameterNode) IsValid() bool {
	if len(n.Inputs()) != 1 {
		return false
	}
	if _, ok := n.Inputs()[0].(*TypeTestNode); !ok {
		return false
	}
	return true
}

func (n *RuleParameterNode) Label() string {
	return fmt.Sprintf("rule input %s", n.Inputs()[0].(*TypeTestNode).Type.String())
}


func GetRuleParameterNode(ttn *TypeTestNode) *RuleParameterNode {
	for _, o := range ttn.Outputs() {
		if rpn, ok := o.(*RuleParameterNode); ok {
			return rpn
		}
	}
	rpn := &RuleParameterNode{}
	Connect(ttn, rpn)
	return rpn
}

func fill_and_call(in *RuleParameterNode, in_item interface{},
	rule_node *RuleNode,
	param_position int,
	parameters []interface{}) {
	// The same RuleParameterNode might appear multiple times in
	// the inputs of a RuleNode if the type represented by that
	// RuleParameterNode appears in multiple argument positions
	// in the rule function.
	//
	// We must use in_item for exactly one such parameter and
	// DoItems for the rest, otherwise we will skip some input
	// combinations.  Since it doesn't matter which of several
	// parameters of that type we pass in_item for, we use it for
	// the first parameter of that type.
	if param_position >= len(parameters) {
		rule_node.RuleSpec.Caller()(rule_node, parameters)
		return
	}
	nth_input := rule_node.Inputs()[param_position]
	if in != nil && nth_input == in {
		parameters[param_position] = in_item
		// Only restrict to in_item at the first opportunity:
		fill_and_call((*RuleParameterNode)(nil), nil, rule_node,
			param_position + 1, parameters)		
	} else {
		nth_input.(AbstractBufferNode).DoItems(
			func(item interface{}) {
				parameters[param_position] = item
				fill_and_call(in, item, rule_node,
					param_position + 1, parameters)
				})
	}
}

func (node *RuleParameterNode) Receive(item interface{}) {
	node.items = append(node.items, item)
	for _, output := range node.Outputs() {
		if rule_node, ok := output.(*RuleNode); ok {
			parameters := make([]interface{},
				len(rule_node.RuleSpec.ParamTypes()))
			fill_and_call(node, item, rule_node, 0, parameters)
		} else {
			output.Receive(item)
		}
	}
}


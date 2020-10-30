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

const DEBUG_FILL_AND_CALL bool = false

func indent(s string, count int) string {
	out := ""
	for i := 0; i < count; i++ {
		out += s
	}
	return out
}

func fill_and_call(in *RuleParameterNode, in_item interface{}, rule_node *RuleNode) {
	if DEBUG_FILL_AND_CALL {
		fmt.Printf("fill_and_call %v %v %v\n", in, in_item, rule_node)
	}
	// The same RuleParameterNode might appear multiple times in
	// the inputs of a RuleNode if the type represented by that
	// RuleParameterNode appears in multiple argument positions
	// in the rule function.
	//
	// Any combinations of input parameters to the rule that do not
	// include in_item have already been considered as previous
	// items have been Received.  We must now consider all
	// parameter combinations where in_item appears as at least
	// one parameter.
	parameters := make([]interface{}, len(rule_node.Inputs()))
	in_count := 0
	for _, node := range rule_node.Inputs() {
		if node == in {
			in_count += 1
		}
	}
	var f func(int, bool)
	/*
	// This one causes duplicate calls when the same type appears
	// as more than one parameter:
	f = func (param_position int, includes_in bool) {
		if DEBUG_FILL_AND_CALL {
			fmt.Printf("fill_and_call/f %d %s %#v\n",
				param_position, in.Label(), parameters)
		}
		if param_position >= len(parameters) {
			if includes_in {
				rule_node.RuleSpec.Caller()(rule_node, parameters)
			}
			return
		}
		nth_input := rule_node.Inputs()[param_position]
		// This parameter_position is the only
		// RuleParameterNode of this type so we only consider
		// in_item for this parameter_position.  If there are
		// more parameter_positions with the same
		// RuleParameterNode (and this parameter type) then we
		// use DoItems to consider all buffered values of that
		// type.
		if nth_input == in && in_count == 1 {
			parameters[param_position] = in_item
			f(param_position + 1, true)
		} else {
			nth_input.(AbstractBufferNode).DoItems(
				func(item interface{}) {
					parameters[param_position] = item
					f(param_position + 1,
						includes_in || item == in_item)
				})
		}
	}
	*/
	f = func (param_position int, includes_in bool) {
		if DEBUG_FILL_AND_CALL {
			fmt.Printf("%s fill_and_call/f %d %s %s\n",
				indent("  ", param_position),
				param_position, in.Label(), parameters)
		}
		if param_position >= len(parameters) {
			if includes_in {
				rule_node.RuleSpec.Caller()(rule_node, parameters)
			}
			return
		}
		nth_input := rule_node.Inputs()[param_position]
		nth_input.(AbstractBufferNode).DoItems(
			func(item interface{}) {
				parameters[param_position] = item
				f(param_position + 1,
					includes_in || item == in_item)
			})
	}
	f(0, false)
}

func (node *RuleParameterNode) Receive(item interface{}) {
	node.items = append(node.items, item)
	// A single rule might have more than one parameter of a given
	// type and thus might appear as more than one output of a
	// given RuleParameterNode.  This is a consequence of the
	// decision that a rule have one input per parameter, rather
	// than one input per parameter type.
	done := map[Node]bool{}
	for _, output := range node.Outputs() {
		if done[output] {
			return
		}
		done[output] = true
		if rule_node, ok := output.(*RuleNode); ok {
			fill_and_call(node, item, rule_node)
		} else {
			output.Receive(item)
		}
	}
}


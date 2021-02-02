package rete

import "fmt"
import "reflect"


// RuleNode implements the application of a rule.
type RuleNode struct {
	BasicNode
	RuleSpec Rule
}

// Validate is part of the node interface.
func (n *RuleNode) Validate() []error {
	errors := ValidateConnectivity(n)
	for i, input := range n.Inputs() {
		if _, ok := input.(AbstractBufferNode); !ok {
			errors = append(errors,
				fmt.Errorf("input %s of %s is not an AbstractBufferNode",
					input.Label(), n.Label()))
		}
		param_type := n.RuleSpec.ParamTypes()[i]
		input_type := input.Inputs()[0].(*TypeTestNode).Type
		if param_type != input_type {
			errors = append(errors,
				fmt.Errorf("input type %v does not match parameter type %v",
					input_type, param_type))
		}
	}
	return errors
}

func (n *RuleNode) Label() string {
	return fmt.Sprintf("rule %s", n.RuleSpec.Name())
}

// InstallRule installs a RuleNode for rule in the rete identified by
// root.
//
// InstallRule also creates a buffer node for each of the rules output
// types.
func InstallRule(root Node, rule Rule) {
	rule_node := &RuleNode {
		RuleSpec: rule,
	}
	for _, param_type := range rule.ParamTypes() {
		// *** Common code that i'm not bothering to abstract out
		// because I plan to introduce a single node type that
		// both filters by type and buffers.
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
	for _, emit_type := range rule.EmitTypes() {
		// *** Common code that i'm not bothering to abstract out
		// because I plan to introduce a single node type that
		// both filters by type and buffers.
		ttn := GetTypeTestNode(root, emit_type)		
		if ttn == nil {
			panic(fmt.Sprintf("GetTypeTestNode returned nil for %v", emit_type))
		}
		rpn := GetRuleParameterNode(ttn)
		if rpn == nil {
			panic("GetRuleParameterNode returned nil for %v")
		}
	}
}


type RuleParameterNode struct {
	BufferNode
}

// Validate is part of the Node interface.
func (n *RuleParameterNode) Validate() []error {
	errors := ValidateConnectivity(n)
	if len(n.Inputs()) != 1 {
		errors = append(errors,
			fmt.Errorf("RuleParameterNode %q should have exactly one input",
				n.Label()))
	}
	if _, ok := n.Inputs()[0].(*TypeTestNode); !ok {
		errors = append(errors,
			fmt.Errorf("the input of a RuleParameterNode (%q) should be a TypeTestNode",
				n.Label()))
	}
	return errors
}

func (n *RuleParameterNode) Label() string {
	return fmt.Sprintf("rule input %s", n.Type().String())
}

func (n *RuleParameterNode) Type() reflect.Type {
	if len(n.Inputs()) == 1 {
		i := n.Inputs()[0]
		if ttn, ok := i.(*TypeTestNode); ok {
			return ttn.Type
		}
	}
	return nil
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

// DEBUG_FILL_AND_CALL_ENTRY_HOOK is meant to allow for debugging
// output.  If non-nil, this function will be called whenever
// fill_and_call is entered.
var DEBUG_FILL_AND_CALL_ENTRY_HOOK func(*RuleParameterNode, interface{}, *RuleNode) = nil

// DEBUG_FILL_AND_CALL_MARSHAL_HOOK, if noy nil, is called as each
// parameter is marshalled prior to the invocation of a rule function.
// The arguments to the hook are the parameter position, the
// *RuleParameterNode to be providing the parameter at that position,
// and a slice containing the parameters marshalled so far.
var DEBUG_FILL_AND_CALL_MARSHAL_HOOK func(int, *RuleParameterNode, []interface{}) = nil

// DEBUG_FILL_AND_CALL_RULE_CALL_HOOK, if not nil, is called with a
// rule node and a slice containing the parameters the rule is being
// applied to.
var DEBUG_FILL_AND_CALL_RULE_CALL_HOOK func(*RuleNode, []interface{}) = nil

// Indent returns a string of count repeated copies of s.
func Indent(s string, count int) string {
	out := ""
	for i := 0; i < count; i++ {
		out += s
	}
	return out
}

func fill_and_call(in *RuleParameterNode, in_item interface{}, rule_node *RuleNode) {
	if DEBUG_FILL_AND_CALL_ENTRY_HOOK != nil {
		DEBUG_FILL_AND_CALL_ENTRY_HOOK(in, in_item, rule_node)
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
	var f func(int, bool)
	f = func (param_position int, includes_in bool) {
		if DEBUG_FILL_AND_CALL_MARSHAL_HOOK != nil {
			DEBUG_FILL_AND_CALL_MARSHAL_HOOK(param_position, in, parameters)
		}
		if param_position >= len(parameters) {
			if includes_in {
				if DEBUG_FILL_AND_CALL_RULE_CALL_HOOK != nil {
					DEBUG_FILL_AND_CALL_RULE_CALL_HOOK(rule_node, parameters)
				}
				rule_node.RuleSpec.Caller()(rule_node, parameters)
			}
			return
		}
		// Because (*RuleParameterNode).Receive adds the new
		// item to the buffer beforew fill_and_call is called,
		// we dn't need any special consideration of in_item
		// other than to avoid applying the rule redundantly
		// to parameter combinations that were considered
		// before in_item was assertred.
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

// DEBUG_RULE_PARAMETER_RECEIVE_HOOK, if not nil, is called each time
// (*RuleParameterNode).Receive is called.  The hook function is
// passed the RuleParameterNode and the item it is receiving.
var DEBUG_RULE_PARAMETER_RECEIVE_HOOK func(*RuleParameterNode, interface{}) = nil

func (node *RuleParameterNode) Receive(item interface{}) {
	if DEBUG_RULE_PARAMETER_RECEIVE_HOOK != nil {
		DEBUG_RULE_PARAMETER_RECEIVE_HOOK(node, item)
	}
	node.items = append(node.items, item)
	// A single rule might have more than one parameter of a given
	// type and thus might appear as more than one output of a
	// given RuleParameterNode.  This is a consequence of the
	// decision that a rule have one input per parameter, rather
	// than one input per parameter type.
	//
	// Each rule should be considerred only once but a rule
	// parameter should be considered (by fill_and_call) in all
	// parameter positions of the corresponding type.
	done := map[Node]bool{}
	for _, output := range node.Outputs() {
		if rule_node, ok := output.(*RuleNode); ok {
			// Consider each RuleNode we output to only once:
			if done[output] {
				continue
			}
			done[output] = true
			fill_and_call(node, item, rule_node)
		} else {
			// if node is not buffering for a RuleNode
			// then delegate to that output:
			output.Receive(item)
		}
	}
}


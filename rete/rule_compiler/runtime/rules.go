// Package runtime defines the runtime data structures that are
// created by the rule_compiler.

package runtime

import "goshua/rete"

type Rule interface {
	Name() string

	// RuleRunctionName is the name of the function that is called by
	// the Caller() function and implements the rule.
	RuleFunctionName() string

	// Installer adds the implementation of the rule to the rete
	// represented by this given root node.
	Installer() func(rete.Node)

	// Caller is the function that destructures the join results and
	// calls the rule function.
	Caller() func(rete.Node, interface{})

	// ParamTypes is the names of the types of the rule's parameters.
	ParamTypes() []string

	// EmitTypess contains the names of the types of objects that the rule can Emit.
	EmitTypes() []string
}

type rule struct {
	name string
	ruleFunctionName string
	installer func(rete.Node)
	caller func(rete.Node, interface{})
	paramTypes []string
	emitTypes []string
}

func (r *rule) Name() string { return r.name }

func (r *rule) Installer() func(rete.Node) { return r.installer }

func (r *rule) Caller() func(rete.Node, interface{}) { return r.caller }

func (r *rule) RuleFunctionName() string {
	return r.ruleFunctionName
}

func (r *rule) ParamTypes() []string { return r.paramTypes }

func (r *rule) EmitTypes() []string { return r.emitTypes }


// AllRules is a catalog describing all compiled rules that are loaded
// into the program
var AllRules []Rule = []Rule{}


// AddRule adds a rule to AllRules.  It is called by the code that is
// generated by the rule_compiler.
func AddRule(name string,
		ruleFunctionName string,
		installer func(rete.Node),
		caller func(rete.Node, interface{}),
		paramTypes []string,
		emitTypes []string) {
	AllRules = append(AllRules, &rule{
		name: name,
		ruleFunctionName: ruleFunctionName,
		installer: installer,
		caller: caller,
		paramTypes: paramTypes,
		emitTypes: emitTypes,
	})
}


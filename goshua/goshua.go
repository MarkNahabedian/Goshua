// Package goshua provides the interface types for the goshua expert system
// shell.
package goshua

import "reflect"

type KnowledgeBase interface {
	// Tell inserts a predication into the KnowledgeBase.
	// Tell can return an error if the predication can't be stored for some
	// reason, for example, if predications of that type are stored externally.
	Tell(interface{}) error

	// UnTell removes predication from the KnowledgeBase.
	UnTell(interface{}) error

	// Unsupported is a method used for truth maintenance to remove support
	// for a predication whose validity depends on a predication that has since
	// been untold.
	Unsupported(interface{})

	// AddRule adds a rule to the KnowledgeBase.
	AddRule(Rule)

	// Ask is used to query the KnowledgeBase.  continuation is called for
	// each Bindings that results from successul unification of query with
	// a value in the KnowledgeBase.
	// Ask can return an error if the predication can't be queried for from
	// external storage.
	Ask(query interface{}, continuation func(Bindings)) error
}

// NewKB returns a new, empty KnowledgeBase.
// It will get set by whatever implementation of KnowledgeBase is linked in.
var NewKb func() KnowledgeBase

// Scope disambiguates logic variables with the same name from one another.
type Scope interface {
	// Lookup returns the unique Variable with the specified name within Scope.
	// A new Variable will be created if Scope does not already have a Variable
	// with the specified name.
	Lookup(name string) Variable
}

// NewScope returns a new Scope.
// It will get set by whatever implementation of Scope is linked in.
var NewScope func() Scope

// Variable represents a logic variable.
type Variable interface {
	Unifier
	// Name returns the name of the logic variable.
	Name() string

	// IsLogicVariable does nothing.
	// Not everything with a name is a Variable.
	IsLogicVariable()

	// SameAs returns true if the receiver and other are the exact same
	// Variable: having the same name in the same Scope.
	SameAs(other Variable) bool
}

// Bindings manages the binding of logic variables.
type Bindings interface {
	// We want to be able to unify one set of bindings to another.
	Unifier

	// Bind returns a new Bindings which associates the given variable with
	// other and inherits the previous bindings of the receiver.  Other could
	// be another Variable, in which case the caller is asserting that variable
	// and other have the same value, though that value might not yet be known.
	// The second return value could be false if the new binding would cause an
	// immediate contradiction.
	Bind(variable Variable, other interface{}) (Bindings, bool)

	// Get returns the Variable's value, if it has one.
	Get(variable Variable) (value interface{}, hasValue bool)

	// Dump the bindings, for debugging.
	Dump()
}

// EmptyBindings returns a new, empty Bindings.
// EmptyBindings is set by whatever bindings implementation is linked in.
var EmptyBindings func() Bindings

// Unify implements unification.  If the two things can be unified then
// the continuation is called with the resulting Bindings as argument.
// Unify is set by whatever implementation of unification is linked in.
var Unify func(interface{}, interface{}, Bindings, func(Bindings))

// Equal implements the notion of equality used by Unify.
// Go's == operator is very strict about what it thinks are equal, for
// example int16(5) is not equal to int32(5).  We want something more
// general.
// The first return value is whether the two arguments are equal.
// The second return value is an error explaining why the values couldn't
// be compared.
var Equal func(interface{}, interface{}) (bool, error)

// Unifier is an interface for objects that can customize the behavior of Unify.
type Unifier interface {
	// Unify unifies the receiver against the provided interface{} and calls
	// the continuation with the establisged Bindings if unification was
	// successful.
	Unify(interface{}, Bindings, func(Bindings))
}

// Query is an interface for unifying and extracting fioeld values from go structs.
// A Query will also unify with another Query if their types are the same and all
// of their values unify.
type Query interface {
	Unifier
	IsQuery() bool
}

// NewQuery makes a Query for unifying against a struct of type t with field.
// values as described in fieldValues.  The values in fieldValues can be
// Variables.  itself can be a Variable to bind the object being queried
// against to if unification succeeds.
var NewQuery func(t reflect.Type, itself Variable, fieldValues map[string]interface{}) Query

// Predication types which implement their own storage (for example,
// in an external database) implement the Tellable interface.
type Tellable interface {
	// The receiver is the predication to be told.
	Tell(KnowledgeBase)

	// The receiver is the predication to be untild.
	UnTell(KnowledgeBase)

	// The receiver is the query predication
	Query(KnowledgeBase, continuation func(interface{}))
}

// Predication types that participate in backward chaining implement
// the Askable interface.
type Askable interface {
	// The receiver is the predication being asked about.
	Ask(KnowledgeBase, continuation func(Bindings))
	// DoBackwardRules
}

// We've not determined how to implement rules yet.
type Rule interface {
	IsRule()

	// If returns the predicate of the
	If() interface{} // a predication

	// Then returns the consequent of the Rule.
	Then() interface{} // a predication
}

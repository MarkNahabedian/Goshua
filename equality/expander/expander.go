// expander is a program that generates a go source file of functions which
// implement equality for different types.
//
// To build and run:
//
//   go install goshua/equality/expander
//   "%GOPATH%\bin\expander.exe"
package main

import "fmt"
import "os"
import "reflect"
import "go/ast"
import "go/parser"
import "go/printer"
import "go/token"
import "strings"
import "goshua/go_tools"

const outputFile = "${GOPATH}/src/goshua/equality/generated.go"

// equalName generates a function name for the function to test if interfaces
// of kind1 and kind2 are equal.
func equalName(kind1, kind2 reflect.Kind) string {
	return fmt.Sprintf("equal_%s_%s", kind1.String(), kind2.String())
}

// addDefs adds defs to the declarations of file.
func addDefs(defs []ast.Decl, file *ast.File) {
	for _, def := range defs {
		file.Decls = append(file.Decls, def)
	}
}

var signedIntegers = []interface{}{
	int(0), int8(0), int16(0), int32(0), int64(0),
}

var unsignedIntegers = []interface{}{
	uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
}

type equalityGroup struct {
	compareAs      string // Name of a method in reflect.Value to get the underlying type
	parameterKinds []reflect.Kind
}

var equalityGroups = []equalityGroup{
	equalityGroup{
		compareAs:      "String",
		parameterKinds: []reflect.Kind{reflect.String},
	},
	equalityGroup{
		compareAs:      "Bool",
		parameterKinds: []reflect.Kind{reflect.Bool},
	},
	equalityGroup{
		compareAs: "Int",
		parameterKinds: []reflect.Kind{
			reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
		},
	},
	equalityGroup{
		compareAs: "Uint",
		parameterKinds: []reflect.Kind{
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64,
		},
	},
	equalityGroup{
		compareAs: "Float",
		parameterKinds: []reflect.Kind{
			reflect.Float32,
			reflect.Float64,
		},
	},
}

const preamble = `// This file is generated by goshua/equality/expander.
package equality

import "reflect"
` // preamble

const equalPrototype = `
package equality
func functionName(thing1, thing2 interface{}) (bool, error) {
    v1 := reflect.ValueOf(thing1).targetType()
    v2 := reflect.ValueOf(thing2).targetType()
    return v1 == v2, nil
}

func init() {
     biadicDispatch[makeBiadicKey(kind1, kind2)] = functionName
}

` // equalPrototype

// defineEqual builds an AST Node to define a function for testing if kind1
// and kind2 are equal.  It does this by casting them both to targetKind
// before using ==.
func defineEqual(fset *token.FileSet, kind1, kind2 reflect.Kind, targetType string) []ast.Decl {
	function := go_tools.MustParse(fset, "equalPrototype", equalPrototype)
	v := go_tools.NewSubstitutingVisitor()
	v.Substitutions["kind1"] = fmt.Sprintf("reflect.%s", strings.Title(kind1.String()))
	v.Substitutions["kind2"] = fmt.Sprintf("reflect.%s", strings.Title(kind2.String()))
	v.Substitutions["functionName"] = equalName(kind1, kind2)
	v.Substitutions["targetType"] = targetType
	ast.Walk(v, function)
	return function.Decls
}

func doEqualityGroups(fset *token.FileSet, file *ast.File) {
	for _, eg := range equalityGroups {
		for _, kind1 := range eg.parameterKinds {
			for _, kind2 := range eg.parameterKinds {
				targetKind := eg.compareAs
				defs := defineEqual(fset, kind1, kind2, targetKind)
				addDefs(defs, file)
			}
		}
	}
}

const signedUnsignedPrototype = `
package equality
func integer_equal_signed_unsigned(signed, unsigned interface{}) (bool, error) {
     s := reflect.ValueOf(signed).Int()
     u := reflect.ValueOf(unsigned).Uint()
     if s < 0 {
     	return false, nil
     }
     if (u & (uint64(1) << 63)) != 0 {
     	return false, nil
     }
     if int64(u) - s != 0 {
     	return false, nil
     }
     return true, nil
}

func integer_equal_unsigned_signed(unsigned, signed interface{}) (bool, error) {
     return integer_equal_signed_unsigned(signed, unsigned)
}

` // signedUnsignedPrototype

const signedUnsignedInit = `
package equality
func init() {
     biadicDispatch[makeBiadicKey(kind1, kind2)] = integer_equal_signed_unsigned
     biadicDispatch[makeBiadicKey(kind2, kind1)] = integer_equal_unsigned_signed
}

` // signedUnsignedPrototype

func doSignedUnsigned(fset *token.FileSet, file *ast.File) {
	addDefs(go_tools.MustParse(fset, "signedUnsignedPrototype", signedUnsignedPrototype).Decls, file)
	for _, i := range signedIntegers {
		for _, u := range unsignedIntegers {
			iv := reflect.ValueOf(i)
			uv := reflect.ValueOf(u)
			function := go_tools.MustParse(fset, "signedUnsignedInit", signedUnsignedInit)
			v := go_tools.NewSubstitutingVisitor()
			v.Substitutions["kind1"] = fmt.Sprintf("reflect.%s", strings.Title(iv.Kind().String()))
			v.Substitutions["kind2"] = fmt.Sprintf("reflect.%s", strings.Title(uv.Kind().String()))
			v.Substitutions["functionName"] = equalName(iv.Kind(), uv.Kind())
			ast.Walk(v, function)
			addDefs(function.Decls, file)
		}
	}
}

func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", preamble, parser.Mode(0))
	if err != nil {
		panic(err)
	}

	doEqualityGroups(fset, file)
	doSignedUnsigned(fset, file)

	f, err := os.Create(os.ExpandEnv(outputFile))
	if err != nil {
		panic(err)
	}
	printer.Fprint(f, fset, file)
	f.Close()
}

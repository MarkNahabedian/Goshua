// rule_compiler translates a source file of rules to a go source file
// with one function definition per fule.  Each such function has the
// signature func RuleName(root_node Node) and will modify the rete rooted
// at root_node my addiing additional nodes to impelement its rule.
package main

import "flag"
import "fmt"
import "os"
import "path"
import "strings"
import "go/ast"
import "go/parser"
import "go/printer"
import "go/token"
import "go/types"

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s rule_files...\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Translates each rule file to a go source file.\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	for _, f := range flag.Args() {
		translateFile(f)
	}
}

func parseExpression(exp string) ast.Expr {
	e, err := parser.ParseExpr(exp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Offending code:\n%s\n", exp)
		panic(err)
	}
	return e
}

func parseDefinition(def string) *ast.File {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "", def, 0)
	if err != nil {
		panic(fmt.Sprintf("Errors:\n%s", err))
	}
	return astFile
}

// makeRuleInserter composes the definition of the function which
// inserts the nodes that implement a rule into the rete.
func makeRuleInserter(pkgName string, ruleDef *ast.FuncDecl) ast.Decl {
	ruleDefName := ruleDef.Name.Name
	ruleName := ruleBaseName(ruleDefName)
	funProto := fmt.Sprintf("package %s\nfunc %s(root_node rete.Node) {}",
		pkgName,
		RuleInserterName(ruleDefName))
	f := parseDefinition(funProto).Decls[0].(*ast.FuncDecl)
	body := f.Body
	addStatement := func(s ast.Stmt) {
		body.List = append(body.List, s)
	}
	typeNodes := []string{}
	// Insert the TypeFilterNodes
	for _, field := range ruleDef.Type.Params.List {
		typeString := types.ExprString(field.Type)
		for _, nameId := range field.Names {
			name := nameId.Name
			typeNodes = append(typeNodes, name)
			addStatement(&ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent(name)},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{parseExpression(fmt.Sprintf(
					"rete.GetTypeFilterNode(root_node, reflect.TypeOf((func() %s { return nil })()))",
					typeString)),
				},
			})
		}
	}
	// Insert the JoinNodes
	addStatement(&ast.DeclStmt{Decl: &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent("previous")},
				Type:  parseExpression("rete.Node"),
			},
		},
	}})
	addStatement(&ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("previous")},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{ast.NewIdent(typeNodes[len(typeNodes)-1])},
	})
	for i := len(typeNodes) - 2; i >= 0; i-- {
		addStatement(&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("previous")},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{parseExpression(fmt.Sprintf(
				"rete.Join(\"%s-%d\", %s, previous)",
				ruleName, i, typeNodes[i])),
			},
		})
	}
	// The node that calls the rule function
	addStatement(&ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("ruleNode")},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{parseExpression(
			fmt.Sprintf("rete.MakeFunctionNode(\"%s\", %s)",
				ruleName, RuleFunctionName(ruleDefName))),
		},
	})
	addStatement(&ast.ExprStmt{parseExpression("rete.Connect(previous, ruleNode)")})
	addStatement(&ast.ExprStmt{parseExpression("rete.Connect(ruleNode, root_node)")})
	return f
}

func translateFile(filename string) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Errors for file %s:\n%s\n", filename, err)
		return
	}
	pkgName := "foo"
	newDecls := []ast.Decl{}
	for _, decl := range astFile.Decls {
		rd := asRuleDefinition(decl)
		if rd == nil {
			newDecls = append(newDecls, decl)
			continue
		}
		// It's a rule definition.
		fmt.Printf("Rule definition for %s\n", ruleBaseName(rd.Name.Name))
		newDecls = append(newDecls, makeRuleInserter(pkgName, rd))
		// ***** Still need to build the rule function
	}
	astFile.Decls = newDecls
	outname := path.Join(path.Dir(filename),
		strings.TrimSuffix(path.Base(filename), ".rules")+".go")
	out, err := os.Create(outname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't create %s: %s", outname, err)
	}
	printer.Fprint(out, fset, astFile)
	out.Close()
}

/*

For each parameter of a rule we use GetTypeFilterNode on its data type
to get a Node that Emits items of that type.  The outputs of those n
Nodes are all joined together using n-1 JoinNodes and the output of
the final join flattened so that each successive element canbe bound
to the corresponding rule parameter name with appropriate type
assertions.

The rule is translated to a function which takes a rete root node as
argument.  That function inserts whatever Node subgraph is required to
impleent the rule.

That subgraph will have a TypeFilterNode for each unique type
mentioned in the rule's parameters.

It will have one fewer JoinNode (and whatever BufferNode and JoinSide
Nodes are rewuired) than the number of rule parameters.

It will have whatever FunctionNodes might be necessary to flatten the
join results or do whatever else may be necessary to facilitate the
binding of rule parameters to the corresponding elements of the join
results.

An additional FunctionNode that takes the combined join results as
input and implements the rule body.  The function of this FunctionNode
initially binds rule parameters to items in the join result.  It then
includes the translated body of the rule.  It will return if the rule
is not satisfied.  If the rule is satisfied then the function will
construct and Emit any items concluded by the rule.  The output of
this function node will be fed back to be an input of the root node.

The rule's parameter list must be translated to

  A) a type filter and join network

  B) a sequence of variable declarations in the body of the final
  FunctionNode.

The rule will result in the definition of two functions:

  A) a function which inserts nodes into an existing rete.  This will
  be an exported function haveing the same name as the rule but with
  "rule_" removed.

  B) the function that's called by the final FunctionNode that implements the rule.  It will have the

*/

const ruleNamePrefix = "rule_"

func ruleBaseName(ruleName string) string {
	return strings.TrimPrefix(ruleName, ruleNamePrefix)
}

// RuleInserterName synthesizes the canonical name for the function that
// inserts a rule implementation into a rete.
func RuleInserterName(ruleName string) string {
	return ruleBaseName(ruleName)
}

// RuleFunctionName synthesizes the canonical name for the function that
// implements the body of the rule.
func RuleFunctionName(ruleName string) string {
	return ruleBaseName(ruleName) + "Function"
}

func asRuleDefinition(astnode ast.Node) *ast.FuncDecl {
	// Test to see if this top level definition looks like a rule
	fd, ok := astnode.(*ast.FuncDecl)
	if !ok {
		return nil
	}
	if strings.HasPrefix(fd.Name.Name, ruleNamePrefix) {
		return fd
	}
	return nil
}


/* 

go build goshua/rete/rule_compiler

.\rule_compiler.exe example\example.rules 

go build goshua/rete/rule_compiler/example

*/

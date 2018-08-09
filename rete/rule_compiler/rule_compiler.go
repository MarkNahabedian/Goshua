// rule_compiler translates a source file of rules to a go source file
// with one function definition per rule.  Each such function has the
// signature func (root_node Node) and will modify the rete rooted
// at root_node by adding additional nodes to impelement its rule.
//
// That added subgraph will have a TypeTestNode for each unique type
// mentioned in the rule's parameters.
//
// It will have one fewer JoinNode (and whatever BufferNode and JoinSide
// Nodes are required) than the number of rule parameters.
//
// An additional FunctionNode that takes the combined join results as
// input and implements the rule body.  The function of this FunctionNode
// initially binds rule parameters to items in the join result.  It then
// includes the translated body of the rule.  It will return if the rule
// is not satisfied.  If the rule is satisfied then the function will
// construct and Emit any items concluded by the rule.  The output of
// this function node will be fed back to be an input of the root node.
//
// Limitations:
// Currently only rules whose parameters are struct pointer types can
// be compiled.
package main

import "bytes"
import "flag"
import "fmt"
import "os"
import "path"
import "strings"
import "text/template"
import "go/ast"
import "go/parser"
import "go/token"
import "go/types"


var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s rule_files...\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Translates each rule file to a go source file.\n")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	for _, f := range flag.Args() {
		translateFile(f)
	}
}


type ruleParameter struct {
	name string
	paramType string
}

func ruleParameters(ruleDef *ast.FuncDecl) []*ruleParameter {
	result := []*ruleParameter{}
	for _, field := range ruleDef.Type.Params.List {
		typeString := types.ExprString(field.Type)
		for _, nameId := range field.Names {
			name := nameId.Name
			result = append(result, &ruleParameter{
				name: name,
				paramType: typeString,
			})
		}
	}
	return result
}

// makeRuleInserter composes the definition of the function which
// inserts the nodes that implement a rule into the rete.
func makeRuleInserter(pkgName string, rd *ruleDefinition, params []*ruleParameter) ast.Decl {
	funProto := fmt.Sprintf("package %s\nfunc %s(root_node rete.Node) {}",
		pkgName,
		rd.ruleInserterName)
	f := parseDefinition(funProto).Decls[0].(*ast.FuncDecl)
	body := f.Body
	addStatement := func(s ast.Stmt) {
		body.List = append(body.List, s)
	}
	// For each parameter of a rule we use rete.GetTypeTestNode
	// on its data type to get a Node that Emits items of that type.
	for _, param := range params {
		addStatement(makeAssignmentStatement(token.DEFINE, param.name,
			parseExpression(fmt.Sprintf(
				"rete.GetTypeTestNode(root_node, \"%s\")",
				param.paramType))))
	}
	// The outputs of those n Nodes are all joined together using n-1 JoinNodes.
	// previous is initially set to the last TypeTestNode.  Each pre-sessive
	// TypeTestNode is Joind in turn such that the result of the final JoinNode
	// is as linked list of the joined items in the same order as in the rule's
	// parameter list.
	addStatement(makeVariableDeclaration("previous",
		 parseExpression("rete.Node"),
		ast.NewIdent(params[len(params)-1].name)))
	for i := len(params) - 2; i >= 0; i-- {
		addStatement(&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("previous")},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{parseExpression(fmt.Sprintf(
				"rete.Join(\"%s-%d\", %s, previous)",
				rd.ruleName, i, params[i].name)),
			},
		})
	}
	// The node that calls the rule function
	addStatement(makeAssignmentStatement(token.DEFINE, "ruleNode",
		parseExpression(
			fmt.Sprintf("rete.MakeFunctionNode(\"%s\", %s)",
				rd.ruleName, rd.ruleFunctionName))))
	addStatement(&ast.ExprStmt{parseExpression("rete.Connect(previous, ruleNode)")})
	addStatement(&ast.ExprStmt{parseExpression("rete.Connect(ruleNode, root_node)")})
	return f
}

func makeRuleFunction(pkgName string, rd *ruleDefinition, params []*ruleParameter) ast.Decl {
	ruleDef := rd.funcdecl
	var funProto string
	if len(params) == 1 {
		funProto = fmt.Sprintf(
			`package %s
			func %s(__node rete.Node, i interface{}) {
				%s := i.(%s)
			}`,
			pkgName, rd.ruleFunctionName, params[0].name, params[0].paramType)
	} else {
		funProto = fmt.Sprintf("package %s\nfunc %s(__node rete.Node, joinResult interface{}) {}",
			pkgName, rd.ruleFunctionName)
	}
	f := parseDefinition(funProto).Decls[0].(*ast.FuncDecl)
	body := f.Body
	addStatement := func(s ast.Stmt) {
		body.List = append(body.List, s)
	}
	if len(params) != 1 {
		// Bind each of the rule parameters to the corresponding element of
		// joinResult.
		// The variable jr is used to walk down the parameters list.
		addStatement(makeAssignmentStatement(token.DEFINE, "jr",
			parseExpression("joinResult.(rete.JoinResult)")))
		// n parameters, n-1 joins, n-2 CDRs.
		for i, param := range params[0: len(params)-1] {
			addStatement(makeVariableDeclaration(param.name,
			  	parseExpression(param.paramType),
				parseExpression(fmt.Sprintf("jr[0].(%s)", param.paramType))))
			if i < len(params) - 2 {
				addStatement(makeAssignmentStatement(token.ASSIGN, "jr",
					parseExpression("jr[1].(rete.JoinResult)")))
			}
		}
		{
			param := params[len(params) - 1]
				addStatement(makeVariableDeclaration(param.name,
				parseExpression(param.paramType),
				parseExpression(fmt.Sprintf("jr[1].(%s)", param.paramType))))
		}
	}
	// We might eventually want to do transformations on the rule body.
	body.List = append(body.List, ruleDef.Body.List...)
	return f
}

func translateFile(filename string) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Errors for file %s:\n%s\n", filename, err)
		return
	}
	fset.Iterate(func(f *token.File) bool {
		fmt.Fprintf(os.Stdout, "  fset has file %s\n", f.Name())
		return true
	})
	for _, imp := range astFile.Imports {
		fmt.Fprintf(os.Stderr, "  imports %#v %#v\n", imp.Name, imp.Path)
	}

	pkgName := astFile.Name.Name
	newAstFile := &ast.File{
		Name:     ast.NewIdent(pkgName),
		Decls:    []ast.Decl{
			makeImportDecl("goshua/rete"),
			makeImportDecl("goshua/rete/rule_compiler/runtime"),
		},
	}
	paramTypes := make(map[string]bool)
	for _, decl := range astFile.Decls {
		rd := asRuleDefinition(decl)
		if rd == nil {
			// It's not a rule definition, just copy it to the output.
			newAstFile.Decls = append(newAstFile.Decls, decl)
			continue
		}
		// It's a rule definition.
		fmt.Printf("Rule definition for %s\n", rd.ruleName)
		params := ruleParameters(rd.funcdecl)
		rf := makeRuleFunction(pkgName, rd, params)
		newAstFile.Decls = append(newAstFile.Decls,
			makeRuleInserter(pkgName, rd, params))
		newAstFile.Decls = append(newAstFile.Decls,
			rf)
		newAstFile.Decls = append(newAstFile.Decls, rd.makeAddRule())
		for _, p := range params {
			paramTypes[p.paramType] = true
		}
	}
	initFunc := makeEmptyInitFunction(pkgName)
	for pType, incl := range paramTypes {
		if incl {
			e := parseExpression(fmt.Sprintf(
				`rete.EnsureTypeTestRegistered("%s", func(i interface{}) bool { _, ok := i.(%s); return ok })`,
					pType, pType))
			debugExpressionPos(e)
			initFunc.Body.List = append(initFunc.Body.List,
				&ast.ExprStmt{ X: e})
		}
	}
	newAstFile.Decls = append(newAstFile.Decls, initFunc)
	outname := path.Join(path.Dir(filename),
		strings.TrimSuffix(path.Base(filename), ".rules")+".go")
	writeFile(fset, newAstFile, outname)
}


// ruleDefinition is used to keep track of information about a rule as it
// is being compiled.
type ruleDefinition struct {
	funcdecl *ast.FuncDecl
	ruleName string
	ruleInserterName string
	ruleFunctionName string
}

func (rd *ruleDefinition) RuleName() string { return rd.ruleName }
func (rd *ruleDefinition) RuleInserterName() string { return rd.ruleInserterName }
func (rd *ruleDefinition) RuleFunctionName() string { return rd.ruleFunctionName }

func asRuleDefinition(astnode ast.Node) *ruleDefinition {
	// Test to see if this top level definition looks like a rule
	fd, ok := astnode.(*ast.FuncDecl)
	if !ok {
		return nil
	}
	if strings.HasPrefix(fd.Name.Name, ruleNamePrefix) {
		return &ruleDefinition{
			funcdecl: fd,
			ruleName: ruleBaseName(fd.Name.Name),
			ruleInserterName: RuleInserterName(fd.Name.Name),
			ruleFunctionName: RuleFunctionName(fd.Name.Name),
		}
	}
	return nil
}

var addRuleTemplate *template.Template = template.Must(template.New("addRuleTemplate").Parse(`
package foo
func init() {
	runtime.AddRule("{{.RuleName}}", {{.RuleInserterName}}, "{{.RuleFunctionName}}")
}
`)) //

func (rd *ruleDefinition) makeAddRule() ast.Decl {
	writer := bytes.NewBufferString("")
	addRuleTemplate.Execute(writer, rd)
	parsed := parseDefinition(writer.String())
	return parsed.Decls[0]
}


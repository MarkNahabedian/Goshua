// Functions to help construct AST nodes.

package main

import "fmt"
import "os"
import "go/ast"
import "go/parser"
import "go/token"


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

func makeEmptyInitFunction(pkgName string) *ast.FuncDecl {
	astFile := parseDefinition(fmt.Sprintf("package %s\nfunc init() {}",
		pkgName))
	return astFile.Decls[0].(*ast.FuncDecl)
}

func makeImportSpec(pkgName string) *ast.ImportSpec {
	return &ast.ImportSpec{
		Path: &ast.BasicLit{ Kind: token.STRING, Value: `"` + pkgName + `"`,
		},
	}
}

func makeImportDecl(pkgName string) *ast.GenDecl {
	return &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{ makeImportSpec(pkgName) },
	}
}

func makeAssignmentStatement(tok token.Token, name string, value ast.Expr) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(name)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{value,
		},
	}
}

func makeVariableDeclaration(name string, typ ast.Expr, value ast.Expr) *ast.DeclStmt {
	s := &ast.DeclStmt{Decl: &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent(name)},
				Type:  typ,
			},
		},
	}}
	if value != nil {
		vs := s.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
		vs.Values = append(vs.Values, value)
	}
	return s
}

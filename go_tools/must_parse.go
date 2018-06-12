package go_tools

import "go/ast"
import "go/parser"
import "go/token"


func MustParse(fset *token.FileSet, sourceName string, code string) *ast.File {
	parsed, err := parser.ParseFile(fset, sourceName, code, parser.Mode(0))
	if err != nil {
		panic(err)
	}
	return parsed
}

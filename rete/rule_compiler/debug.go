package main

import "fmt"
import "go/ast"
import "os"

func debugExpressionPos(e ast.Expr) {
	fmt.Fprintf(os.Stderr, "debugExpressionPos %T: Pos: %v, End: %v\n", e, e.Pos(), e.End())
}

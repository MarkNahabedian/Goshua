package main

import "fmt"
import "go/ast"
import "os"

func debugExpressionPos(e ast.Expr) {
	fmt.Fprintf(os.Stderr, "debugExpressionPos: Pos: %v, End: %v\n", e.Pos(), e.End())
}

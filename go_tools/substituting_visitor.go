package go_tools

import "go/ast"


// SubstitutingVisitor is used to replace identifier names when walking an AST.
type SubstitutingVisitor struct {
	Substitutions map[string]string
}

func NewSubstitutingVisitor() *SubstitutingVisitor {
	return &SubstitutingVisitor{Substitutions: make(map[string]string)}
}

func (v *SubstitutingVisitor) Visit(node ast.Node) ast.Visitor {
	// node.Pos = token.NoPos
	// node.End = token.NoPos
	if n, ok := node.(*ast.Ident); ok {
		if newName, ok := v.Substitutions[n.Name]; ok {
			n.Name = newName
		}
	}
	return v
}

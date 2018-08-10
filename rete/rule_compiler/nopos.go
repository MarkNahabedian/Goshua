package main

import "fmt"
import "go/ast"
import "reflect"


// NoPos strips all position information from the ast.
func NoPos(n interface{}) interface{} {
	// Parameter and result types are not ast.Node because NoPos also
	// takes slicews.
	//
	// I've observed that the ast data structure can be circular, in particular
	// ast.Ident.
	converted := make(map[interface{}] interface{})

	note := func(n interface{}, c interface{}) interface{} {
		if c1, ok := converted[n]; ok {
			if c1 == c {
				return c
			}
			panic(fmt.Sprintf("%#v already cached.", n))
		}
		converted[n] = c
		return c
	}

	var nopos func(n interface{}) interface{}
	nopos = func(n interface{}) interface{} {
		if n == nil {
			return n
		}
		{
			v := reflect.ValueOf(n)
			switch v.Kind() {
			case reflect.Interface:
				if v.InterfaceData()[1] == uintptr(0) {
					return n
				}
			case reflect.Ptr:
				if v.Pointer() == uintptr(0) {
					return n
				}
			}
		}
		switch n1 := n.(type) {
		case *ast.AssignStmt:
			return &ast.AssignStmt{
				Lhs: nopos(n1.Lhs).([]ast.Expr),
				Tok: n1.Tok,
				Rhs: nopos(n1.Rhs).([]ast.Expr),
			}
		case *ast.BasicLit:
			return &ast.BasicLit{
				Kind: n1.Kind,
				Value: n1.Value,
			}
		case *ast.BinaryExpr:
			return &ast.BinaryExpr{
				X: nopos(n1.X).(ast.Expr),
				Op: n1.Op,
				Y: nopos(n1.Y).(ast.Expr),
			}
		case *ast.BlockStmt:
			return &ast.BlockStmt{
				List: nopos(n1.List).([]ast.Stmt),
			}
		case *ast.BranchStmt:
			return &ast.BranchStmt{
				Tok: n1.Tok,
				Label: nopos(n1.Label).(*ast.Ident),
			}
		case *ast.CallExpr:
			return &ast.CallExpr{
				Fun: nopos(n1.Fun).(ast.Expr),
				Args: nopos(n1.Args).([]ast.Expr),
			}
		case *ast.CaseClause:
			return &ast.CaseClause{
				List: nopos(n1.List).([]ast.Expr),
				Body: nopos(n1.Body).([]ast.Stmt),
			}
		case *ast.ChanType:
			return &ast.ChanType{
				Dir: n1.Dir,
				Value: nopos(n1.Value).(ast.Expr),
			}
		case *ast.CommClause:
			return &ast.CommClause{
				Comm: nopos(n1.Comm).(ast.Stmt),
				Body: nopos(n1.Body).([]ast.Stmt),
			}
		case *ast.Comment:
			return &ast.Comment{
				Text: n1.Text,
			}
		case *ast.CommentGroup:
			return &ast.CommentGroup{
				List: nopos(n1.List).([]*ast.Comment),
			}
		case *ast.CompositeLit:
			return &ast.CompositeLit{
				Type: nopos(n1.Type).(ast.Expr),
				Elts: nopos(n1.Elts).([]ast.Expr),
			}
		case *ast.DeferStmt:
			return &ast.DeferStmt{
				Call: nopos(n1.Call).(*ast.CallExpr),
			}
		case *ast.Ellipsis :
			return &ast.Ellipsis {
				Elt: nopos(n1.Elt).(ast.Expr),
			}
		case *ast.EmptyStmt:
			return &ast.EmptyStmt{
				Implicit: n1.Implicit,
			}
		case *ast.ExprStmt:
			return &ast.ExprStmt{
				X: nopos(n1.X).(ast.Expr),
			}
		case *ast.Field:
			return &ast.Field{
				Doc: nopos(n1.Doc).(*ast.CommentGroup),
				Names: nopos(n1.Names).([]*ast.Ident),
				Type: nopos(n1.Type).(ast.Expr),
				Tag: nopos(n1.Tag).(*ast.BasicLit),
				Comment: nopos(n1.Comment).(*ast.CommentGroup),
			}
		case *ast.FieldList:
			return &ast.FieldList{
				List: nopos(n1.List).([]*ast.Field),
			}
		case *ast.ForStmt:
			return &ast.ForStmt {
				Init: nopos(n1.Init).(ast.Stmt),
				Cond: nopos(n1.Cond).(ast.Expr),
				Post: nopos(n1.Post).(ast.Stmt),
				Body: nopos(n1.Body).(*ast.BlockStmt),
			}
		case *ast.FuncDecl:
			return &ast.FuncDecl{
				Doc: nopos(n1.Doc).(*ast.CommentGroup),
				Recv: nopos(n1.Recv).(*ast.FieldList),
				Name: nopos(n1.Name).(*ast.Ident),
				Type: nopos(n1.Type).(*ast.FuncType),
				Body: nopos(n1.Body).(*ast.BlockStmt),
			}
		case *ast.FuncLit:
			return &ast.FuncLit{
				Type: nopos(n1.Type).(*ast.FuncType),
				Body: nopos(n1.Body).(*ast.BlockStmt),
			}
		case *ast.FuncType:
			return &ast.FuncType{
				Params: nopos(n1.Params).(*ast.FieldList),
				Results: nopos(n1.Results).(*ast.FieldList),
			}
		case *ast.GenDecl:
			return &ast.GenDecl {
				Doc: nopos(n1.Doc).(*ast.CommentGroup),
				Tok: n1.Tok,
				Specs: nopos(n1.Specs).([]ast.Spec),
			}
		case *ast.GoStmt:
			return &ast.GoStmt{
				Call: nopos(n1.Call).(*ast.CallExpr),
			}
		case *ast.Ident:
			if c, ok := converted[n]; ok {
				return c
			}
			c := &ast.Ident{ Name: n1.Name }
			note(n, c)
			c.Obj = nopos(n1.Obj).(*ast.Object)
			return c
		case *ast.IfStmt:
			return &ast.IfStmt{
				Init: nopos(n1.Init).(ast.Stmt),
				Cond: nopos(n1.Cond).(ast.Expr),
				Body: nopos(n1.Body).(*ast.BlockStmt),
				Else: nopos(n1.Else).(ast.Stmt),
			}
		case *ast.IncDecStmt :
			return &ast.IncDecStmt{
				X: nopos(n1.X).(ast.Expr),
				Tok: n1.Tok,
			}
		case *ast.IndexExpr:
			return &ast.IndexExpr{
				X: nopos(n1.X).(ast.Expr),
				Index: nopos(n1.Index).(ast.Expr),
			}
		case *ast.InterfaceType:
			return &ast.InterfaceType{
				Methods: nopos(n1.Methods).(*ast.FieldList),
				Incomplete: n1.Incomplete,
			}
		case *ast.KeyValueExpr:
			return &ast.KeyValueExpr{
				Key: nopos(n1.Key).(ast.Expr),
				Value: nopos(n1.Value).(ast.Expr),
			}
		case *ast.LabeledStmt:
			return &ast.LabeledStmt{
				Label: nopos(n1.Label).(*ast.Ident),
				Stmt: nopos(n1.Stmt).(ast.Stmt),
			}
		case *ast.MapType:
			return &ast.MapType{
				Key: nopos(n1.Key).(ast.Expr).(ast.Expr),
				Value: nopos(n1.Value).(ast.Expr).(ast.Expr),
			}
		case *ast.Object:
			return &ast.Object{
				Kind: n1.Kind,
				Name: n1.Name,
				Decl: nopos(n1.Decl),
				Data: nopos(n1.Data),
				Type: nopos(n1.Type),
			}
	
		case *ast.RangeStmt:
			return &ast.RangeStmt{
				Key: nopos(n1.Key).(ast.Expr),
				Value: nopos(n1.Value).(ast.Expr),
				Tok: n1.Tok,
				X: nopos(n1.X).(ast.Expr),
				Body: nopos(n1.Body).(*ast.BlockStmt),
			}
		case *ast.ReturnStmt:
			return &ast.ReturnStmt{
				Results: nopos(n1.Results).([]ast.Expr),
			}
		case *ast.SelectStmt:
			return &ast.SelectStmt{
				Body: nopos(n1.Body).(*ast.BlockStmt),
			}
		case *ast.SelectorExpr:
			return &ast.SelectorExpr{
				X: nopos(n1.X).(ast.Expr),
				Sel: nopos(n1.Sel).(*ast.Ident),
			}
		case *ast.SendStmt:
			return &ast.SendStmt{
				Chan: nopos(n1.Chan).(ast.Expr),
				Value: nopos(n1.Value).(ast.Expr),
			}
		case *ast.SliceExpr:
			return &ast.SliceExpr{
				X: nopos(n1.X).(ast.Expr),
				Low: nopos(n1.Low).(ast.Expr),
				High: nopos(n1.High).(ast.Expr),
				Max: nopos(n1.Max).(ast.Expr),
				Slice3: n1.Slice3,
			}
		case *ast.StarExpr:
			return &ast.StarExpr{
				X: nopos(n1.X).(ast.Expr),
			}
		case *ast.StructType:
			return &ast.StructType{
				Fields: nopos(n1.Fields).(*ast.FieldList),
			}
		case *ast.SwitchStmt:
			return &ast.SwitchStmt{
				Init: nopos(n1.Init).(ast.Stmt),
				Tag: nopos(n1.Tag).(ast.Expr),
				Body: nopos(n1.Body).(*ast.BlockStmt),
			}
		case *ast.TypeAssertExpr:
			return &ast.TypeAssertExpr{
				X: nopos(n1.X).(ast.Expr),
				Type: nopos(n1.Type).(ast.Expr),
			}
		case *ast.TypeSpec:
			return &ast.TypeSpec{
				Doc: nopos(n1.Doc).(*ast.CommentGroup),
				Name: nopos(n1.Name).(*ast.Ident),
				Type: nopos(n1.Type).(ast.Expr),
				Comment: nopos(n1.Comment).(*ast.CommentGroup),
			}
		case *ast.TypeSwitchStmt:
			return &ast.TypeSwitchStmt{
				Init: nopos(n1.Init).(ast.Stmt),
				Assign: nopos(n1.Assign).(ast.Stmt),
				Body: nopos(n1.Body).(*ast.BlockStmt),
			}
		case *ast.UnaryExpr:
			return &ast.UnaryExpr{
				Op: n1.Op,
				X: nopos(n1.X).(ast.Expr),
			}
		case *ast.ValueSpec:
			return &ast.ValueSpec{
				Doc: nopos(n1.Doc).(*ast.CommentGroup),
				Names: nopos(n1.Names).([]*ast.Ident),
			}
	
		case []*ast.Comment:
			r := make([]*ast.Comment, len(n1))
			for i, c := range n1 {
				r[i] = nopos(c).(*ast.Comment)
			}
			return r
		case []*ast.Field:
			r := make([]*ast.Field, len(n1))
			for i, f := range n1 {
				r[i] = nopos(f).(*ast.Field)
			}
			return r
		case []*ast.Ident:
			r := make([]*ast.Ident, len(n1))
			for i, e := range n1 {
				r[i] = nopos(e).(*ast.Ident)
			}
			return r
	
		case []ast.Expr:
			r := make([]ast.Expr, len(n1))
			for i, e := range n1 {
				r[i] = nopos(e).(ast.Expr)
			}
			return r
		case []ast.Spec:
			r := make([]ast.Spec, len(n1))
			for i, s := range n1 {
				r[i] = nopos(s).(ast.Spec)
			}
			return r
		case []ast.Stmt:
			r := make([]ast.Stmt, len(n1))
			for i, s := range n1 {
				r[i] = nopos(s).(ast.Stmt)
			}
			return r
	
		default:
			panic(fmt.Sprintf("nopos does not support type %T.", n))
		}
	}

	return nopos(n)
}

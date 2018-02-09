package walk

import (
	"fmt"
	"go/ast"
)

type Visitor interface {
	Visit(node interface{}) (w Visitor)
}

//go:generate stringer -type=TypeAST
type TypeAST uint64

const (
	// 基本值定义
	_ TypeAST = 1 << (63 - iota)
	Comment
	Ident
	BadExpr
	BasicLit
	BadStmt
	BadDecl
	EmptyStmt

	CommentGroup
	Field
	FieldList
	Ellipsis
	FuncLit
	CompositeLit
	ParenExpr
	SelectorExpr
	IndexExpr
	SliceExpr
	TypeAssertExpr
	CallExpr
	StarExpr
	UnaryExpr
	BinaryExpr
	KeyValueExpr
	ArrayType
	StructType
	FuncType
	InterfaceType
	MapType
	ChanType
	DeclStmt
	LabeledStmt
	ExprStmt
	SendStmt
	IncDecStmt
	AssignStmt
	GoStmt
	DeferStmt
	ReturnStmt
	BranchStmt
	BlockStmt
	IfStmt
	CaseClause
	SwitchStmt
	TypeSwitchStmt
	CommClause
	SelectStmt
	ForStmt
	RangeStmt
	ImportSpec
	ValueSpec
	TypeSpec
	GenDecl
	FuncDecl
	File
	Package

	// 按语法分类
	LitAll  TypeAST = BasicLit | FieldList | Ellipsis | FuncLit | CompositeLit
	ExprAll TypeAST = BadExpr | ParenExpr | SelectorExpr | IndexExpr | SliceExpr | TypeAssertExpr | CallExpr | StarExpr | UnaryExpr | BinaryExpr | KeyValueExpr
	TypeAll TypeAST = ArrayType | StructType | FuncType | InterfaceType | MapType | ChanType
	StmtAll TypeAST = EmptyStmt | BadStmt | DeclStmt | LabeledStmt | ExprStmt | SendStmt | IncDecStmt | AssignStmt | GoStmt | DeferStmt | ReturnStmt | BranchStmt | BlockStmt | IfStmt | CaseClause | SwitchStmt | TypeSwitchStmt | CommClause | SelectStmt | ForStmt | RangeStmt
	SpecAll TypeAST = ImportSpec | ValueSpec | TypeSpec
	DeclAll TypeAST = BadDecl | GenDecl | FuncDecl

	AstAll TypeAST = CommentGroup | Field | LitAll | ExprAll | TypeAll | StmtAll | SpecAll | DeclAll | File | Package
)

type walkFilter func(interface{}) bool

func (f walkFilter) Visit(node interface{}) Visitor {
	if node == nil {
		return nil
	}

	if f(node) {
		return f
	}
	return nil
}

func WalkFilter(node interface{}, ta TypeAST, f func(interface{}) bool) {
	Walk(walkFilter(f), ta, node)
}

// Helper functions for common node lists. They may be empty.

func walkIdentList(v Visitor, ta TypeAST, list []*ast.Ident) {
	for _, x := range list {
		Walk(v, ta, x)
	}
}

func walkExprList(v Visitor, ta TypeAST, list []ast.Expr) {
	for _, x := range list {
		Walk(v, ta, x)
	}
}

func walkStmtList(v Visitor, ta TypeAST, list []ast.Stmt) {
	for _, x := range list {
		Walk(v, ta, x)
	}
}

func walkDeclList(v Visitor, ta TypeAST, list []ast.Decl) {
	for _, x := range list {
		Walk(v, ta, x)
	}
}

// Walk traverses an AST in depth-first order: It starts by calling
func Walk(v Visitor, ta TypeAST, node interface{}) {

	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Comments and fields
	case *ast.Comment:
		if ta&Comment == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

	case *ast.CommentGroup:
		if ta&CommentGroup == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		for _, c := range n.List {
			Walk(v, ta, c)
		}

	case *ast.Field:
		if ta&Field == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Doc != nil {
			Walk(v, ta, n.Doc)
		}
		walkIdentList(v, ta, n.Names)
		Walk(v, ta, n.Type)
		if n.Tag != nil {
			Walk(v, ta, n.Tag)
		}
		if n.Comment != nil {
			Walk(v, ta, n.Comment)
		}

	case *ast.FieldList:
		if ta&FieldList == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		for _, f := range n.List {
			Walk(v, ta, f)
		}

	// Expressions
	case *ast.BadExpr:
		if ta&BadExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

	case *ast.Ident:
		if ta&Ident == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

	case *ast.BasicLit:
		if ta&BasicLit == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

	case *ast.Ellipsis:
		if ta&Ellipsis == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Elt != nil {
			Walk(v, ta, n.Elt)
		}

	case *ast.FuncLit:
		if ta&FuncLit == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Type)
		Walk(v, ta, n.Body)

	case *ast.CompositeLit:
		if ta&CompositeLit == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Type != nil {
			Walk(v, ta, n.Type)
		}
		walkExprList(v, ta, n.Elts)

	case *ast.ParenExpr:
		if ta&ParenExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)

	case *ast.SelectorExpr:
		if ta&SelectorExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)
		Walk(v, ta, n.Sel)

	case *ast.IndexExpr:
		if ta&IndexExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)
		Walk(v, ta, n.Index)

	case *ast.SliceExpr:
		if ta&SliceExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)
		if n.Low != nil {
			Walk(v, ta, n.Low)
		}
		if n.High != nil {
			Walk(v, ta, n.High)
		}
		if n.Max != nil {
			Walk(v, ta, n.Max)
		}

	case *ast.TypeAssertExpr:
		if ta&TypeAssertExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)
		if n.Type != nil {
			Walk(v, ta, n.Type)
		}

	case *ast.CallExpr:
		if ta&CallExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Fun)
		walkExprList(v, ta, n.Args)

	case *ast.StarExpr:
		if ta&StarExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)

	case *ast.UnaryExpr:
		if ta&UnaryExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)

	case *ast.BinaryExpr:
		if ta&BinaryExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)
		Walk(v, ta, n.Y)

	case *ast.KeyValueExpr:
		if ta&KeyValueExpr == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Key)
		Walk(v, ta, n.Value)

	// Types
	case *ast.ArrayType:
		if ta&ArrayType == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Len != nil {
			Walk(v, ta, n.Len)
		}
		Walk(v, ta, n.Elt)

	case *ast.StructType:
		if ta&StructType == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Fields)

	case *ast.FuncType:
		if ta&FuncType == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Params != nil {
			Walk(v, ta, n.Params)
		}
		if n.Results != nil {
			Walk(v, ta, n.Results)
		}

	case *ast.InterfaceType:
		if ta&InterfaceType == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Methods)

	case *ast.MapType:
		if ta&MapType == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Key)
		Walk(v, ta, n.Value)

	case *ast.ChanType:
		if ta&ChanType == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Value)

	// Statements
	case *ast.BadStmt:
		if ta&BadStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

	case *ast.DeclStmt:
		if ta&DeclStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Decl)

	case *ast.EmptyStmt:
		if ta&EmptyStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

	case *ast.LabeledStmt:
		if ta&LabeledStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Label)
		Walk(v, ta, n.Stmt)

	case *ast.ExprStmt:
		if ta&ExprStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)

	case *ast.SendStmt:
		if ta&SendStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Chan)
		Walk(v, ta, n.Value)

	case *ast.IncDecStmt:
		if ta&IncDecStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.X)

	case *ast.AssignStmt:
		if ta&AssignStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		walkExprList(v, ta, n.Lhs)
		walkExprList(v, ta, n.Rhs)

	case *ast.GoStmt:
		if ta&GoStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Call)

	case *ast.DeferStmt:
		if ta&DeferStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Call)

	case *ast.ReturnStmt:
		if ta&ReturnStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		walkExprList(v, ta, n.Results)

	case *ast.BranchStmt:
		if ta&BranchStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Label != nil {
			Walk(v, ta, n.Label)
		}

	case *ast.BlockStmt:
		if ta&BlockStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		walkStmtList(v, ta, n.List)

	case *ast.IfStmt:
		if ta&IfStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Init != nil {
			Walk(v, ta, n.Init)
		}
		Walk(v, ta, n.Cond)
		Walk(v, ta, n.Body)
		if n.Else != nil {
			Walk(v, ta, n.Else)
		}

	case *ast.CaseClause:
		if ta&CaseClause == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		walkExprList(v, ta, n.List)
		walkStmtList(v, ta, n.Body)

	case *ast.SwitchStmt:
		if ta&SwitchStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Init != nil {
			Walk(v, ta, n.Init)
		}
		if n.Tag != nil {
			Walk(v, ta, n.Tag)
		}
		Walk(v, ta, n.Body)

	case *ast.TypeSwitchStmt:
		if ta&TypeSwitchStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Init != nil {
			Walk(v, ta, n.Init)
		}
		Walk(v, ta, n.Assign)
		Walk(v, ta, n.Body)

	case *ast.CommClause:
		if ta&CommClause == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Comm != nil {
			Walk(v, ta, n.Comm)
		}
		walkStmtList(v, ta, n.Body)

	case *ast.SelectStmt:
		if ta&SelectStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		Walk(v, ta, n.Body)

	case *ast.ForStmt:
		if ta&ForStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Init != nil {
			Walk(v, ta, n.Init)
		}
		if n.Cond != nil {
			Walk(v, ta, n.Cond)
		}
		if n.Post != nil {
			Walk(v, ta, n.Post)
		}
		Walk(v, ta, n.Body)

	case *ast.RangeStmt:
		if ta&RangeStmt == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Key != nil {
			Walk(v, ta, n.Key)
		}
		if n.Value != nil {
			Walk(v, ta, n.Value)
		}
		Walk(v, ta, n.X)
		Walk(v, ta, n.Body)

	// Declarations
	case *ast.ImportSpec:
		if ta&ImportSpec == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Doc != nil {
			Walk(v, ta, n.Doc)
		}
		if n.Name != nil {
			Walk(v, ta, n.Name)
		}
		Walk(v, ta, n.Path)
		if n.Comment != nil {
			Walk(v, ta, n.Comment)
		}

	case *ast.ValueSpec:
		if ta&ValueSpec == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Doc != nil {
			Walk(v, ta, n.Doc)
		}
		walkIdentList(v, ta, n.Names)
		if n.Type != nil {
			Walk(v, ta, n.Type)
		}
		walkExprList(v, ta, n.Values)
		if n.Comment != nil {
			Walk(v, ta, n.Comment)
		}

	case *ast.TypeSpec:
		if ta&TypeSpec == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Doc != nil {
			Walk(v, ta, n.Doc)
		}
		Walk(v, ta, n.Name)
		Walk(v, ta, n.Type)
		if n.Comment != nil {
			Walk(v, ta, n.Comment)
		}

	case *ast.BadDecl:
		if ta&BadDecl == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

	case *ast.GenDecl:
		if ta&GenDecl == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Doc != nil {
			Walk(v, ta, n.Doc)
		}
		for _, s := range n.Specs {
			Walk(v, ta, s)
		}

	case *ast.FuncDecl:
		if ta&FuncDecl == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Doc != nil {
			Walk(v, ta, n.Doc)
		}
		if n.Recv != nil {
			Walk(v, ta, n.Recv)
		}
		Walk(v, ta, n.Name)
		Walk(v, ta, n.Type)
		if n.Body != nil {
			Walk(v, ta, n.Body)
		}

	// Files and packages
	case *ast.File:
		if ta&File == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		if n.Doc != nil {
			Walk(v, ta, n.Doc)
		}
		Walk(v, ta, n.Name)
		walkDeclList(v, ta, n.Decls)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		if ta&Package == 0 {
			break
		}
		if v = v.Visit(node); v == nil {
			return
		}

		for _, f := range n.Files {
			Walk(v, ta, f)
		}

	case map[string]*ast.Package:
		for _, f := range n {
			Walk(v, ta, f)
		}
	default:
		panic(fmt.Sprintf("Walk: unexpected node type %T", n))
	}

	// v.Visit(nil)
}

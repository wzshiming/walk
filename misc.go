package walk

import (
	"go/ast"
	"strings"
)

const (
	Dot   = "." // 分隔 类型和成员
	Colon = ":" // 分隔 类型和方法
	Sharp = "#" //
	At    = "@" //
	Space = " " //
)

func getNameSuf(expr interface{}) string {
	name := getName(expr)
	i := strings.Index(name, Dot)
	if i != -1 {
		return name[i+1:]
	}
	return name
}

func getName(expr interface{}) string {
	switch b := expr.(type) {
	case *ast.Ident:
		return b.Name
	case *ast.StarExpr:
		return getName(b.X)
	case *ast.SelectorExpr:
		return getName(b.X) + Dot + getName(b.Sel)
	case *ast.BasicLit:
		return b.Value
	default:
		return ""
	}
}

type Type struct {
	expr ast.Expr
}

func NewType(expr ast.Expr) *Type {
	return &Type{
		expr: expr,
	}
}

func (t *Type) String() string {
	return getName(t.expr)
}

// 判断一个名字是否是公开的
func IsExported(name string) bool {
	for _, v := range strings.Split(name, Colon) {
		if !ast.IsExported(v) {
			return false
		}
	}
	return true
}

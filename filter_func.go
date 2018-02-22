package walk

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"strings"
)

func FilterImport(i interface{}) bool {
	return i.(*ast.GenDecl).Tok == token.IMPORT
}

func FilterConst(i interface{}) bool {
	return i.(*ast.GenDecl).Tok == token.CONST
}

func FilterVar(i interface{}) bool {
	return i.(*ast.GenDecl).Tok == token.VAR
}

func FilterType(i interface{}) bool {
	return i.(*ast.GenDecl).Tok == token.TYPE
}

func FilterName(name string) func(interface{}) bool {
	return func(node interface{}) bool {
		return filterName(name, node)
	}
}

func filterName(name string, node interface{}) bool {
	switch n := node.(type) {
	case *ast.Ident:
		return n.Name == name

	// Fields
	case *ast.Field:
		for _, v := range n.Names {
			if filterName(name, v) {
				return true
			}
		}
	case *ast.FieldList:
		for _, v := range n.List {
			if filterName(name, v) {
				return true
			}
		}

	// Types
	case *ast.StructType:
		return filterName(name, n.Fields)
	case *ast.InterfaceType:
		return filterName(name, n.Methods)

	// Specs
	case *ast.ImportSpec:
		if n.Name != nil {
			return filterName(name, n.Name)
		}

		p := path.Base(strings.Trim(n.Path.Value, `"`))
		if i := strings.LastIndex(p, "."); i != -1 {
			p = p[:i]
		}
		return p == name

	case *ast.ValueSpec:
		for _, v := range n.Names {
			if filterName(name, v) {
				return true
			}
		}

	case *ast.TypeSpec:
		return filterName(name, n.Name)
	case *ast.GenDecl:
		for _, v := range n.Specs {
			if filterName(name, v) {
				return true
			}
		}
	case *ast.FuncDecl:
		return filterName(name, n.Name)
	case *ast.File:
		for _, v := range n.Decls {
			if filterName(name, v) {
				return true
			}
		}
	case *ast.Package:
		for _, v := range n.Files {
			if filterName(name, v) {
				return true
			}
		}
	case map[string]*ast.Package:
		_, ok := n[name]
		return ok
	default:
		panic(fmt.Sprintf("FilterName: unexpected node type %T", n))
	}
	return false
}

func GetChild(nodes ...interface{}) map[string]interface{} {
	r := map[string]interface{}{}
	for _, node := range nodes {
		getChild(node, r)
	}
	return r
}

func getChild(node interface{}, nexts map[string]interface{}) {
	switch n := node.(type) {
	case *ast.Ident:
		return

	// Fields
	case *ast.Field:
		for _, v := range n.Names {
			nexts[v.String()] = n
		}
		return
	case *ast.FieldList:
		for _, v := range n.List {
			getChild(v, nexts)
		}
		return
	// Types
	case *ast.StructType:
		getChild(n.Fields, nexts)
		return
	case *ast.InterfaceType:
		getChild(n.Methods, nexts)
		return

	// Specs
	case *ast.ImportSpec:
		if n.Name != nil {
			nexts["import."+n.Name.String()] = n
			return
		}

		p := path.Base(strings.Trim(n.Path.Value, `"`))
		if i := strings.LastIndex(p, "."); i != -1 {
			p = p[:i]
		}
		nexts["import."+p] = n
		return

	case *ast.ValueSpec:
		for _, v := range n.Names {
			nexts["var."+v.Name] = n
		}
		return
	case *ast.TypeSpec:
		nexts["type."+n.Name.String()] = n
		return
	case *ast.GenDecl:
		for _, v := range n.Specs {
			getChild(v, nexts)
		}
	case *ast.FuncDecl:
		if n.Recv == nil {
			nexts["func."+n.Name.String()] = n
		} else {
			tt := n.Recv.List[0].Type
			switch t := tt.(type) {
			case *ast.StarExpr:
				tt = t.X
			}
			nexts["func."+fmt.Sprint(tt)+"."+n.Name.String()] = n
		}
		return
	case *ast.File:
		for _, v := range n.Decls {
			getChild(v, nexts)
		}
		return
	case *ast.Package:
		for _, v := range n.Files {
			getChild(v, nexts)
		}
		return
	case map[string]*ast.Package:
		for _, v := range n {
			getChild(v, nexts)
		}
		return
	default:
		panic(fmt.Sprintf("getNext: unexpected node type %T", n))
	}
	return
}

package gowalk

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"path/filepath"
	"strings"
)

type Node struct {
	name string
	p    *walk
	tar  []interface{} // 索引0 是名字 等于的节点  1 是当前节点 接下来的 是到这里检索的路径
}

func newNode(name string, p *walk, tar []interface{}) *Node {
	return &Node{
		name: name,
		p:    p,
		tar:  append([]interface{}{}, tar...),
	}
}

func (w *Node) Doc() *ast.CommentGroup {
	if w == nil {
		return nil
	}
	tar := w.Value()
	switch b := tar.(type) {
	case *ast.Field:
		return b.Doc
	case *ast.ImportSpec:
		return b.Doc
	case *ast.ValueSpec:
		return b.Doc
	case *ast.TypeSpec:
		return b.Doc
	case *ast.GenDecl:
		return b.Doc
	case *ast.FuncDecl:
		return b.Doc
	}
	return &ast.CommentGroup{}
}

// 输出当前节点源码
func (w *Node) Src() string {
	n, _ := w.Value().(ast.Node)
	if n == nil {
		return ""
	}
	buf := bytes.NewBuffer(nil)
	printer.Fprint(buf, w.p.fileSet, n)
	return buf.String()
}

// 输出源码位置
func (w *Node) Pos() token.Position {
	n, _ := w.Value().(ast.Node)
	if n == nil {
		return token.Position{}
	}
	return w.p.fileSet.Position(n.Pos())
}

func (w *Node) Name() string {
	return w.name
}

// 值
func (w *Node) Value() interface{} {
	if w == nil {
		return nil
	}
	return w.tar[1]
}

// 查询子节点
func (w *Node) Child(name ...string) *Node {
	if w == nil {
		return nil
	}
	n := w
	ss := []string{}
	for _, v := range name {
		ss = append(ss, strings.Split(v, Dot)...)
	}
	if len(ss) == 0 {
		ss = append(ss, "")
	}
	for _, v := range ss {
		n = n.ChildForm(v)
	}
	return n
}

// 进入子节点
func (w *Node) ChildForm(name string) *Node {
	if w == nil {
		return nil
	}
	l := w.parse(w.Value(), name)
	if len(l) > 1 {
		return newNode(name, w.p, l)
	}
	t := w.Type()
	if t != nil {
		return t.ChildForm(name)
	}
	s := w.Name()
	if s != "" {
		return w.p.root.Child(s)
	}
	return nil
}

// 取类型
func (w *Node) Type() *Node {
	if w == nil {
		return nil
	}
	t := w.typ()
	if t == nil {
		return nil
	}
	return newNode(getName(t), w.p, append([]interface{}{nil, t}, w.tar[1:]...))
}

func (w *Node) typ() ast.Expr {
	if w == nil {
		return nil
	}
	tar := w.Value()
	switch b := tar.(type) {
	case *ast.ValueSpec: // var const 里的一条定义
		return b.Type
	case *ast.TypeSpec:
		return b.Type
	case *ast.Field:
		return b.Type
	case *ast.FuncDecl:
		return b.Type
	}
	return nil
}

// 获取所有子节点列表
func (w *Node) ChildList() []string {
	if w == nil {
		return nil
	}
	r := w.getChildList(w.Value())
	if len(r) != 0 && r[0] == w.name {
		r = r[1:]
	}

	if len(r) == 0 {
		t := w.Type()
		if t != nil {
			return t.ChildList()
		}
	}
	return r
}

// 定位到节点
func (w *Node) parse(tar interface{}, name string) (r []interface{}) {
	switch b := tar.(type) {
	case map[string]*ast.Package: // 文件夹
		if name == "" {
			return []interface{}{nil, b}
		}
		for _, v := range b {
			if r = w.parse(v.Files, name); r != nil {
				break
			}
		}
	case map[string]*ast.File: // 包
		for _, v := range b {
			if r = w.parse(v.Decls, name); r != nil {
				break
			}
		}
	case []ast.Decl: // 顶级关键字
		for _, v := range b {
			if r = w.parse(v, name); r != nil {
				break
			}
		}
	case *ast.GenDecl: // import type const var
		r = w.parse(b.Specs, name)
	case *ast.FuncDecl: // func
		s := ""
		if b.Recv != nil {
			if len(b.Recv.List) == 1 {
				s = getName(b.Recv.List[0].Type)
				if s != "" {
					s = s + Colon
				}
			}
		}
		if s+b.Name.String() == name {
			return []interface{}{b.Name, b}
		}
	case []ast.Spec: // 顶级关键字内容
		for _, v := range b {
			if r = w.parse(v, name); r != nil {
				break
			}
		}
	case *ast.ImportSpec: // import 里的一条定义
		s := ""
		path := strings.Replace(getName(b.Path), `"`, ``, -1)
		if b.Name == nil {
			s = filepath.Base(path)
		} else {
			s = getName(b.Name)
		}
		if s == name {
			pkg, _ := w.p.open(path)
			return w.parse(pkg, "")
		}
	case *ast.ValueSpec: // var const 里的一条定义
		for _, v := range b.Names {
			if r = w.parse(v, name); r != nil {
				break
			}
		}
	case *ast.TypeSpec: // type 的一条定义
		r = w.parse(b.Name, name)
	case *ast.StructType: // token struct
		r = w.parse(b.Fields, name)
	case *ast.FieldList: // token field
		r = w.parse(b.List, name)
	case []*ast.Field:
		for _, v := range b {
			if r = w.parse(v, name); r != nil {
				break
			}
		}
	case *ast.Field:
		if b.Names != nil {
			for _, v := range b.Names {
				if r = w.parse(v, name); r != nil {
					break
				}
			}
		} else {
			n := getNameSuf(b.Type) // 组合的字段
			if n == name {
				return []interface{}{b.Type, b}
			}
		}
	case []*ast.Ident:
		for _, v := range b {
			if getName(v) == name {
				return []interface{}{v}
			}
		}
	case *ast.Ident:
		if getName(b) == name {
			return []interface{}{b}
		}
	}
	if len(r) != 0 {
		r = append(r, tar)
	}
	return
}

// 获取当前节点可以走的子节点
func (w *Node) getChildList(tar interface{}) (r []string) {
	switch b := tar.(type) {
	case map[string]*ast.Package: // 文件夹
		for _, v := range b {
			r = append(r, w.getChildList(v.Files)...)
		}
	case map[string]*ast.File: // 包
		for _, v := range b {
			r = append(r, w.getChildList(v.Decls)...)
		}
	case []ast.Decl: // 顶级关键字
		for _, v := range b {
			r = append(r, w.getChildList(v)...)
		}
	case *ast.GenDecl: // import type const var
		r = w.getChildList(b.Specs)
	case *ast.FuncDecl: // func
		s := ""
		if b.Recv != nil {
			if len(b.Recv.List) == 1 {
				s = getName(b.Recv.List[0].Type)
				if s != "" {
					s = s + Colon
				}
			}
		}
		r = append(r, s+b.Name.String())
	case []ast.Spec: // 顶级关键字内容
		for _, v := range b {
			r = append(r, w.getChildList(v)...)
		}
	case *ast.ImportSpec: // import 里的一条定义
		s := ""
		path := strings.Replace(getName(b.Path), `"`, ``, -1)
		if b.Name == nil {
			s = filepath.Base(path)
		} else {
			s = getName(b.Name)
		}
		r = append(r, s)
	case *ast.ValueSpec: // var const 里的一条定义
		for _, v := range b.Names {
			r = append(r, getName(v))
		}
	case *ast.TypeSpec: // type 的一条定义
		name := getName(b.Name)
		r = []string{name}
	case *ast.StructType: // token struct
		r = w.getChildList(b.Fields)
	case *ast.FieldList: // token field
		r = w.getChildList(b.List)
	case []*ast.Field:
		for _, v := range b {
			r = append(r, w.getChildList(v)...)
		}
	case *ast.Field:
		if b.Names != nil {
			for _, v := range b.Names {
				r = append(r, getName(v))
			}
		} else {
			r = append(r, getNameSuf(b.Type))
		}
	}
	return
}

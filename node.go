package walk

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"path/filepath"
	"strings"
)

type Node struct {
	name string
	walk *walk
	tar  []interface{} // 索引0 是名字 等于的节点  1 是当前节点 接下来的 是到这里检索的路径
}

func newNode(name string, walk *walk, tar []interface{}) *Node {
	return &Node{
		name: name,
		walk: walk,
		tar:  append([]interface{}{}, tar...),
	}
}

func (w *Node) Save() error {
	n, _ := w.Value().(ast.Node)
	if n == nil {
		return fmt.Errorf("当前节点无法保存")
	}

	f := w.walk.fileSet.Position(n.Pos())
	return w.walk.save(f.Filename)
}

func (w *Node) in(name string, v ...interface{}) *Node {
	return newNode(name, w.walk, append(v, w.tar[1:]...))
}

// 取变量
func (w *Node) Var(name string) *Node {
	if w == nil {
		return nil
	}
	tar := w.Value()
	var ss []ast.Stmt
	switch b := tar.(type) {
	case *ast.FuncDecl:
		v := w.Type().Var(name)
		if v != nil {
			return v
		}
		n := w.Body().Var(name)
		if n != nil {
			return n
		}
		if b.Recv == nil || len(b.Recv.List) == 0 {
			return nil
		}
		return w.in(name, nil, b.Recv.List[0].Type)
	case *ast.FuncType: // 函数参数里取变量
		if b.Params != nil {
			v := w.parse(b.Params, name)
			if v != nil {
				return w.in(name, v...)
			}
		}
		if b.Results != nil {
			v := w.parse(b.Results, name)
			if v != nil {
				return w.in(name, v...)
			}
		}
		return nil
	case *ast.BlockStmt:
		ss = b.List
	case []ast.Stmt:
		ss = b
	default:
		return w.childForm(name)
	}
	for _, v := range ss {
		switch va := v.(type) {
		case *ast.AssignStmt:
			v2 := w.parse(va, name)
			if len(v2) != 0 {
				return w.in(name, v2...)
			}
		case *ast.DeclStmt:
			v2 := w.parse(va.Decl, name)
			if len(v2) != 0 {
				return w.in(name, v2...)
			}
		}
	}
	return w.childForm(name)
}

// 内容  花括号里面的
func (w *Node) Body() *Node {
	if w == nil {
		return nil
	}
	tar := w.Value()
	var l *ast.BlockStmt
	switch b := tar.(type) {
	case *ast.FuncDecl:
		l = b.Body
	case *ast.FuncLit:
		l = b.Body
	case *ast.IfStmt:
		l = b.Body // 返回 成功的
	case *ast.SwitchStmt:
		l = b.Body
	case *ast.TypeSwitchStmt:
		l = b.Body
	case *ast.SelectStmt:
		l = b.Body
	case *ast.ForStmt:
		l = b.Body
	case *ast.BlockStmt:
		l = b
	default:
		return nil
	}

	return w.in("", nil, l)
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

func (w *Node) Comment() *ast.CommentGroup {
	if w == nil {
		return nil
	}
	tar := w.Value()
	switch b := tar.(type) {
	case *ast.Field:
		return b.Comment
	case *ast.ImportSpec:
		return b.Comment
	case *ast.ValueSpec:
		return b.Comment
	case *ast.TypeSpec:
		return b.Comment

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
	printer.Fprint(buf, w.walk.fileSet, n)
	return buf.String()
}

// 输出源码位置
func (w *Node) Pos() token.Position {
	n, _ := w.Value().(ast.Node)
	if n == nil {
		return token.Position{}
	}
	return w.walk.fileSet.Position(n.Pos())
}

func (w *Node) Tars() []interface{} {
	return w.tar
}

func (w *Node) Ident() interface{} {
	if w == nil {
		return nil
	}
	return w.tar[0]
}

func (w *Node) Name() string {
	if w == nil {
		return ""
	}
	if w.name != "" {
		return w.name
	}
	v := w.Ident()
	if v != nil {
		return getName(v)
	}
	return ""
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
		n = n.childForm(v)
	}
	return n
}

func (w *Node) childTypeImport(name string) *Node {
	// 在类型下找
	t := w.Return()
	if t != nil {
		s := t.Name()
		if s == "" {
			s = name
		} else {
			s = s + Dot + name
		}
		return t.Child(s)
	}
	return nil
}

// 进入子节点
func (w *Node) childForm(name string) *Node {
	if w == nil {
		return nil
	}

	// 直接在当前 节点下找
	if l := w.parse(w.Value(), name); len(l) > 1 {
		return newNode(name, w.walk, l)
	}

	// 在类型下找
	if n := w.childTypeImport(name); n != nil {
		return n
	}

	// 查找类型的方法
	if typ := w.Name(); typ != "" {
		if v := w.walk.root.Child(typ + ":" + name); v != nil {
			return v
		}
	}

	// 在根目录下找
	if w.walk.root != w {
		return w.walk.root.childForm(name)
	}
	return nil
}

func (w *Node) parseType(tar ast.Expr) string {
	switch b := tar.(type) {
	case *ast.CallExpr:
		return w.parseType(b.Fun)
	case *ast.SelectorExpr:
		//		s := w.parse(b.X) + Dot + getName(b.Sel)
		//		ffmt.Mark(s, w.Child(s).Src())
	}
	return ""
}

//// 表达式执行完的结果
//func (w *Node) Result() *Node {
//	if w == nil {
//		return nil
//	}
//	tar := w.Value()
//	switch b := tar.(type) {
//	case ast.BadExpr:
//	}
//}

// 返回值 的类型
func (w *Node) Return() *Node {
	if w == nil {
		return nil
	}
	tar := w.Value()

	var t *ast.FuncType
	switch b := tar.(type) {
	case *ast.FuncDecl:
		t = b.Type
	case *ast.FuncType:
		t = b
	default:
		return w.Type()
	}
	return w.in("", nil, t.Results)
}

// 取类型
func (w *Node) Type() *Node {
	if w == nil {
		return nil
	}
	tar := w.Value()

	var t ast.Expr
	switch b := tar.(type) {
	case *ast.ValueSpec: // var const 里的一条定义
		t = b.Type
	case *ast.TypeSpec:
		t = b.Type
	case *ast.Field:
		t = b.Type
	case *ast.FuncDecl:
		t = b.Type
	default:
		//ffmt.Mark(ffmt.Sp(b))
		return nil
	}

	return w.in(getName(t), nil, t)
}

// 获取所有子节点列表
func (w *Node) ChildList() []string {
	if w == nil {
		return nil
	}
	r := w.getChildList(w.Value())
	if len(r) != 0 {
		if r[0] == w.name {
			if len(r) != 1 {
				return r[1:]
			}
		} else {
			return r
		}
	}

	//ffmt.Puts(w.Type().Value(), w.Type().ChildList())
	//	s := w.Type().ChildList()
	//	ffmt.Mark(s)
	//	//	if s == "" {
	//	//		s = name
	//	//	} else {
	//	//		s = s + Dot + name
	//	//	}
	//	//	return t.Child(s)

	return w.Type().ChildList()
}

// 定位到节点
func (w *Node) parse(tar interface{}, name string) (r []interface{}) {
	if tar == nil {
		return
	}
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
			pkg, _ := w.walk.open(path)
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
		if b.List != nil {
			r = w.parse(b.List, name)
		}
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

	case *ast.AssignStmt: // :=
		if b.Tok == token.DEFINE {
			r = w.parse(b.Lhs, name)
		}
	case []ast.Expr:
		for _, v := range b {
			if getName(v) == name {
				return []interface{}{v}
			}
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

func (w *Node) Index(i int) *Node {
	return w.index(w.Value(), i)
}

// 列表索引
func (w *Node) index(tar interface{}, i int) *Node {

	t := []*ast.Field{}
	switch b := tar.(type) {
	case *ast.FieldList:
		t = b.List
	case []*ast.Field:
		t = b
	default:
		return nil
	}
	sum := 0
	for _, v := range t {
		m := len(v.Names)
		if m == 0 {
			m = 1
		}
		if sum+m > i {
			if len(v.Names) != 0 {
				n := v.Names[i-sum]
				return w.in(getName(n), n, v)
			}
			return w.in("", nil, v)
		}
		sum += m
	}
	return nil
}

// 列表长度
func (w *Node) Len() int {
	return w.len(w.Value())
}

// 列表长度
func (w *Node) len(tar interface{}) int {
	t := []*ast.Field{}
	switch b := tar.(type) {
	case *ast.FieldList:
		t = b.List
	case []*ast.Field:
		t = b
	default:
		return -1
	}
	sum := 0
	for _, v := range t {
		s := len(v.Names)
		if s == 0 {
			s = 1
		}
		sum += s
	}
	return sum
}

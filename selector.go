package walk

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
)

type Selector struct {
	selectorItems

	fset   *token.FileSet
	gopath []string
	mode   parser.Mode
	pkgs   map[string]map[string]*ast.Package

	stack []*Selector
}

func NewSelectorMust(path string) *Selector {
	s, err := NewSelector(path)
	if err != nil {
		panic(err)
	}
	return s
}

func NewSelector(path string) (*Selector, error) {
	p, err := getPath()
	if err != nil {
		return nil, err
	}
	s := &Selector{
		mode:   parser.ParseComments,
		gopath: p,
		fset:   token.NewFileSet(),
		pkgs:   map[string]map[string]*ast.Package{},
	}
	err = s.Load(path)
	if err != nil {
		return nil, err
	}
	s.selectorItems = *newSelectorItems(s, []interface{}{s.pkgs[path]})

	return s, nil
}

func (s *Selector) clone(s0 *selectorItems) *Selector {
	n := *s
	n.selectorItems = *s0
	return &n
}

func (s *Selector) Push() *Selector {
	s.stack = append(s.stack, s)
	return s
}

func (s *Selector) Pop() *Selector {
	if len(s.stack) == 0 {
		return nil
	}
	s0 := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return s0
}

func (s *Selector) Back() *Selector {
	p := s.selectorItems.Back()
	if p == nil {
		return nil
	}
	return p.selector
}

// Load Package to resolve in AST from gopath
func (s *Selector) Load(path string) (first error) {
	if _, ok := s.pkgs[path]; ok {
		return nil
	}

	for _, v := range s.gopath {
		dir := filepath.Join(v, path)
		pkg, err := parser.ParseDir(s.fset, dir, nil, s.mode)
		if err == nil {
			s.pkgs[path] = pkg
			return nil
		}
		if first == nil {
			first = err
		}
	}
	return
}

func (s *Selector) Walk(ta TypeAST, f ...func(interface{}) bool) *Selector {
	return s.clone(s.selectorItems.Walk(ta, f...))
}

func (s *Selector) Slice(i, j int) *Selector {
	return s.clone(s.selectorItems.Slice(i, j))
}

func (s *Selector) First() *Selector {
	return s.clone(s.selectorItems.First())
}

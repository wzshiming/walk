package walk

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
)

type selectorItems struct {
	list     []interface{}
	back     *selectorItems
	selector *Selector
}

func newSelectorItems(selector *Selector, list []interface{}) *selectorItems {
	return &selectorItems{
		list:     list,
		selector: selector,
	}
}

func (s *selectorItems) First() *selectorItems {
	return s.Slice(0, 1)
}

func (s *selectorItems) Slice(i, j int) *selectorItems {
	l := len(s.list)
	if j > l {
		j = l
	}
	if i > l {
		i = l
	}
	return s.clone(s.list[i:j])
}

func (s *selectorItems) Len() int {
	return len(s.list)
}

func (s *selectorItems) List() []interface{} {
	return s.list
}

func (s *selectorItems) clone(list []interface{}) *selectorItems {
	return &selectorItems{
		list:     list,
		back:     s,
		selector: s.selector,
	}
}

func (s *selectorItems) fset() *token.FileSet {
	return s.selector.fset
}

func (s *selectorItems) Back() *selectorItems {
	return s.back
}

func (s *selectorItems) Walk(ta TypeAST, f ...func(interface{}) bool) *selectorItems {
	l := []interface{}{}

	var f0 func(interface{}) bool
	switch len(f) {
	case 0:
		f0 = func(i interface{}) bool {
			l = append(l, i)
			return true
		}
	case 1:
		f0 = func(i interface{}) bool {
			if f[0](i) {
				l = append(l, i)
				return true
			}
			return false
		}
	default:
		f0 = func(i interface{}) bool {
			for _, v := range f {
				ok := v(i)
				if !ok {
					return false
				}
			}
			l = append(l, i)
			return true
		}
	}
	for _, v := range s.list {
		WalkFilter(v, ta, f0)
	}
	return s.clone(l)
}

func (s *selectorItems) String() string {
	ss, err := s.Format()
	if err != nil {
		return err.Error()
	}
	return ss
}

// Format AST formatting to text
func (s *selectorItems) Format() (string, error) {
	buf := bytes.NewBuffer(nil)
	for _, v := range s.list {
		switch t := v.(type) {
		case *ast.Package:
			for k, v := range t.Files {
				buf.WriteString(fmt.Sprintf("\n// with %s: \n", k))
				err := printer.Fprint(buf, s.fset(), v)
				if err != nil {
					return "", err
				}
			}
		case map[string]*ast.Package:
			for _, v0 := range t {
				for k, v := range v0.Files {
					buf.WriteString(fmt.Sprintf("\n// with %s: \n", k))
					err := printer.Fprint(buf, s.fset(), v)
					if err != nil {
						return "", err
					}
				}
			}
		default:
			buf.WriteByte('\n')
			err := printer.Fprint(buf, s.fset(), t)
			if err != nil {
				return "", err
			}
			buf.WriteByte('\n')
		}
	}

	return buf.String(), nil
}

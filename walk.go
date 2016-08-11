package gowalk

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// 源码历遍 用于找到自己想要的部分
type walk struct {
	fileSet *token.FileSet
	mode    parser.Mode
	root    *Node

	gopath []string
	pkgs   map[string]map[string]*ast.Package
}

func NewWalk(path string) *Node {
	w := &walk{
		fileSet: token.NewFileSet(),
		mode:    parser.ParseComments,
		pkgs:    map[string]map[string]*ast.Package{},
	}
	w.gopath = w.getPath()
	w.root = w.find(path)
	return w.root
}

// 打开引用包
//  在可能存在的目录下寻找
func (w *walk) open(path string) (pkg map[string]*ast.Package, first error) {
	if pkg, ok := w.pkgs[path]; ok {
		return pkg, nil
	}
	for _, v := range w.gopath {
		dir := filepath.Join(v, path)
		pkg, err := parser.ParseDir(w.fileSet, dir, func(fi os.FileInfo) bool {
			name := fi.Name()
			return !fi.IsDir() &&
				len(name) > 0 &&
				name[0] != '.' &&
				strings.HasSuffix(name, ".go") &&
				!strings.HasSuffix(name, "_test.go")
		}, w.mode)
		if err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		w.pkgs[path] = pkg
		return pkg, nil
	}
	return
}

func (w *walk) find(path string) *Node {
	pkg, err := w.open(path)
	if err != nil {
		return nil
	}
	return newNode("", w, []interface{}{nil, pkg})
}

// 获得 全部能引用包的路径
func (w *walk) getPath() []string {
	gopath := []string{}
	gopath = append(gopath, "./", "./vendor/")

	for _, v := range []string{"../", "../../"} {
		gopath = append(gopath, v, filepath.Join(v, "src"))
	}

	for _, v := range strings.Split(os.Getenv("GOPATH"), ";") {
		gopath = append(gopath, filepath.Join(v, "src"))
	}

	gopath = append(gopath, filepath.Join(os.Getenv("GOROOT"), "src"))

	for i := 0; i != len(gopath); {
		gopath[i] = filepath.Clean(gopath[i])
		fi, err := os.Stat(gopath[i])
		if err != nil || !fi.IsDir() {
			gopath = append(gopath[:i], gopath[i+1:]...)
			continue
		}
		i++
	}
	return gopath
}

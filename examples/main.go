package main

import (
	"go/ast"
	"go/token"

	"gopkg.in/ffmt.v1"
	walk "gopkg.in/walk.v2"
)

func main() {
	selector := walk.NewSelectorMust("wjs_api/models")

	ff := selector.Walk(walk.Package | walk.File)

	s0 := ff.Walk(walk.File|walk.GenDecl).Walk(walk.GenDecl, func(i interface{}) bool {
		return i.(*ast.GenDecl).Tok == token.TYPE
	})
	ffmt.Puts(s0.Slice(0, 10))

	s1 := ff.Walk(walk.File | walk.FuncDecl).Walk(walk.FuncDecl)
	ffmt.Puts(s1.Slice(0, 10))

	// src, err := s.Format()
	// if err != nil {
	// 	ffmt.Mark(err)
	// 	return
	// }
	// ffmt.Puts(src)

}

package gowalk

import (
	"fmt"
	"testing"

	"github.com/wzshiming/ffmt"
)

func check(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func TestA(t *testing.T) {
	f := NewWalk("github.com/wzshiming/gowalk")
	b := f.Child("walk").Child("fileSet")

	fmt.Println(b.Pos())
	//	fmt.Println(b.Src())
	//	fmt.Println(b.ChildList())
}

func TestB(t *testing.T) {
	f := NewWalk("github.com/wzshiming/gowalk")

	b := f.Child("walk").Child("save")
	//r := b.Return()
	//fmt.Println(r.Len())
	ffmt.P(b)
	//	fmt.Println(b.Src())
	//	fmt.Println(b.ChildList())
}

var c = ""

func TestC(t *testing.T) {
	f := NewWalk("github.com/wzshiming/gowalk")

	b := f.Child("TestC").Var("c")

	ffmt.P(b.Value())
	ffmt.P(b.Type().Value())
	//ffmt.P(r.Index(0).Value())
	//	fmt.Println(b.Src())
	//	fmt.Println(b.ChildList())
}

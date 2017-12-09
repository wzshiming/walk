package walk

import (
	"fmt"
	"testing"

	"gopkg.in/ffmt.v1"
)

func check(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func TestA(t *testing.T) {
	f := NewWalk("gopkg.in/walk.v1")
	b := f.Child("walk").Child("fileSet")

	fmt.Println(b.Pos())
	//	fmt.Println(b.Src())
	//	fmt.Println(b.ChildList())
}

func TestB(t *testing.T) {
	f := NewWalk("gopkg.in/walk.v1")

	b := f.Child("walk").Child("save")
	//r := b.Return()
	//fmt.Println(r.Len())
	ffmt.P(b)
	//	fmt.Println(b.Src())
	//	fmt.Println(b.ChildList())
}

var c = ""

func TestC(t *testing.T) {
	f := NewWalk("gopkg.in/walk.v1")

	b := f.Child("TestC").Var("c")

	ffmt.P(b.Value())
	ffmt.P(b.Type().Value())
	//ffmt.P(r.Index(0).Value())
	//	fmt.Println(b.Src())
	//	fmt.Println(b.ChildList())
}

func TestD(t *testing.T) {
	f := NewWalk("wjs_app/controllers")

	rui := f.Child("apim.ReqUserId")
	ffmt.Puts(rui.Name())
	ffmt.Puts(rui.ChildList())

	for _, v := range rui.ChildList() {
		b := rui.Child(v)

		t := b.Type()
		ffmt.P(b.Name(), rui.Child(t.Name()).Type().Value())
	}

	//ffmt.P(r.Index(0).Value())
	//	fmt.Println(b.Src())
	//	fmt.Println(b.ChildList())
}

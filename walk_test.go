package gowalk

import (
	"fmt"
	"testing"

	"github.com/wzshiming/ffmt"
)

var A = 1
var fff = ffmt.NewOptional(6, ffmt.StlyeP, ffmt.CanRowSpan)

func check(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
}

func TestA(t *testing.T) {
	f := NewWalk("github.com/wzshiming/gowalk")

	b := f.Child("Walk").Child("fileSet").Child()

	fmt.Println(b.Pos())
	fmt.Println(b.Src())
	fmt.Println(b.ChildList())
}

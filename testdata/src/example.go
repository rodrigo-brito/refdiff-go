// Some header to include offset in parser
// Another line to test parser
// One more...

package main

import (
	"fmt"
	"image"
	"strconv"

	"parser/testdata/src/lib"
)

type Printer interface {
	PrintString(value string)
	PrintInt(int)
}

type TypeAlias int

type Test struct {
	Name      string
	Year      int
	validator *lib.Validator
}

func myfunc(*image.Point, []float64) {}

func (t Test) Foo(bla int) int {
	bla = bla + 1
	return bla
}

func Foo(bla int) int {
	bla = bla + 1
	return bla
}

func Bar(bla int) int {
	bla = bla + 1
	return Foo(bla)
}

func (t *Test) PrintString(value string) {
	fmt.Println(value)
}

func (t *Test) PrintInt(value int) {
	content := strconv.Itoa(value)
	validator := new(lib.Validator)
	fmt.Println("[áç~!]" +
		"[áç~!]" +
		"[áç~!]" +
		"[áç~!]" +
		"[áç~!]")
	res := Bar(value)
	if t.validator.ValidNumber(value, res) && validator.ValidName("test") {
		t.PrintString(content)
	}
}

// Some header to include offset in parser
// Another line to test parser
// One more...

package testdata

import (
	"fmt"
	"image"
	"strconv"
)

type Printer interface {
	PrintString(value string)
	PrintInt(int)
}

type Writer func(string)

type (
	TypeAlias  int
	emptyWatch chan Test
)

type Test struct {
	Name      string
	Year      int
	validator *Validator
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
	validator := new(Validator)
	fmt.Println("[áç~!]") // test utf8 parse
	res := Bar(value)
	if t.validator.ValidNumber(value, res) && validator.ValidName("test") {
		t.PrintString(content)
	}
}

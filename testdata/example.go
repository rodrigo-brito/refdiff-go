package testdata

import (
	"fmt"
	"image"
)

type Printer interface {
	Print()
}

type Test struct {
	Name string
	Year int
	test *Test
}

func myfunc(i int, s string, err error, pt image.Point, x []float64) {}

func (t *Test) Print() {
	fmt.Println("test")
}

func Bar(bla int) int {
	bla = bla + 1
	return bla
}

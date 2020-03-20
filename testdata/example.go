package testdata

import "fmt"

type Test struct {
	Name string
	Year int
	test *Test
}

func (t *Test) Foo() {
	fmt.Println("test")
}

func Bar(bla int) int {
	bla = bla + 1
	return bla
}

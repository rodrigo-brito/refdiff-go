package main

import (
	"fmt"
	"image"
	"parser/testdata/src/lib"
	"strconv"
)

type Printer interface {
	PrintString(value string)
	PrintInt(int)
}

type Test struct {
	Name string
	Year int
	test *Test
}

func myfunc(*image.Point, []float64) {}

func (t *Test) PrintString(value string) {
	fmt.Println(value)
}

func (t *Test) PrintInt(value int) {
	content := strconv.Itoa(value)
	validator := new(lib.Validator)
	valid := validator.ValidName(content)
	if valid {
		t.PrintString(content)
	}
}

func Bar(bla int) int {
	bla = bla + 1
	return bla
}

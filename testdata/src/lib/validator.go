// Some header to include offset in parser
// Another line to test parser
// One more...

package lib

import "strings"

type Validator struct{}

func (Validator) ValidNumbers(m1 int, m2 ...int) bool {
	return true
}

func (v Validator) ValidNumber(n1, n2 int) bool {
	return true
}

func (v *Validator) ValidName(value string) bool {
	if len(strings.TrimSpace(value)) > 0 {
		return true
	}
	return false
}

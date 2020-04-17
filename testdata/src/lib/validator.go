package lib

import "strings"

type Validator struct{}

func (Validator) ValidName(value string) bool {
	if len(strings.TrimSpace(value)) > 0 {
		return true
	}
	return false
}

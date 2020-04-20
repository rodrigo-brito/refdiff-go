package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizePackage(t *testing.T) {
	tt := []struct {
		pkg      string
		base     string
		expected string
	}{
		{pkg: "src", base: "src", expected: "src"},
		{pkg: "fmt/Println", base: "src/lib", expected: "fmt/Println"},
		{pkg: "src/lib", base: "src/lib", expected: "src/lib"},
		{pkg: "src/lib/test", base: "src/lib", expected: "src/lib/test"},
		{pkg: "parser/testdata/src/lib/validator", base: "src/lib", expected: "src/lib/validator"},
		{pkg: "parser/testdata/src/lib", base: "src/lib", expected: "src/lib"},
		{pkg: "command-line-arguments", base: "src/lib", expected: "src/lib"},
	}

	for _, tc := range tt {
		t.Run(tc.pkg, func(t *testing.T) {
			res := normalizePackage(tc.pkg, tc.base)
			require.Equal(t, tc.expected, res)
		})
	}
}

func TestGetCallGraph(t *testing.T) {
	calls, err := GetCallGraph("/home/rodrigo/development/dep", "internal/...")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(calls)
}

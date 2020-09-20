package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractor_Extract(t *testing.T) {
	extractor, err := NewExtractor(".", "testdata/calls.go")
	require.Nil(t, err)

	nodes := extractor.Extract()
	require.Len(t, nodes, 14)

	require.Equal(t, FileType, nodes[0].Type)
	require.Equal(t, "calls.go", nodes[0].Name)

	require.Equal(t, InterfaceType, nodes[1].Type)
	require.Equal(t, "Printer", nodes[1].Name)

	require.Equal(t, FunctionType, nodes[2].Type)
	require.Equal(t, "PrintString", nodes[2].Name)

	require.Equal(t, FunctionType, nodes[3].Type)
	require.Equal(t, "PrintInt", nodes[3].Name)

	require.Equal(t, PrimitiveType, nodes[4].Type)
	require.Equal(t, "Writer", nodes[4].Name)

	require.Equal(t, PrimitiveType, nodes[5].Type)
	require.Equal(t, "TypeAlias", nodes[5].Name)
}

func TestExtractor_getNamespace(t *testing.T) {
	tt := []struct {
		file   string
		sufix  string
		result string
	}{
		{"a/b/c.go", "MyStruct", "a/b/MyStruct."},
		{"a/b/c.go", "", "a/b/"},
		{"c.go", "", ""},
		{"c.go", "Context", "Context."},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%s-%s", tc.file, tc.sufix), func(t *testing.T) {
			extractor := Extractor{rootFolder: ".", fileName: tc.file}
			result := extractor.getNamespace(tc.sufix)
			require.Equal(t, tc.result, result)
		})
	}
}

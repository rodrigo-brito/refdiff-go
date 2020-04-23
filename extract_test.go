package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractor_Extract(t *testing.T) {
	extractor, err := NewExtractor(".", "testdata/src/lib/validator.go")
	require.Nil(t, err)

	nodes := extractor.Extract()
	require.Len(t, nodes, 9)

	require.Equal(t, FileType, nodes[0].Type)
	require.Equal(t, "example.go", nodes[0].Name)

	require.Equal(t, InterfaceType, nodes[1].Type)
	require.Equal(t, "Printer", nodes[1].Name)

	require.Equal(t, StructType, NodeType(nodes[2].Type))
	require.Equal(t, "Test", nodes[2].Name)

	require.Equal(t, FunctionType, nodes[3].Type)
	require.Equal(t, "myfunc", nodes[3].Name)

	require.Equal(t, FunctionType, nodes[4].Type)
	require.Equal(t, "Print", nodes[4].Name)

	require.Equal(t, FunctionType, nodes[5].Type)
	require.Equal(t, "Bar", nodes[5].Name)
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

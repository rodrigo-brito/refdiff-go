package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractor_Extract(t *testing.T) {
	extractor, err := NewExtractor("testdata/example.go")
	require.Nil(t, err)

	nodes := extractor.Extract()
	require.Len(t, nodes, 5)

	require.Equal(t, InterfaceType, nodes[0].Type)
	require.Equal(t, "Printer", nodes[0].Name)

	require.Equal(t, StructType, nodes[1].Type)
	require.Equal(t, "Test", nodes[1].Name)

	require.Equal(t, FunctionType, nodes[2].Type)
	require.Equal(t, "myfunc", nodes[2].Name)

	require.Equal(t, FunctionType, nodes[3].Type)
	require.Equal(t, "Test.Print", nodes[3].Name)

	require.Equal(t, FunctionType, nodes[4].Type)
	require.Equal(t, "Bar", nodes[4].Name)
}

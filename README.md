# go-ast-parser

![tests](https://github.com/rodrigo-brito/go-ast-parser/workflows/tests/badge.svg)
[![Go Report](https://goreportcard.com/badge/github.com/rodrigo-brito/go-ast-parser)](https://goreportcard.com/report/github.com/rodrigo-brito/go-ast-parser)
[![GoDoc](https://godoc.org/github.com/rodrigo-brito/go-ast-parser?status.svg)](https://godoc.org/github.com/rodrigo-brito/go-ast-parser)

Parser Go File to a list of structs, interfaces and functions in JSON format.

## Usage

```bash
$ go build
$ ./parser -file testdata/example.go
```

## Example of output

```json
[
  {
    "type": "File",
    "start": 1,
    "end": 27,
    "name": "example.go",
    "namespace": "testdata/",
    "parent": null,
    "tokens": [
      "1-9",
      "9-18",
      "17-19",
      ...
      "315-317"
    ]
  },
  {
    "type": "Interface",
    "start": 8,
    "end": 10,
    "name": "Printer",
    "namespace": "",
    "parent": "testdata/example.go"
  },
  {
    "type": "Struct",
    "start": 12,
    "end": 16,
    "name": "Test",
    "namespace": "testdata/",
    "parent": "testdata/example.go"
  },
  {
    "type": "Function",
    "start": 18,
    "end": 18,
    "name": "myfunc",
    "namespace": "testdata/",
    "parent": "testdata/example.go",
    "parameter_names": ["i", "s", "err", "pt", "x"],
    "parameter_types": ["int", "string", "error", "image.Point", "[]float64"]
  },
  {
    "type": "Function",
    "start": 20,
    "end": 22,
    "name": "Print",
    "namespace": "testdata/Test.",
    "parent": "testdata/Test"
  },
  {
    "type": "Function",
    "start": 24,
    "end": 27,
    "name": "Bar",
    "namespace": "testdata/",
    "parent": "testdata/example.go",
    "parameter_names": ["bla"],
    "parameter_types": ["int"]
  }
]
```

## License

Distributed under [MIT License](LICENSE)
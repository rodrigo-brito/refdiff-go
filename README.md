# go-ast-parser

Parser Go File in a list of structs, interfaces and functions

## Usage

```bash
$ go build
$ ./parser  -file testdata/example.go
```

## Example of output

```json
[
  {
    "start": 14,
    "end": 14,
    "tokens": [
      "{",
      "}"
    ],
    "name": "myfunc(int,string,error,image.Point,[]float64)",
    "type": "METHOD",
    "namespace": "testdata",
    "parameter_names": [
      "i",
      "s",
      "err",
      "pt",
      "x"
    ],
    "parameter_types": [
      "int",
      "string",
      "error",
      "image.Point",
      "[]float64"
    ]
  },
  {
    "start": 16,
    "end": 18,
    "tokens": [
      "{",
      "fmt",
      ".",
      "Println",
      "(",
      "\"test\"",
      ")",
      "}"
    ],
    "name": "(Test) Foo()",
    "type": "METHOD",
    "namespace": "testdata",
    "parameter_names": null,
    "parameter_types": null
  },
]
```

## License

Distributed under [MIT License](LICENSE)
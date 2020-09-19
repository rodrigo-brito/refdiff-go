# refdiff-go

![tests](https://github.com/rodrigo-brito/go-ast-parser/workflows/tests/badge.svg)
[![Go Report](https://goreportcard.com/badge/github.com/rodrigo-brito/go-ast-parser)](https://goreportcard.com/report/github.com/rodrigo-brito/go-ast-parser)
[![GoDoc](https://godoc.org/github.com/rodrigo-brito/go-ast-parser?status.svg)](https://godoc.org/github.com/rodrigo-brito/go-ast-parser)

Support for Go programming language in RefDiff

## Enviroment Variables

```bash
REFDIFF_GO_PARSER=/usr/bin/refdiff-go-parser # customize parser location
```

## Go AST Parser

```bash
$ go build -o refdiff-go-parser parser/main.go
$ ./refdiff-go-parser -file parser/testdata/example.go
```

## Installation

- Follow installation instructions for [RefDiff Core](https://github.com/aserg-ufmg/RefDiff)
- Copy `refdiff-plugin` to your project envrironment
- Download the Go parser in [release page](https://github.com/rodrigo-brito/refdiff-go/releases) and include in your PATH (optional)

## Example of usage

```java
package refdiff.examples;

import java.io.File;

import refdiff.core.RefDiff;
import refdiff.core.diff.CstDiff;
import refdiff.core.diff.Relationship;
import refdiff.parsers.go.GoPlugin;

public class RefDiffExampleGolang {
	public static void main(String[] args) throws Exception {
		runExample();
	}

	private static void runExample() throws Exception {
		// This is a temp folder to clone or checkout git repositories.
		File tempFolder = new File("temp");

		// Creates a RefDiff instance configured with the Go plugin.
		try (GoPlugin goPlugin = new GoPlugin(tempFolder)) {
			RefDiff refDiffGo = new RefDiff(goPlugin);

			File repo = refDiffGo.cloneGitRepository(
				new File(tempFolder, "go-refactoring-example.git"),
				"https://github.com/rodrigo-brito/go-refactoring-example.git");

			CstDiff diffForCommit = refDiffGo.computeDiffForCommit(repo, "2a2bc542e3b9ea549936556c08ebeaaf7e98adbc");
			printRefactorings("Refactorings found in go-refactoring-example 2a2bc542e3b9ea549936556c08ebeaaf7e98adbc", diffForCommit);
		}
	}

	private static void printRefactorings(String headLine, CstDiff diff) {
		System.out.println(headLine);
		for (Relationship rel : diff.getRefactoringRelationships()) {
			System.out.println(rel.getStandardDescription());
		}
	}
}
```

## License

Distributed under [MIT License](LICENSE)

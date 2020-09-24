//Copyright 2020 Google LLC

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"text/template"
)

// Contains the function information.
// Line keeps the definition line of the function in the source code.
type funcCoverBlock struct {
	Name string
	Line int32
}

// Parses given source code and returns a funcCoverBlock array.
func saveFuncs(src string, content []byte) ([]funcCoverBlock, error) {
	fset := token.NewFileSet()

	parsedFile, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var funcBlocks []funcCoverBlock

	// Finds function declerations to instrument and saves them to funcCover
	for _, decl := range parsedFile.Decls {
		switch t := decl.(type) {
		// Function decleration is found.
		case *ast.FuncDecl:
			if t.Body == nil {
				continue
			}
			funcBlocks = append(funcBlocks, funcCoverBlock{
				Name: t.Name.Name,
				Line: int32(fset.Position(t.Pos()).Line),
			})
		}
	}

	return funcBlocks, nil
}

// Returns the source code representation of a AST file
func astToByte(fset *token.FileSet, f *ast.File) []byte {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	return buf.Bytes()
}

// Inserts necessary set instrucions for instrumentation to function definitions.
// Writes the instrumented output using w.
func insertInstructions(w io.Writer, content []byte, coverVar string) (bool, error) {
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return false, err
	}

	var contentLength = len(content)
	var events []int
	var mainLbrace = -1

	// Iterates over functions to find the positions to insert instructions.
	// Saves the positions to array events.
	// Finds the position of main function if exists.
	// This will be used to insert a defer call to an exit hook for coverage report on exit.
	for _, decl := range parsedFile.Decls {
		switch t := decl.(type) {
		// Function decleration is found.
		case *ast.FuncDecl:
			if t.Body == nil {
				continue
			}
			events = append(events, int(t.Body.Lbrace))
			if t.Name.Name == "main" {
				mainLbrace = int(t.Body.Lbrace) - 1
			}
		}
	}

	// Writes the instrumented code using w io.Writer.
	// Insert set instructions to the functions.
	// f() {
	// 	coverVar.Executed[funcNumber] = true;
	// 	...
	// }
	// Also inserts "defer (*FuncCoverExitHook)()" to the beginning of main().
	// Initially hook just points to an empty function but the handler can override it.
	// func main {
	// 	defer (*FuncCoverExitHook)()
	//	...
	// }
	eventIndex := 0
	for i := 0; i < contentLength; i++ {
		if eventIndex < len(events) && i == events[eventIndex] {
			fmt.Fprintf(w, "\n\t%s.Executed[%v] = true;", coverVar, eventIndex)
			eventIndex++
		}
		fmt.Fprintf(w, "%s", string(content[i:i+1]))
		if i == mainLbrace {
			fmt.Fprintf(w, "\n\tdefer (*FuncCoverExitHook)()\n")
		}
	}

	return mainLbrace != -1, nil
}

// Writes the declaration of cover variables using go templates.
// If source has a main function, defines FuncCoverExitHook function pointer.
// Initializes FuncCoverExitHook with an empty function.
// Embedded report libraries can implement an init() function to use FuncCoverExitHook.
func declCover(w io.Writer, coverVar, source string, hasMain bool, funcCover []funcCoverBlock) {

	funcTemplate, err := template.New("cover variables").Parse(declTmpl)

	if err != nil {
		panic(err)
	}

	var declParams = struct {
		CoverVar   string
		SourceName string
		HasMain    bool
		FuncBlocks []funcCoverBlock
	}{coverVar, source, hasMain, funcCover}

	err = funcTemplate.Execute(w, declParams)

	if err != nil {
		panic(err)
	}
}

var declTmpl = `
var {{.CoverVar}} = struct {
	SourcePath		string
	FuncNames		[]string
	FuncLines		[]int32
	Executed		[]bool
} {
	SourcePath: "{{.SourceName}}",
	FuncNames: []string{ {{range .FuncBlocks}}
		"{{.Name}}",{{end}}
	},
	FuncLines: []int32{ {{range .FuncBlocks}}
		{{.Line}},{{end}}
	},
	Executed: []bool{ {{range .FuncBlocks}}
		false,{{end}}
	},
}

{{ if eq .HasMain true }}
var FuncCoverExitHookEmptyFunc = func() {}
var FuncCoverExitHook *func() = &FuncCoverExitHookEmptyFunc{{end}}`

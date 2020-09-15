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

// FuncCoverBlock contains tha name and line of a function
// Line contains the line of the definition in the source code
type FuncCoverBlock struct {
	Name string
	Line int32
}

// SaveFuncs parses given source code and returns a FuncCover instance
func SaveFuncs(src string, content []byte) ([]FuncCoverBlock, error) {

	fset := token.NewFileSet()

	parsedFile, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var funcBlocks []FuncCoverBlock

	// Find function declerations to instrument and save them to funcCover
	for _, decl := range parsedFile.Decls {
		switch t := decl.(type) {
		// Function Decleration
		case *ast.FuncDecl:
			funcBlocks = append(funcBlocks, FuncCoverBlock{
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

// InsertInstructions writes necessary set instrucions for instrumentation to function definitions
func InsertInstructions(w io.Writer, content []byte, coverVar string) (bool, error) {

	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return false, err
	}

	var contentLength = len(content)
	var events []int
	var mainLbrace = -1

	// Iterates over functions to find the positions to insert instructions
	// Saves the positions to array events
	// Finds the position of main function if exists
	// This will be used to insert a defer call, so that coverage can be collected just before exit
	// Positions are changed due to imports but saved information in funcCover will be the same with the
	// initial source code
	for _, decl := range parsedFile.Decls {
		switch t := decl.(type) {
		// Function Decleration
		case *ast.FuncDecl:
			events = append(events, int(t.Body.Lbrace))
			if t.Name.Name == "main" {
				mainLbrace = int(t.Body.Lbrace) - 1
			}
		}
	}

	// Writes the instrumented code using w io.Writer
	// Insert set instructions to the functions
	// f() {
	// 	coverVar.Counts[funcNumber] = 1;
	// 	...
	// }
	// Also inserts defer LastCallForFunccoverReport() to the beginning of main()
	// Initially this is just an empty function but handler can override it
	// func main {
	// 	defer LastCallForFunccoverReport()
	//	...
	// }
	eventIndex := 0
	for i := 0; i < contentLength; i++ {
		if eventIndex < len(events) && i == events[eventIndex] {
			fmt.Fprintf(w, "\n\t%s.Flags[%v] = true;", coverVar, eventIndex)
			eventIndex++
		}
		fmt.Fprintf(w, "%s", string(content[i]))
		if i == mainLbrace {
			fmt.Fprintf(w, "\n\tdefer (*LastCallForFunccoverReport)()\n")
		}
	}

	return mainLbrace != -1, nil
}

// declCover writes the declaration of cover variable to the end of the given source file writer using go templates
// If source has a main function, defines LastCallForFunccoverReport function variable and assign an empty
// function to it
// Embedded report libraries can implement an init() function that assigns LastCallForFunccoverReport
// a different function referance
func declCover(w io.Writer, coverVar, source string, hasMain bool, funcCover []FuncCoverBlock) {

	funcTemplate, err := template.New("cover variables").Parse(declTmpl)

	if err != nil {
		panic(err)
	}

	var declParams = struct {
		CoverVar   string
		SourceName string
		HasMain    bool
		FuncBlocks []FuncCoverBlock
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
	Flags			[]bool
} {
	SourcePath: "{{.SourceName}}",
	FuncNames: []string{ {{range .FuncBlocks}}
		"{{.Name}}",{{end}}
	},
	FuncLines: []int32{ {{range .FuncBlocks}}
		{{.Line}},{{end}}
	},
	Flags: []bool{ {{range .FuncBlocks}}
		false,{{end}}
	},
}

{{ if eq .HasMain true }}
var EmptyVoidFunctionThisNameIsLongToAvoidCollusions082 = func() {}
var LastCallForFunccoverReport *func() = &EmptyVoidFunctionThisNameIsLongToAvoidCollusions082{{end}}`

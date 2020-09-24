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

// This code implements source file instrumentation function
package main

import (
	"bytes"
	"fmt"
	"go/token"
	"io/ioutil"
)

// funccoverInstrumenter struct have the data necessary for instrumentation.
type funccoverInstrumenter struct {
	fset     *token.FileSet
	content  []byte
	coverVar string
	outPath  string
	srcName  string
}

// Saves given file's content to funccoverInstrumenter instance.
func (h *funccoverInstrumenter) saveFile(src string) error {
	if h.fset == nil {
		h.fset = token.NewFileSet()
	}

	content, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	h.content = content

	return nil
}

// Instruments the source code content saved in h.
func (h *funccoverInstrumenter) instrument() ([]byte, error) {
	var funcCover = []funcCoverBlock{}

	// Saves the function data to funcCover
	funcCover, err := saveFuncs(h.srcName, h.content)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	// Inserts necessary instructions to the functions
	hasMain, err := insertInstructions(buf, h.content, h.coverVar)

	if err != nil {
		return nil, err
	}

	// Appends necessary variable definitions
	declCover(buf, h.coverVar, h.srcName, hasMain, funcCover)

	return buf.Bytes(), nil
}

// Writes the instrumented content to h.outPath
func (h *funccoverInstrumenter) writeInstrumented(instrumented []byte) error {
	if err := ioutil.WriteFile(h.outPath, instrumented, 0666); err != nil {
		return fmt.Errorf("Instrumentation failed: %v", err)
	}
	return nil
}

// Instruments the file given and writes it to outPath
func instrumentForFunctionCoverage(srcPath, srcName, coverVar, outPath string) error {
	var instrumenter = funccoverInstrumenter{
		coverVar: coverVar,
		outPath:  outPath,
		srcName:  srcName,
	}

	err := instrumenter.saveFile(srcPath)
	if err != nil {
		return err
	}

	instrumented, err := instrumenter.instrument()
	if err != nil {
		return err
	}

	err = instrumenter.writeInstrumented(instrumented)
	if err != nil {
		return err
	}

	return nil
}

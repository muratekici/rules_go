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

package coverdata

// FunctionCover contains the coverage data needed
type FunctionCover struct {
	SourcePaths   []string
	FunctionNames []string
	FunctionLines []int32
	Flags         []*bool
}

// FuncCover keeps the coverage data
// It is exported so that other packages can use it to report coverage
var FuncCover FunctionCover

// RegisterFileFuncCover eegisters functions to exported variable FuncCover
func RegisterFileFuncCover(SourcePath string, FuncNames []string, FuncLines []int32, Flags []bool) {
	for i, funcName := range FuncNames {
		FuncCover.SourcePaths = append(FuncCover.SourcePaths, SourcePath)
		FuncCover.FunctionNames = append(FuncCover.FunctionNames, funcName)
		FuncCover.FunctionLines = append(FuncCover.FunctionLines, FuncLines[i])
		FuncCover.Flags = append(FuncCover.Flags, &(Flags[i]))
	}
}

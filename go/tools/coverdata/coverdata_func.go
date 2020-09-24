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

// FuncCover contains the coverage data needed
type FuncCover struct {
	SourcePaths []string
	FuncNames   []string
	FuncLines   []int32
	Executed    []*bool
}

// FuncCoverData keeps the coverage data in runtime
// Instrumented packages registers their coverage data to FuncCoverData
var FuncCoverData FuncCover

// RegisterFileFunc eegisters functions to exported variable FuncCoverData
func RegisterFileFunc(SourcePath string, FuncNames []string, FuncLines []int32, Executed []bool) {
	for i, funcName := range FuncNames {
		FuncCoverData.SourcePaths = append(FuncCoverData.SourcePaths, SourcePath)
		FuncCoverData.FuncNames = append(FuncCoverData.FuncNames, funcName)
		FuncCoverData.FuncLines = append(FuncCoverData.FuncLines, FuncLines[i])
		FuncCoverData.Executed = append(FuncCoverData.Executed, &(Executed[i]))
	}
}

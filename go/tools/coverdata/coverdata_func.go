package coverdata

type FunctionCover struct {
	SourcePaths   []string
	FunctionNames []string
	FunctionLines []int32
	Counts        []*bool
}

var FuncCover FunctionCover

// Registers functions  to FuncCover
func RegisterFileFuncCover(SourcePath string, FuncNames []string, FuncLines []int32, Counts []bool) {
	for i, funcName := range FuncNames {
		FuncCover.SourcePaths = append(FuncCover.SourcePaths, SourcePath)
		FuncCover.FunctionNames = append(FuncCover.FunctionNames, funcName)
		FuncCover.FunctionLines = append(FuncCover.FunctionLines, FuncLines[i])
		FuncCover.Counts = append(FuncCover.Counts, &(Counts[i]))
	}
}

package validate

import (
	"fmt"
	"os"

	"bitbucket.org/sealuzh/gopper/data/input"
	"bitbucket.org/sealuzh/gopper/util"
)

func AnalysisFunc(sps input.SubPrograms, in input.Config) bool {
	if len(sps.Occurrences[input.SpAnalyse]) == 0 {
		return false
	}

	funcName := in.Analyse.Name

	if funcName == "" {
		fmt.Printf("Sup-program %s specified, but no information about function in config file.\n", input.SpAnalyse)
		return false
	}

	valid := false
	for _, f := range input.AnalyseFuncs {
		if f == funcName {
			valid = true
			break
		}
	}

	if !valid {
		fmt.Printf("Analysis function '%s' invalid. Must be one of %v\n", funcName, input.AnalyseFuncs)
		return false
	}

	valid = true
	// check if script is an r file
	path, err := input.StringParam(in.Analyse, 0)
	if err != nil {
		fmt.Printf("Analysis function parameter can not be retrieved: %v\n", err)
		return false
	}
	s, err := os.Stat(util.AbsolutePath(path))
	if err != nil {
		fmt.Printf("Analysis function script not accessible: %v\n", err)
		valid = false
	} else {
		if s.IsDir() {
			fmt.Printf("Analysis function script is a directory\n")
			valid = false
		}
	}

	return valid
}

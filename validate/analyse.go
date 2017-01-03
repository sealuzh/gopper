package validate

import (
	"fmt"

	"bitbucket.org/sealuzh/gopper/data/input"
)

func AnalysisFunc(sps input.SubPrograms, in input.Config) bool {
	if len(sps.Occurrences[input.SpAnalyse]) == 0 {
		return true
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

	return true
}

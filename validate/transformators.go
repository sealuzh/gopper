package validate

import (
	"fmt"

	"bitbucket.org/sealuzh/gopper/data/input"
)

func Transformators(sps input.SubPrograms, in input.Config) bool {
	valid := true

	lSpTrans := len(sps.Occurrences[input.SpFilter])
	if lSpTrans == 0 {
		return true
	}

	// validate Transform
	if len(in.Transform) != 0 {
		for _, t := range in.Transform {
			var contains bool
			for _, tf := range input.TransFuncs {
				if t.Name == tf {
					contains = true
					break
				}
			}

			if !contains {
				fmt.Printf("Invalid transformer function '%s'. Must be one of %v.\n", t, input.TransFuncs)
				valid = false
				break
			}
		}
	} else {
		valid = false
	}

	return valid
}

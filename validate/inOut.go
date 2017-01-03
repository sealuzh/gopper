package validate

import (
	"fmt"

	"github.com/sealuzh/gopper/data/input"
)

func InOut(sps input.SubPrograms, in input.Config) bool {
	valid := true
	lIn := len(in.In)
	lOut := len(in.Out.TestResults)
	// input files
	if lIn == 0 {
		fmt.Printf("At least one input file is mandatory (was %d)\n", lIn)
		valid = false
	} else {
		if len(sps.Occurrences[input.SpMerge]) > 0 {
			valid = lIn >= 2 && lOut == 1
			if !valid {
				fmt.Printf("Merge requires exactly one output file (was %d) and at least 2 input files (was %d)\n", lOut, lIn)
			}
		} else {
			// Either no out files specified or number of in and out files is the same
			valid = lOut == 0 || lIn == lOut
			if !valid {
				fmt.Printf("Output file numbers must be 0, or input and output file number must be equivalent (was in=%d, out=%d)", lIn, lOut)
			}
		}
	}

	return valid
}

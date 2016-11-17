package validate

import (
	"flag"
	"fmt"
	"os"

	"bitbucket.org/sealuzh/gopper/data/input"
)

func Arguments(sps input.SubPrograms, in input.Config) {
	var invalid bool
	const (
		argMissing = "Argument missing: %s\n"
		argInvalid = "Argument invalid: %s\n"
	)

	// validate sub-program usage
	if sps.Count == 0 {
		fmt.Printf(argMissing, "sub programm")
		invalid = true
	} else {
		// check if all suprogramms are allowed
		allowed := sps.Count == (len(sps.Merge) + len(sps.Plot) + len(sps.Transform) + len(sps.Analyse))
		if !allowed {
			fmt.Printf("Sub-programs specified invalid. Every sub-program must be from: %v\n", input.SubProgs)
			invalid = true
		}
	}

	// validate config file, in is mandatory
	if len(in.In) == 0 {
		fmt.Printf(argMissing, "config file")
		invalid = true
	}

	invalid = invalid || !InOut(sps, in)
	invalid = invalid || !Transformators(sps, in)
	invalid = invalid || !Plot(sps, in)
	invalid = invalid || !AnalysisFunc(sps, in)

	if invalid {
		fmt.Println()
		flag.Usage()
		os.Exit(-1)
	}
}

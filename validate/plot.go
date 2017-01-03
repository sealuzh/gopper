package validate

import "github.com/sealuzh/gopper/data/input"

func Plot(sps input.SubPrograms, in input.Config) bool {
	l := len(sps.Occurrences[input.SpPlot])
	if l == 0 {
		return true
	}

	if in.Out.Plot != "" {
		return true
	}
	return false
}

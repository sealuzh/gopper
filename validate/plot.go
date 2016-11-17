package validate

import "bitbucket.org/sealuzh/gopper/data/input"

func Plot(sps input.SubPrograms, in input.Config) bool {
	l := len(sps.Plot)
	if l == 0 {
		return true
	}

	if in.Plot != "" {
		return true
	}
	return false
}

package save

import (
	"encoding/csv"
	"fmt"
	"os"

	"bitbucket.org/sealuzh/gopper/data"
	"bitbucket.org/sealuzh/gopper/util"
)

func TestResults(stageNr int, d []data.TestResults, outPaths []string) {
	lo := len(outPaths)
	ld := len(d)
	if len(outPaths) == 0 {
		return
	}

	if lo != ld {
		// should have been checked by input validators
		panic(fmt.Sprintf("Number of results (%d) and number of output files (%d) must be the same", ld, lo))
	}

	for i, r := range d {
		outPath := util.AbsolutePath(outPaths[i])
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("ERROR - Could not open output file '%v': %v", outPath, err)
		} else {
			defer f.Close()
			w := csv.NewWriter(f)
			w.Comma = comma
			defer w.Flush()
			w.Write(r.Heading())
			w.Flush()
			for _, n := range r.TestNames() {
				rs, ok := r.Get(n)
				if !ok {
					// should not be the case
					fmt.Printf("ERROR - Could not retrieve test with name '%s'", n)
					continue
				}
				for _, c := range rs.Commits() {
					ers, ok := rs.ExecutionResult(c)
					if !ok {
						panic(fmt.Sprintf("Inconsistent test result: %s @ %s", r, c))
					}
					for _, r := range ers {
						w.Write(r.AsStringArray())
					}
				}
				w.Flush()
			}
		}
	}
}

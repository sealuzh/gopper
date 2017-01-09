package save

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/sealuzh/gopper/data"
	"github.com/sealuzh/gopper/util"
)

func ChangePoints(stageNr int, cps []data.ChangePoints, paths []string) {
	lcps := len(cps)
	lpaths := len(paths)
	if lcps != lpaths {
		fmt.Printf("ERROR - length of change points (%d) and paths (%d) not equal\n", lcps, lpaths)
		return
	}

	for i, cp := range cps {
		outPath := util.AbsolutePath(paths[i])
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("ERROR - Could not open output file '%v': %v\n", outPath, err)
		} else {
			copy := cp.Copy()
			sort.Sort(sort.Reverse(copy))
			e := json.NewEncoder(f)
			e.SetIndent("", "    ") // indentation is 4 spaces
			e.Encode(copy)
			f.Close()
		}
	}
}

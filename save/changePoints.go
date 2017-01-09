package save

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sealuzh/gopper/data"
	"github.com/sealuzh/gopper/util"
)

func ChangePoints(stageNr int, trs []data.TestResults, cps []data.ChangePoints, paths []string) {
	lcps := len(cps)
	ltrs := len(trs)
	lpaths := len(paths)
	if lcps != lpaths || lcps != ltrs || lpaths != ltrs {
		fmt.Printf("ERROR - length of test results (%d) and change points (%d) and paths (%d) not equal\n", ltrs, lcps, lpaths)
		return
	}

	for i, cp := range cps {
		changePoints := cp.All()
		cpsAgg := make(map[string]*changePointTypes, len(cps))
		for _, c := range changePoints {
			commit := c.Commit()
			testNames := c.TestNames()
			testNamesLen := len(testNames)
			cpAgg, ok := cpsAgg[commit]
			if !ok {
				cpAgg = &changePointTypes{}
				cpsAgg[commit] = cpAgg
			}
			cpt := c.Type()
			if cpt.IsImprovement() {
				cpAgg.Im = testNamesLen
			} else if cpt.IsRegression() {
				cpAgg.Reg = testNamesLen
			}
		}

		// save json file
		saveJson(paths[i], cp)

		// save csv
		saveCsv(paths[i], commitOrder(trs[i]), cpsAgg)
	}
}

func commitOrder(trs data.TestResults) []string {
	ret := make([]string, 0)
	for tr := range trs.All() {
		commits := tr.Commits()
		if len(ret) < len(commits) {
			ret = commits
		}
	}
	return ret
}

func saveJson(path string, cp data.ChangePoints) {
	op := util.AbsolutePath(outPath(path, ".json"))
	f, err := os.Create(op)
	if err != nil {
		fmt.Printf("ERROR - Could not open output file '%v': %v\n", op, err)
	} else {
		copy := cp.Copy()
		sort.Sort(sort.Reverse(copy))
		e := json.NewEncoder(f)
		e.SetIndent("", "    ") // indentation is 4 spaces
		e.Encode(copy)
		f.Close()
	}
}

func saveCsv(path string, commits []string, cps map[string]*changePointTypes) {
	op := util.AbsolutePath(outPath(path, ".csv"))
	f, err := os.Create(op)
	if err != nil {
		fmt.Printf("ERROR - Could not open output file '%v': %v\n", op, err)
	} else {
		defer f.Close()

		w := csv.NewWriter(f)
		w.Comma = comma
		defer w.Flush()
		w.Write([]string{"Commit", "Run"})
		w.Flush()
		for _, c := range commits {
			cpt, ok := cps[c]
			if !ok {
				w.Write([]string{c, "0 / 0 "})
			} else {
				w.Write([]string{c, fmt.Sprintf("%d / %d", cpt.Im, cpt.Reg)})
			}
			w.Flush()
		}
	}
}

func outPath(path string, suffix string) string {
	if strings.HasSuffix(path, suffix) {
		return path
	}
	return path + suffix
}

type changePointTypes struct {
	Im  int
	Reg int
}

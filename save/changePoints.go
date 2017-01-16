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
		cpsAgg := make(map[string]data.ChangePoints, len(cps))
		for _, c := range changePoints {
			commit := c.Commit()
			cpsc := cp.At(commit)
			cpsAgg[commit] = cpsc
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

func saveCsv(path string, commits []string, cps map[string]data.ChangePoints) {
	op := util.AbsolutePath(outPath(path, ".csv"))
	f, err := os.Create(op)
	if err != nil {
		fmt.Printf("ERROR - Could not open output file '%v': %v\n", op, err)
	} else {
		defer f.Close()

		w := csv.NewWriter(f)
		w.Comma = comma
		defer w.Flush()

		cpTypes := data.AllChangePointTypes()
		columns := len(cpTypes) + 1
		// write csv heading line
		heading := make([]string, columns)
		heading[0] = "Commit"
		for i := 1; i < columns; i++ {
			heading[i] = cpTypes[i-1].String()
		}
		w.Write(heading)
		w.Flush()

		// write csv content
		for _, c := range commits {
			cpt, ok := cps[c]
			if !ok {
				w.Write(emptyLine(c, cpTypes))
			} else {
				w.Write(nonEmptyLine(c, cpTypes, cpt))
			}
			w.Flush()
		}
	}
}

func emptyLine(commit string, cpTypes []data.ChangePointType) []string {
	l := len(cpTypes) + 1
	line := make([]string, l)
	line[0] = commit
	for i := 1; i < l; i++ {
		line[i] = "0"
	}
	return line
}

func nonEmptyLine(commit string, cpTypes []data.ChangePointType, cps data.ChangePoints) []string {
	l := len(cpTypes) + 1
	line := make([]string, l)
	line[0] = commit

	for i, t := range cpTypes {
		k := i + 1
		cp, ok := cps.Get(commit, t)
		if ok {
			line[k] = fmt.Sprintf("%d", len(cp.TestNames()))
		} else {
			line[k] = "0"
		}
	}

	return line
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

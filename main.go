package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"bitbucket.org/sealuzh/gopper/analyse"
	"bitbucket.org/sealuzh/gopper/data"
	"bitbucket.org/sealuzh/gopper/data/input"
	"bitbucket.org/sealuzh/gopper/save"
	"bitbucket.org/sealuzh/gopper/transform/changepoints"
	"bitbucket.org/sealuzh/gopper/transform/filter"
	"bitbucket.org/sealuzh/gopper/transform/testresults"
	"bitbucket.org/sealuzh/gopper/util"
	"bitbucket.org/sealuzh/gopper/validate"
)

const comma = ';'

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	sps, config := parseArguments()
	validate.Arguments(sps, config)

	ctx := context.Background()
	ctx, f := context.WithCancel(ctx)
	defer f()

	var outTr []data.TestResults
	var outCp []data.ChangePoints

	// read in data
	var ins []data.TestResults = make([]data.TestResults, len(config.In))
	for i, in := range config.In {
		in := util.AbsolutePath(in)
		r, err := data.TestResultsFromFile(in)
		if err != nil {
			fmt.Printf("ERROR - could not read/parse file '%s': %v\n", in, err)
			return
		}
		ins[i] = r
	}

	// execute sub-programs
	outTr = ins
	startTime := time.Now()
	for i, sp := range sps.List {
		spUpper := strings.ToUpper(sp)
		stageNumber := i + 1
		fmt.Printf("# %d - %s: start stage\n", stageNumber, spUpper)
		stageStart := time.Now()
		// sequentially compute stages
		switch sp {
		case input.SpSave:
			handleSave(ctx, i, outTr, outCp, config)
		case input.SpPlot:
			handlePlot(ctx, i, outTr, config)
		case input.SpTRsToCPs:
			outCp = handleTRsToCPs(ctx, i, outTr, config)
		case input.SpMerge:
			outTr = []data.TestResults{data.Merge(ctx, outTr)}
		case input.SpRmDupTns:
			outCp = handleRmDupTns(ctx, i, outCp)
		case input.SpAnalyse:
			// only supports a single analyse function
			anFunc := analysisFuncFromIn(config)
			outTr = siso(ctx, sp, outTr, config, anFunc)
		case input.SpFilter:
			outTr = siso(ctx, sp, outTr, config, nil)
		default:
			panic(fmt.Sprintf("ERROR - Unknown Sub-Program '%v'\n", sp))
		}
		fmt.Printf("# %d - %s: finished stage in %v\n", stageNumber, spUpper, time.Since(stageStart))
	}
	fmt.Printf("# Total execution time: %v\n", time.Since(startTime))
}

func handleSave(ctx context.Context, stageNr int, trs []data.TestResults, cps []data.ChangePoints, config input.Config) {
	if cps != nil {
		save.ChangePoints(stageNr, cps, config.Out.ChangePoints)
	}
	if trs != nil {
		save.TestResults(stageNr, trs, config.Out.TestResults)
	}
	if trs == nil && cps == nil {
		// save provided but no results available
		panic("Sub-program save specified but no output available")
	}
}

func handleRmDupTns(ctx context.Context, stageNr int, cps []data.ChangePoints) []data.ChangePoints {
	cps, err := changepoints.RemoveDuplicateTestNames(ctx, cps)
	if err != nil {
		panic(err)
	}
	return cps
}

func handleTRsToCPs(ctx context.Context, stageNr int, trs []data.TestResults, config input.Config) []data.ChangePoints {
	cps := make([]data.ChangePoints, 0, len(trs))
	for _, tr := range trs {
		cp, err := testresults.ToChangePoints(ctx, tr)
		if err != nil {
			panic(err)
		}
		cps = append(cps, cp)
	}
	return cps
}

func handlePlot(ctx context.Context, stageNr int, trs []data.TestResults, config input.Config) {
	for _, tr := range trs {
		save.Plots(ctx, tr, fmt.Sprintf("%s%d", util.AbsolutePath(config.Out.Plot), stageNr))
	}
}

func siso(ctx context.Context, sp string, ins []data.TestResults, in input.Config, af data.AnalysisFunc) []data.TestResults {
	l := len(ins)
	c := make(chan data.TestResults)
	done := make(chan struct{})
	for _, v := range ins {
		v := v
		go func() {
			switch sp {
			case input.SpFilter:
				c <- data.Transform(ctx, v, transFuncsFromIn(in)...)
			case input.SpAnalyse:
				c <- data.Analyse(ctx, v, af)
			default:
				fmt.Printf("ERROR - Unknown Sub-Program '%v'\n", sp)
			}
			done <- struct{}{}
		}()
	}

	go func() {
		l := l
	Loop:
		for {
			select {
			case <-ctx.Done():
				break Loop
			case <-done:
				l -= 1
				if l == 0 {
					break Loop
				}
			}
		}
		close(c)
		close(done)
	}()

	res := make([]data.TestResults, 0, l)
	for r := range c {
		res = append(res, r)
	}
	return res
}

func analysisFuncFromIn(in input.Config) data.AnalysisFunc {
	var f data.AnalysisFunc
	funcName := in.Analyse.Name
	switch funcName {
	case input.AnalyseBcp:
		probability, err := input.Float64Param(in.Analyse, 0)
		if err != nil {
			panic(err)
		}
		fn, err := analyse.Bcp(probability)
		if err != nil {
			panic(err)
		}
		f = fn
	case input.AnalyseTwitter:
		minMean, err := input.IntParam(in.Analyse, 0)
		if err != nil {
			panic(err)
		}
		fn, err := analyse.Twitter(minMean)
		if err != nil {
			panic(err)
		}
		f = fn
	case input.AnalyseTtest:
		sig, err := input.Float64Param(in.Analyse, 0)
		if err != nil {
			panic(err)
		}
		paired, err := input.BoolParam(in.Analyse, 1)
		if err != nil {
			panic(err)
		}
		fn, err := analyse.Ttest(sig, paired)
		if err != nil {
			panic(err)
		}
		f = fn
	default:
		// shoud not happen, validity of function already checked by validateAnalysisFunc
		panic(fmt.Sprintf("Invalid analysis function name '%s'", funcName))
	}
	return f
}

func transFuncsFromIn(in input.Config) []data.TransFunc {
	fs := make([]data.TransFunc, 0, len(in.Transform))
	for _, f := range in.Transform {
		switch f.Name {
		case input.FilterMinMean:
			v, err := input.Float64Param(f, 0)
			if err != nil {
				panic(err)
			}
			fs = append(fs, filter.MinMeanRuntime(v))
		case input.FilterMinMedian:
			v, err := input.Float64Param(f, 0)
			if err != nil {
				panic(err)
			}
			fs = append(fs, filter.MinMedianRuntime(v))
		case input.FilterMinVersions:
			v, err := input.IntParam(f, 0)
			if err != nil {
				panic(err)
			}
			fs = append(fs, filter.MinVersions(v))
		}
	}
	return fs
}

func parseArguments() (input.SubPrograms, input.Config) {
	i := flag.String("c", "", "config file")
	flag.Parse()

	if *i == "" {
		panic("ERROR - no configuration file provided")
	}

	// open file
	f, err := os.Open(*i)
	if err != nil {
		panic(fmt.Sprintf("ERROR - configuration file not a file: %v", err))
	}

	var d input.Config
	jd := json.NewDecoder(f)
	err = jd.Decode(&d)
	if err != nil {
		panic(fmt.Sprintf("ERROR - could not json decode configuration file: %v", err))
	}

	args := flag.Args()
	sp := input.SubPrograms{
		Count:       len(args),
		List:        args,
		Occurrences: map[string][]int{},
	}
	for k, s := range args {
		allowed := s == input.SpAnalyse ||
			s == input.SpFilter ||
			s == input.SpMerge ||
			s == input.SpPlot ||
			s == input.SpTRsToCPs ||
			s == input.SpSave ||
			s == input.SpRmDupTns
		if allowed {
			_, ok := sp.Occurrences[s]
			if ok {
				sp.Occurrences[s] = append(sp.Occurrences[s], k)
			} else {
				sp.Occurrences[s] = []int{k}
			}
		} else {
			panic(fmt.Sprintf("ERROR - dissallowed sub program specified '%s'", s))
		}
	}

	return sp, d
}

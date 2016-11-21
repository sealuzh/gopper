package main

import (
	"context"
	"encoding/csv"
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
	"bitbucket.org/sealuzh/gopper/plot"
	"bitbucket.org/sealuzh/gopper/transform/filter"
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

	var out []data.TestResults

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

	anFunc := analysisFuncFromIn(config)

	// execute sub-programs
	out = ins
	for i, sp := range sps.List {
		spUpper := strings.ToUpper(sp)
		stageNumber := i + 1
		fmt.Printf("# %d - %s: start stage\n", stageNumber, spUpper)
		startTime := time.Now()
		// decide on whether we have single or multi input and output
		// sequentially compute stages
		if sp == input.SpMerge {
			// multi-input, single-output
			out = []data.TestResults{data.Merge(ctx, out)}
		} else {
			// single-input, single-output
			// hacky solution for passing the analysis function anFunc to this function
			out = siso(ctx, sp, out, config, anFunc)
		}
		fmt.Printf("# %d - %s: finished stage in %v\n", stageNumber, spUpper, time.Since(startTime))
	}

	saveOut(out, config.Out)
}

func siso(ctx context.Context, sp string, ins []data.TestResults, in input.Config, af data.AnalysisFunc) []data.TestResults {
	l := len(ins)
	c := make(chan data.TestResults)
	done := make(chan struct{})
	for i, v := range ins {
		i := i
		v := v
		go func() {
			switch sp {
			case input.SpPlot:
				c <- plot.TimeSeries(ctx, v, fmt.Sprintf("%s%d", util.AbsolutePath(in.Plot), i))
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
		p, err := input.StringParam(in.Analyse, 0)
		if err != nil {
			panic(err)
		}
		probability, err := input.Float64Param(in.Analyse, 1)
		if err != nil {
			panic(err)
		}
		fn, err := analyse.Bcp(util.AbsolutePath(p), probability)
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
			v, err := input.Float32Param(f, 0)
			if err != nil {
				panic(err)
			}
			fs = append(fs, filter.MinMeanRuntime(v))
		case input.FilterMinMedian:
			v, err := input.Float32Param(f, 0)
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

func saveOut(d []data.TestResults, outPaths []string) {
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
				for _, r := range rs.ExecutionResults {
					w.Write(r.AsStringArray())
				}
				w.Flush()
			}
		}
	}
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
		Count: len(args),
		List:  args,
	}
	for k, s := range args {
		switch s {
		case input.SpFilter:
			sp.Transform = append(sp.Transform, k)
		case input.SpPlot:
			sp.Plot = append(sp.Plot, k)
		case input.SpMerge:
			sp.Merge = append(sp.Merge, k)
		case input.SpAnalyse:
			sp.Analyse = append(sp.Analyse, k)
		}
	}

	return sp, d
}

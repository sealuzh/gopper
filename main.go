package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"

	"bitbucket.org/sealuzh/gopper/data"
	"bitbucket.org/sealuzh/gopper/plot"
)

const (
	comma            = ';'
	spPlot           = "plot"
	spFilter         = "filter"
	spMerge          = "merge"
	transMinMean     = "minMean"
	transMinVersions = "minVersions"
	transMinMedian   = "minMedian"
)

var subProgs = [...]string{spPlot, spFilter, spMerge}
var transFuncs = [...]string{transMinMean, transMinMedian, transMinVersions}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	sps, input := parseArguments()
	validateArguments(sps, input)

	ctx := context.Background()
	ctx, f := context.WithCancel(ctx)
	defer f()

	var out []data.Results

	// read in data
	var ins []data.Results = make([]data.Results, len(input.In))
	for i, in := range input.In {
		in := path(in)
		r, err := data.ResultsFromFile(in)
		if err != nil {
			fmt.Printf("ERROR - could not read/parse file '%s': %v\n", in, err)
			return
		}
		ins[i] = r
	}

	// execute sub-programs
	out = ins
	for _, sp := range sps.List {
		// decide on whether we have single or multi inoput and output
		// sequentially compute stages
		if sp == spMerge {
			// multi-input, single-output
			out = []data.Results{data.Merge(ctx, out)}
		} else {
			// single-input, single-output
			out = siso(ctx, sp, out, input)
		}
	}

	saveOut(out, input.Out)
}

func path(p string) string {
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("No current user available: %v", err))
	}
	tilde := "~/"
	if len(p) > 2 && p[:2] == tilde {
		return filepath.Join(usr.HomeDir, p[2:])
	}
	return p
}

func siso(ctx context.Context, sp string, ins []data.Results, in data.Input) []data.Results {
	l := len(ins)
	c := make(chan data.Results)
	done := make(chan struct{})
	for i, v := range ins {
		i := i
		v := v
		go func() {
			switch sp {
			case spPlot:
				c <- plot.TimeSeries(ctx, v, fmt.Sprintf("%s%d", path(in.Plot), i))
			case spFilter:
				c <- data.Transform(ctx, v, transFuncsFromIn(in)...)
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

	res := make([]data.Results, 0, l)
	for r := range c {
		res = append(res, r)
	}
	return res
}

func transFuncsFromIn(in data.Input) []data.TransFunc {
	fs := make([]data.TransFunc, 0, len(in.Transform))
	for _, f := range in.Transform {
		switch f.TransFunc {
		case transMinMean:
			fs = append(fs, data.MinMeanRuntime(singleFloat32Param(f)))
		case transMinMedian:
			fs = append(fs, data.MinMedianRuntime(singleFloat32Param(f)))
		case transMinVersions:
			fs = append(fs, data.MinVersions(singleIntParam(f)))
		}
	}
	return fs
}

func singleFloat32Param(f data.InputTransform) float32 {
	pc := len(f.TransParams)
	if pc != 1 {
		panic(fmt.Sprintf("%s must have 1 parameter. Config provided %d", f.TransFunc, pc))
	}

	p := f.TransParams[0]
	switch p := p.(type) {
	case float64:
		return float32(p)
	default:
		panic(fmt.Sprintf("%s parameter is of incompatible type: %v", f.TransFunc, reflect.TypeOf(p)))
	}
}

func singleIntParam(f data.InputTransform) int {
	pc := len(f.TransParams)
	if pc != 1 {
		panic(fmt.Sprintf("%s must have 1 parameter. Config provided %d", f.TransFunc, pc))
	}

	p := f.TransParams[0]
	switch p := p.(type) {
	case float64:
		return int(p)
	default:
		panic(fmt.Sprintf("%s parameter is of incompatible type: %v", f.TransFunc, reflect.TypeOf(p)))
	}
}

func saveOut(d []data.Results, outPaths []string) {
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
		outPath := path(outPaths[i])
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("ERROR - Could not open output file '%v': %v", outPath, err)
		} else {
			defer f.Close()
			w := csv.NewWriter(f)
			w.Comma = comma
			defer w.Flush()
			w.Write(r.Heading())
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

func parseArguments() (data.SubPrograms, data.Input) {
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

	var d data.Input
	jd := json.NewDecoder(f)
	err = jd.Decode(&d)
	if err != nil {
		panic(fmt.Sprintf("ERROR - could not json decode configuration file: %v", err))
	}

	args := flag.Args()
	sp := data.SubPrograms{
		Count: len(args),
		List:  args,
	}
	for k, s := range args {
		switch s {
		case spFilter:
			sp.Transform = append(sp.Transform, k)
		case spPlot:
			sp.Plot = append(sp.Plot, k)
		case spMerge:
			sp.Merge = append(sp.Merge, k)
		}
	}

	return sp, d
}

func validateArguments(sps data.SubPrograms, in data.Input) {
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
		allowed := sps.Count == (len(sps.Merge) + len(sps.Plot) + len(sps.Transform))
		if !allowed {
			fmt.Printf("Sub-programs specified invalid. Every sub-program must be from: %v\n", subProgs)
			invalid = true
		}
	}

	// validate config file, in is mandatory
	if len(in.In) == 0 {
		fmt.Printf(argMissing, "config file")
		invalid = true
	}

	invalid = !validateInOut(sps, in)
	invalid = !validateTransformators(sps, in)
	invalid = !validatePlot(sps, in)

	if invalid {
		fmt.Println()
		flag.Usage()
		os.Exit(-1)
	}
}

func validateInOut(sps data.SubPrograms, in data.Input) bool {
	valid := true
	lIn := len(in.In)
	lOut := len(in.Out)
	// input files
	if lIn == 0 {
		fmt.Printf("At least one input file is mandatory (was %d)\n", lIn)
		valid = false
	} else {
		if len(sps.Merge) > 0 {
			valid = lIn > 2 && lOut == 1
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

func validateTransformators(sps data.SubPrograms, in data.Input) bool {
	valid := true

	lSpTrans := len(sps.Transform)
	if lSpTrans == 0 {
		return true
	}

	// validate Transform
	if len(in.Transform) != 0 {
		for _, t := range in.Transform {
			var contains bool
			for _, tf := range transFuncs {
				if t.TransFunc == tf {
					contains = true
					break
				}
			}

			if !contains {
				fmt.Printf("Invalid transformer function '%s'. Must be one of %v.\n", t, transFuncs)
				valid = false
				break
			}
		}
	} else {
		valid = false
	}

	return valid
}

func validatePlot(sps data.SubPrograms, in data.Input) bool {
	if len(sps.Plot) != 0 && in.Plot != "" {
		return true
	}
	return false
}

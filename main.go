package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"bitbucket.org/sealuzh/gopper/data"
	"bitbucket.org/sealuzh/gopper/plot"
)

const (
	plotDirName = "/plots"
	spPlot      = "plot"
)

var subProgs = [...]string{spPlot}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	args := parseArguments()
	validateArguments(args)

	// get data
	h, d, err := data.ResultsFromFile(args.in)
	if err != nil {
		fmt.Printf("ERROR - could not create data from file: %v", err)
	}

	ctx := context.TODO()

	var out data.Results
	for _, sp := range args.subProg {
		switch sp {
		case spPlot:
			out = plot.TimeSeries(ctx, d, args.plotDir)
		default:
			fmt.Printf("ERROR - Unknown Sub-Program '%v'\n", sp)
		}
	}

	saveOut(h, out, args.out)
}

func saveOut(heading []string, d data.Results, outPath string) {
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Printf("ERROR - Could not open output file '%v': %v", outPath, err)
		} else {
			defer f.Close()
			w := csv.NewWriter(f)
			defer w.Flush()
			w.Write(heading)
			for _, n := range d.TestNames() {
				rs, ok := d.Get(n)
				if !ok {
					// should not be the case
					fmt.Printf("ERROR - Could not retrieve test with name '%s'", n)
					continue
				}
				for _, r := range rs {
					w.Write(r.AsStringArray())
				}
				w.Flush()
			}
		}
	}
}

type args struct {
	subProg []string
	in      string
	out     string
	plotDir string
}

func parseArguments() args {
	i := flag.String("i", "", "input file")
	o := flag.String("o", "", "output file")
	plotDir := flag.String("plotDir", filepath.Dir(*i)+plotDirName, "path to directory for plots")
	flag.Parse()

	return args{
		in:      *i,
		out:     *o,
		subProg: flag.Args(),
		plotDir: *plotDir,
	}
}

func validateArguments(args args) {
	var invalid bool
	const (
		argMissing = "Argument missing: %s\n"
		argInvalid = "Argument invalid: %s\n"
	)

	// validate sub-program usage
	if len(args.subProg) == 0 {
		fmt.Printf(argMissing, "sub programm")
		invalid = true
	} else {
		var contains bool
	SpLoop:
		for _, e := range subProgs {
			for _, sp := range args.subProg {
				if sp == e {
					contains = true
					break SpLoop
				}
			}
		}
		if !contains {
			fmt.Printf("Sub-programm missing. Must be one of: %v\n", subProgs)
			invalid = true
		}
	}

	// validate input file, in is mandatory
	if args.in == "" {
		fmt.Printf(argMissing, "input file")
		invalid = true
	}

	// validate plotDir, plotDir is optional
	if args.plotDir != "" {
		fi, err := os.Stat(args.plotDir)
		if err == nil {
			if !fi.IsDir() {
				fmt.Printf(argInvalid, "plot directory is not a directory")
				invalid = true
			}
		}
	}

	if invalid {
		fmt.Println()
		flag.Usage()
		os.Exit(-1)
	}
}

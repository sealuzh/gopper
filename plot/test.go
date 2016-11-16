package plot

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"bitbucket.org/sealuzh/gopper/data"
	pl "github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
)

const (
	nrTicks     = 20
	xLabel      = "Versions"
	yLabel      = "Time"
	extension   = ".png"
	minPlotData = 3
)

var multipleTestNames = 0
var o sync.Once
var ch chan pd

type pd struct {
	plotDir string
	data    *data.TestResult
}

func TimeSeries(ctx context.Context, in data.Results, plotDir string) data.Results {
	l := in.Length()
	fmt.Printf("# Plot time series for %d tests\n", l)
	handleDirectory(plotDir)

	o.Do(func() {
		ch = make(chan pd)
		go printPlot(ch)
	})

	for _, name := range in.TestNames() {
		td, ok := in.Get(name)
		if !ok {
			panic(fmt.Sprintf("ERROR - Could not retrieve test '%s' from results", name))
		}
		o.Do(func() {
			ch = make(chan pd)
			go printPlot(ch)
		})
		ch <- pd{
			plotDir: plotDir,
			data:    td,
		}
	}

	fmt.Printf("# %d tests plotted\n", l)

	return in
}

func handleDirectory(plotDir string) {
	fi, err := os.Stat(plotDir)
	if err == nil {
		if fi.IsDir() {
			// delete content
			err = os.RemoveAll(plotDir)
			if err != nil {
				panic(fmt.Sprintf("ERROR - Could not remove plot directory '%s'", plotDir))
			}
		} else {
			panic("plotDir is a file")
		}
	} else {
		if !os.IsNotExist(err) {
			panic(fmt.Sprintf("ERROR - Unknown error occurred during stat of plotDir: %v", err))
		}
	}

	err = os.MkdirAll(plotDir, 0777)
	if err != nil {
		panic(fmt.Sprintf("ERROR - Could not create plot directory: %v", err))
	}
}

func printPlot(c <-chan pd) {
	for pd := range c {
		p, err := pl.New()
		if err != nil {
			panic("ERROR - Could not create new plot")
		}

		d := pd.data
		plotDir := pd.plotDir

		title := d.Test

		fmt.Printf("  # Plot for test '%s'\n", title)

		plotData, xTicks := plotData(d)
		dataLength := len(plotData)
		if dataLength < minPlotData {
			fmt.Printf("    DEBUG - Not enough plot data available: %d\n", dataLength)
			return
		}

		p.Title.Text = title
		p.X.Label.Text = xLabel
		p.X.Tick.Marker = xTicks
		p.X.Tick.Label.Rotation = math.Pi / 2
		p.X.Tick.Label.XAlign = draw.XRight
		p.X.Tick.Label.YAlign = draw.YCenter
		p.Y.Label.Text = yLabel

		// display data
		_, points, err := plotter.NewLinePoints(plotData)
		points.Shape = draw.CircleGlyph{}
		points.Color = color.RGBA{R: 0, G: 255, B: 255}
		points.Radius = 2

		p.Add(points)

		// filename
		i := strings.Index(title, "[")
		fileName := title
		if i != -1 {
			multipleTestNames += 1
			fileName = fmt.Sprintf("%s%d", fileName[:i], multipleTestNames)
		}
		fileName = fmt.Sprintf("%s%s", fileName, extension)
		fileName = filepath.Join(plotDir, fileName)
		err = p.Save(30*vg.Centimeter, 20*vg.Centimeter, fileName)
		if err != nil {
			fmt.Printf("    ERROR - Could not save plot: %v\n", err)
		}

	}
}

func plotData(testResult *data.TestResult) (plotter.XYs, VersionTicker) {
	d := testResult.ExecutionResults
	l := len(d)
	data := make(plotter.XYs, l, l)
	ticks := make([]pl.Tick, l, l)
	for i, r := range d {
		data[i].X = float64(i)
		data[i].Y = float64(r.RawVal)
		ticks[i].Label = r.SHA
		ticks[i].Value = float64(i)
	}
	return data, VersionTicker(ticks)
}

type VersionTicker []pl.Tick

func (t VersionTicker) Ticks(min, max float64) []pl.Tick {
	everyExact := (min + max) / nrTicks
	every := math.Floor(everyExact)

	inc := 1
	if every > 1 {
		inc = int(every)
	}

	mi := int(min)
	ma := int(max)

	ret := make([]pl.Tick, 0, nrTicks+1)
	for i := mi; i < ma; i += inc {
		ret = append(ret, pl.Tick{
			Label: t[i].Label,
			Value: t[i].Value,
		})
	}
	ret = append(ret, pl.Tick{
		Label: t[ma].Label,
		Value: t[ma].Value,
	})

	return ret
}

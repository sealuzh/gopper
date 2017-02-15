package save

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"

	pl "github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"github.com/sealuzh/gopper/data"
)

const (
	nrTicks     = 20
	xLabel      = "Versions"
	yLabel      = "Time"
	extension   = ".png"
	minPlotData = 3
)

var multipleTestNames = 0

type pd struct {
	plotDir string
	data    data.TestResult
}

func Plots(ctx context.Context, in data.TestResults, plotDir string) {
	//TODO: support multiple stages, e.g. return parameterless function
	l := in.Len()
	fmt.Printf("  Plot time series for %d tests\n", l)
	handleDirectory(plotDir)

	ch := make(chan pd)
	done := make(chan int)
	go printPlot(ch, done)

	for _, name := range in.TestNames() {
		td, ok := in.Get(name)
		if !ok {
			panic(fmt.Sprintf("ERROR - Could not retrieve test '%s' from results", name))
		}
		ch <- pd{
			plotDir: plotDir,
			data:    td,
		}
	}
	close(ch)
	printed := <-done

	fmt.Printf("  %d tests plotted\n", printed)
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

func printPlot(c <-chan pd, done chan<- int) {
	counter := 0
	for pd := range c {
		func() {
			p, err := pl.New()
			if err != nil {
				panic("ERROR - Could not create new plot")
			}

			d := pd.data
			plotDir := pd.plotDir

			title := d.Test()

			fmt.Printf("    Plot for test '%s'\n", title)

			//plotData, cps, xTicks := plotData(d)
			plotData, cps, xTicks := boxPlots(d)
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

			// display boxPlots
			p.Add(plotData...)
			p.Add(cps...)

			/*// display data
			points, err := plotter.NewScatter(plotData)
			points.Shape = draw.CircleGlyph{}
			points.Color = color.RGBA{R: 0, G: 255, B: 255}
			points.Radius = 2
			p.Add(points)

			// cps
			cpPoints, err := plotter.NewScatter(cps)
			cpPoints.Shape = draw.CircleGlyph{}
			cpPoints.Color = color.RGBA{R: 255, G: 255, B: 0}
			cpPoints.Radius = 2
			p.Add(cpPoints)
			*/
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
			counter++
		}()
	}
	done <- counter
}

func boxPlots(testResult data.TestResult) ([]pl.Plotter, []pl.Plotter, VersionTicker) {
	commits := testResult.Commits()
	cps := testResult.ChangePoints()

	lc := len(commits)
	lcps := cps.Len()

	bpsData := make([]pl.Plotter, 0, lc-lcps)
	bpsCps := make([]pl.Plotter, 0, lcps)
	ticks := make([]pl.Tick, lc)

	for i, c := range commits {
		ers, ok := testResult.ExecutionResults(c)
		if !ok {
			panic(fmt.Sprintf("Inconsistent test result: %s @ %s", testResult.Test(), c))
		}
		b, err := plotter.NewBoxPlot(vg.Points(20), float64(i), plotter.Values(ers.Values()))
		if err != nil {
			panic(err)
		}

		hasCps := cps.At(c)
		isCp := hasCps.Len() > 0

		if isCp {
			c := color.RGBA{R: 0, G: 255, B: 255}
			b.MedianStyle.Color = c
			b.BoxStyle.Color = c
			b.GlyphStyle.Color = c
			b.WhiskerStyle.Color = c
			bpsCps = append(bpsCps, b)
		} else {
			bpsData = append(bpsData, b)
		}

		ticks[i].Label = c
		ticks[i].Value = float64(i)
	}

	return bpsData, bpsCps, VersionTicker(ticks)
}

/*func plotData(testResult data.TestResult) (plotter.XYs, plotter.XYs, VersionTicker) {
	d := testResult.ExecutionResults
	l := len(d)
	lcps := len(testResult.ChangePoints.All())
	data := make(plotter.XYs, l-lcps)
	dataCount := 0
	cps := make(plotter.XYs, lcps)
	cpCount := 0
	ticks := make([]pl.Tick, l, l)

	for i, r := range d {
		_, ok := testResult.ChangePoints.Get(r.SHA)
		if ok {
			cps[cpCount].X = float64(i)
			cps[cpCount].Y = float64(r.RawVal)
			cpCount++
		} else {
			data[dataCount].X = float64(i)
			data[dataCount].Y = float64(r.RawVal)
			dataCount++
		}
		ticks[i].Label = r.SHA
		ticks[i].Value = float64(i)
	}
	return data, cps, VersionTicker(ticks)
}*/

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

# gopper - Performance History Anaylsis
gopper is a command line tool that takes performance metrics from tests (performance tests or unit tests) over multiple versions as input and applies different operations to it.
Typical test results are multiple tests (with possibly multiple executions) for multiple versions (e.g. git commits) of a software.
Supported operations (aka sub-programs) are merging, filtering, anaylsing, plotting and saving.

## Installation
* install [Go programming language](https://golang.org) (tested with version go1.7.4 darwin/amd64)
* Set $GOROOT and $GOPATH accordingly
* install [Glide](https://github.com/Masterminds/glide)
* run `glide install` from the base directory
* install [Docker](https://www.docker.com)
* run `docker pull sealuzh/gopper-rserve`
* run `docker run -tid --name gopper-rserve -p 6311:6311 sealuzh/gopper-rserve`

## Usage
After the installation as successful and the gopper-rserve container is running, gopper is ready to be used for historical performance analysis.

### Command Line Parameters
gopper takes a single input file in json format that is specified by the parameter -c, and a non-empty set of sup-program parameters.

The following sup-programs are supported:

* `merge` - merges n input files to a single output file.
* `filter` - filters the performance metrics. See section "Configuration File" for details.
* `analyse` - analyses performance metrics with respect to performance changes.
* `plot` - plots the performance results as either line or box plots.
* `toChangePoints` - transforms test results into change points.
* `save` - saves test results and (if previously transformed) change points.

All sub-programs are optional an can be applied in any possible order, as long as at least one is present.
A sample execution of gopper, which filters and analyses perfromance metrics and then plots and saves the reuslts, may look like:
```
gopper -c config.json filter analyse plot toChangePoints save
```
This execution takes the sample configuration file from the next section.

### Configuration File
The configuration file specifies the details that are necessary for an execution of gopper. It is in [JSON](json.org)-format and looks like the one below. The four main elements are:

* "IN" - a non-empty list of input files. The format is CSV, exactly the same output as [hopper](https://github.com/sealuzh/hopper).
* "OUT" - three different out types are possible:
    * "TestResults" - the possible filtered (with sub-program `filter`) input files, with the same format. Supports multiple paths, in case multiple "IN" paths were provided and sup-program `merge` was not executed (same amount required).
    * "ChangePoints" - the detected change points by the anaylsis function ("Analyse"). Change points are only saved if the sub-program `toChangePoints` was executed. Same as with "TestResults", multiple output paths are supported.
    * "Plot" - specifies the path to the plot directory. Saving of plots requires executing the `plot`sub-program.
* "Analyse" - Specifies the type of analysis function ("Name") and its parameters ("Params"):
    * "ttest" - Welch's T-Test. for multiple performance metrics per test per version. Parameters: significance level [float]; paired T-test [bool]
    * "bcp" - [Bayesian Change Point Analysis](https://cran.r-project.org/web/packages/bcp/bcp.pdf). For single performance metrics per test per version. Parameters: probability [float]
    * "twitter" - [Twitter's BreakoutDetection](https://github.com/twitter/BreakoutDetection). For single performance metrics per test per version. Parameters:
* "Transform" - Specifies the filter rules applied with the sub-program `filter`. Three different filters are available:
    * "minVersion" - Test metrics with less than n versions ("Params") are filtered.
    * "minMean" - Test metrics with a mean value over all versions with less then x ("Params") are filtered.
    * "minMedian - Test metrics with a median value over all versions with less then x ("Params") are filtered.

```JSON
{
	"In": [
		"~/gopper/in.csv"
	],
	"Out": {
		"TestResults": ["~/gopper/out.csv"],
		"ChangePoints": ["~/gopper/cps.json"],
		"Plot": "~/gopper/plots"
	},
	"Analyse": {
		"Name": "ttest",
		"Params": [0.99, true]
	},
	"Transform": [
		{
			"Name": "minVersions",
			"Params": [5]
		},
		{
			"Name": "minMean",
			"Params": [0.01]
		},
		{
			"Name": "minMedian",
			"Params": [0.01]
		}
	]
}
```
package input

const (
	SpPlot            = "plot"
	SpFilter          = "filter"
	SpMerge           = "merge"
	SpAnalyse         = "analyse"
	SpTRsToCPs        = "toChangePoints"
	SpSave            = "save"
	FilterMinMean     = "minMean"
	FilterMinVersions = "minVersions"
	FilterMinMedian   = "minMedian"
	AnalyseBcp        = "bcp"
	AnalyseTwitter    = "twitter"
)

var SubProgs = [...]string{SpPlot, SpFilter, SpMerge, SpAnalyse, SpTRsToCPs, SpSave}
var TransFuncs = [...]string{FilterMinMean, FilterMinMedian, FilterMinVersions}
var AnalyseFuncs = [...]string{AnalyseBcp, AnalyseTwitter}

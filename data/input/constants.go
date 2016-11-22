package input

const (
	SpPlot            = "plot"
	SpFilter          = "filter"
	SpMerge           = "merge"
	SpAnalyse         = "analyse"
	FilterMinMean     = "minMean"
	FilterMinVersions = "minVersions"
	FilterMinMedian   = "minMedian"
	AnalyseBcp        = "bcp"
	AnalyseTwitter    = "twitter"
)

var SubProgs = [...]string{SpPlot, SpFilter, SpMerge, SpAnalyse}
var TransFuncs = [...]string{FilterMinMean, FilterMinMedian, FilterMinVersions}
var AnalyseFuncs = [...]string{AnalyseBcp, AnalyseTwitter}

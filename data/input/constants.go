package input

const (
	SpPlot            = "plot"
	SpFilter          = "filter"
	SpMerge           = "merge"
	SpAnalyse         = "analyse"
	SpTRsToCPs        = "toChangePoints"
	SpRmDupTns        = "rmDuplicates"
	SpSave            = "save"
	FilterMinMean     = "minMean"
	FilterMinVersions = "minVersions"
	FilterMinMedian   = "minMedian"
	AnalyseBcp        = "bcp"
	AnalyseTwitter    = "twitter"
	AnalyseTtest      = "ttest"
)

var SubProgs = [...]string{SpPlot, SpFilter, SpMerge, SpAnalyse, SpTRsToCPs, SpSave, SpRmDupTns}
var TransFuncs = [...]string{FilterMinMean, FilterMinMedian, FilterMinVersions}
var AnalyseFuncs = [...]string{AnalyseBcp, AnalyseTwitter, AnalyseTtest}

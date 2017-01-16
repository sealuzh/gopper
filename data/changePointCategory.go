package data

type ChangeCategory int

const (
	Naughties ChangeCategory = iota
	Tens
	Twenties
	Thirties
	Fourties
	Fifties
	Sixties
	Seventies
	Eighties
	Nineties
)

const (
	changeCategoryCount = 10
)

func allChangeCategories() []ChangeCategory {
	ret := make([]ChangeCategory, changeCategoryCount)
	for i := 0; i < changeCategoryCount; i++ {
		ret[i] = ChangeCategory(i)
	}
	return ret
}

func calcChangeCategory(mean1, mean2 float64) ChangeCategory {
	var bigger float64
	var smaller float64
	if mean1 < mean2 {
		bigger = mean2
		smaller = mean1
	} else {
		bigger = mean1
		smaller = mean2
	}

	diff := bigger - smaller
	changePercentage := diff / bigger

	var cc ChangeCategory
	if changePercentage < 0.5 {
		// below 50
		if changePercentage < 0.1 {
			cc = Naughties
		} else if changePercentage < 0.2 {
			cc = Tens
		} else if changePercentage < 0.3 {
			cc = Twenties
		} else if changePercentage < 0.4 {
			cc = Thirties
		} else {
			cc = Fourties
		}
	} else {
		// above 50
		if changePercentage < 0.6 {
			cc = Fifties
		} else if changePercentage < 0.7 {
			cc = Sixties
		} else if changePercentage < 0.8 {
			cc = Seventies
		} else if changePercentage < 0.9 {
			cc = Eighties
		} else {
			cc = Nineties
		}
	}
	return cc
}

package data

type Input struct {
	In        []string
	Out       []string
	Transform []InputTransform
	Plot      string
}

type InputTransform struct {
	TransFunc   string
	TransParams []interface{}
}

type SubPrograms struct {
	Count     int
	List      []string
	Plot      []int
	Transform []int
	Merge     []int
}

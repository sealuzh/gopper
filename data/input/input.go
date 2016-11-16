package input

type Config struct {
	In        []string
	Out       []string
	Transform []Transform
	Plot      string
}

type Transform struct {
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

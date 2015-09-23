package netest

type PatternType uint8

const (
	AllOnesPattern PatternType = iota
	AllZerosPattern
)

var Patterns = [][]byte{
	[]byte{0xff},
	[]byte{0x00},
}

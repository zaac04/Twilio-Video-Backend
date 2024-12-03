package mediaconvert

import "fmt"

type Definition int32
type File struct {
	VideoUrl string
	AudioUrl string
}

var definitionStrings = map[Definition]string{
	HD720p: "HD720p",
	SD480p: "SD480p",
	SD360p: "SD360p",
}

func (d Definition) ToString() string {
	if str, ok := definitionStrings[d]; ok {
		return str
	}
	return fmt.Sprintf("Unknown (%d)", d)
}

const (
	HD720p Definition = iota
	SD480p
	SD360p
)

type resolution struct {
	width, height, bitrate int32
}

var ResolutionMap = map[Definition]resolution{
	HD720p: {width: 1280, height: 720, bitrate: 5000000},
	SD480p: {width: 854, height: 480, bitrate: 2500000},
	SD360p: {width: 640, height: 360, bitrate: 1000000},
}

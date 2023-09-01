package cat

const (
	DebugMode   = "debug"
	ReleaseMode = "release"
	//TestMode = "test"
)

var catMode = ReleaseMode

func SetMode(mode string) {
	if mode == "" {
		return
	}
	switch mode {
	case DebugMode:
		catMode = mode
	default:
		panic("cat mode unknown: " + mode + " (available mode: debug)")
	}
}

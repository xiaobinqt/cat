package cat

import "log"

func IsDebugging() bool {
	return catMode == DebugMode
}

func debugPrintf(format string, value ...interface{}) {
	if IsDebugging() {
		log.Printf(format, value...)
	}
}

func debugPrintln(value ...interface{}) {
	if IsDebugging() {
		log.Println(value...)
	}
}

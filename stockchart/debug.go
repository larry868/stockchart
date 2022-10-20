package stockchart

import "fmt"

type DebugOutput int

const (
	DBG_OFF       DebugOutput = 0b000000000
	DBG_REDRAW    DebugOutput = 0b000000001
	DBG_EVENT     DebugOutput = 0b000000010
	DBG_RESIZE    DebugOutput = 0b000000100
	DBG_SELCHANGE DebugOutput = 0b000001000
	DBG_ALL       DebugOutput = 0b11111111
)

// Change this global variable to activate debug mode
var DEBUG = DBG_OFF //DBG_SELCHANGE | DBG_REDRAW

// if debug match with the DEBUG global flag, then Debug print out strprint
func Debug(debug DebugOutput, format string, data ...any) {
	if DEBUG&debug > 0 {
		fmt.Printf(format, data...)
		fmt.Println()
	}
}

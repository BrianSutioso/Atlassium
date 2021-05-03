package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

// Debug is optional logger for debugging
var Debug *log.Logger

// Out is logger to Stdout
var Out *log.Logger

// Err is logger to Stderr
var Err *log.Logger

// init initializes the loggers.
func init() {
	Debug = log.New(ioutil.Discard, "DEBUG: ", 0)
	Out = log.New(os.Stdout, "INFO: ", log.Ltime|log.Lshortfile)
	Err = log.New(os.Stderr, "ERROR: ", log.Ltime|log.Lshortfile)
}

// SetDebug turns debug print statements on or off.
func SetDebug(enabled bool) {
	if enabled {
		Debug.SetOutput(os.Stdout)
	} else {
		Debug.SetOutput(ioutil.Discard)
	}
}

func FmtAddr(addr string) string {
	if addr == "" {
		return ""
	}
	colors := []string{"\033[41m", "\033[42m", "\033[43m", "\033[44m", "\033[45m", "\033[46m", "\033[47m"}
	port, _ := strconv.ParseInt(strings.Split(addr, ":")[1], 10, 64)
	randomColor := colors[int(port)%len(colors)]
	return fmt.Sprintf("%v\033[97m[%v]\033[0m", randomColor, addr)
}

func Colorize(s string, seed int) string {
	lowestColor, highestColor := 104, 226
	return fmt.Sprintf("\033[38;5;%vm%v\033[0m", seed%(highestColor-lowestColor)+lowestColor, s)
}

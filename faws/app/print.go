package app

import (
	"fmt"

	"github.com/faws-vcs/console"
)

// Fatal terminates the program and displays args as a fatal error message
func Fatal(args ...any) {
	console.Fatal(args...)
}

// Info displays args as a message containing generic information
func Info(args ...any) {
	console.Println(args...)
}

// Warning displays args as a warning message
func Warning(args ...any) {
	console.Println(args...)
}

// Quote makes args stand out in the terminal, use this for items demanding special attention
func Quote(args ...any) {
	console.Quote(args...)
}

// Log is currently the same as Info, but may in the future be used to write to log files
func Log(args ...any) {
	console.Println(args...)
}

// Header displays args as a title message underlined by dashes, use this to highlight the name of an object when pretty-printing (faws cat-file -p)
func Header(args ...any) {
	console.Header(fmt.Sprint(args...))
}

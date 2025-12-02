package app

import (
	"fmt"

	"github.com/faws-vcs/console"
)

func Fatal(args ...any) {
	console.Fatal(args...)
}

func Info(args ...any) {
	console.Println(args...)
}

func Warning(args ...any) {
	console.Println(args...)
}

func Quote(args ...any) {
	console.Quote(args...)
}

func Log(args ...any) {
	console.Println(args...)
}

func Header(args ...any) {
	console.Header(fmt.Sprint(args...))
}

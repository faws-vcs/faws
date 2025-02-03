package app

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/fatih/color"
)

func Fatal(args ...any) {
	color.Set(color.FgRed)
	fmt.Print("fatal: ")
	color.Unset()
	fmt.Println(args...)
	os.Exit(1)
}

func Log(args ...any) {
	fmt.Println(args...)
}

func Info(args ...any) {
	fmt.Println(args...)
}

func Warning(args ...any) {
	Log(args...)
}

func Header(str string) {
	fmt.Println(str)
	for i := 0; i < utf8.RuneCountInString(str); i++ {
		fmt.Print("-")
	}
	fmt.Println()
}

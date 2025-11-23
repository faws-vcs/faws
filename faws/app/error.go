package app

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/pterm/pterm"
)

// func Fatal(args ...any) {
// 	color.Set(color.FgRed)
// 	fmt.Print("fatal: ")
// 	color.Unset()
// 	fmt.Println(args...)
// 	os.Exit(1)
// }

// func Log(args ...any) {
// 	fmt.Println(args...)
// }

// func Info(args ...any) {
// 	fmt.Println(args...)
// }

// func Warning(args ...any) {
// 	Log(args...)
// }

// func Header(str string) {
// 	fmt.Println(str)
// 	for i := 0; i < utf8.RuneCountInString(str); i++ {
// 		fmt.Print("-")
// 	}
// 	fmt.Println()
// }

// func Quote(args ...any) {
// 	fmt.Printf("\n  %s\n\n", fmt.Sprint(args...))
// }

func Fatal(args ...any) {
	// color.Set(color.FgRed)
	// fmt.Print("fatal: ")
	// color.Unset()
	// fmt.Println(args...)
	// os.Exit(1)
	multi := pterm.DefaultMultiPrinter

	text_printer := pterm.DefaultBasicText.WithWriter(multi.NewWriter())
	text_printer.Print(pterm.LightRed("fatal: "))
	text_printer.Println(args...)
	multi.Stop()
	os.Exit(1)
}

func Log(args ...any) {

	fmt.Println(args...)
}

func Info(args ...any) {
	// multi := pterm.DefaultMultiPrinter
	// text_printer := pterm.DefaultBasicText.WithWriter(multi.NewWriter())
	// text_printer.Println(args...)
	pterm.Println(args...)
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

func Quote(args ...any) {
	Println(fmt.Sprintf("\n  %s\n", fmt.Sprint(args...)))
}

func Stage(n int, message string) {

}

func StartPullSpinner() {

}

func Println(args ...any) {
	pterm.Println(args...)
}

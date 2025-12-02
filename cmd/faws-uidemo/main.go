package main

import (
	"time"

	"github.com/faws-vcs/faws/faws/app/console"
)

func main() {
	// app.Info("hello world")
	// // Create a multi printer for managing multiple printers
	// multi := pterm.DefaultMultiPrinter

	// // Create two spinners with their own writers
	// pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Spinner 1")
	// // spinner2, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Spinner 2")

	// // spinner1.Start()

	// multi.Start()

	// time.Sleep(50 * time.Second)
	// // spinner1.Stop()

	// app.Fatal("goodbye cruel world")
	console.Open()

	// console.Fatal("goodbye cruel world")

	progress := float64(0)

	var spin console.Spinner
	spin.Stylesheet.Sequence[0] = console.Cell{'⡿', console.BrightBlue, 0}
	spin.Stylesheet.Sequence[1] = console.Cell{'⣟', console.BrightBlue, 0}
	spin.Stylesheet.Sequence[2] = console.Cell{'⣯', console.BrightBlue, 0}
	spin.Stylesheet.Sequence[3] = console.Cell{'⣷', console.BrightBlue, 0}
	spin.Stylesheet.Sequence[4] = console.Cell{'⣾', console.BrightBlue, 0}
	spin.Stylesheet.Sequence[5] = console.Cell{'⣽', console.BrightBlue, 0}
	spin.Stylesheet.Sequence[6] = console.Cell{'⣻', console.BrightBlue, 0}
	spin.Stylesheet.Sequence[7] = console.Cell{'⢿', console.BrightBlue, 0}
	spin.Frequency = 300 * time.Millisecond

	console.RenderFunc(func(h *console.Hud) {
		var text console.Text
		text.Stylesheet.Width = 30
		text.Add("this is a hud message", console.BrightGreen, console.None)
		h.Line(&text)

		var progress_indicator console.Text
		progress_indicator.Stylesheet.Width = 16
		progress_indicator.Add("Progress n%    ", console.Black, console.BrightGreen)

		var pb console.ProgressBar
		pb.Stylesheet.Width = console.Width() - 16
		pb.Stylesheet.Sequence[console.PbCaseLeft] = console.Cell{'[', 0, 0}
		pb.Stylesheet.Sequence[console.PbCaseRight] = console.Cell{']', 0, 0}
		pb.Stylesheet.Sequence[console.PbFluid] = console.Cell{'=', 0, 0}
		pb.Stylesheet.Sequence[console.PbVoid] = console.Cell{' ', 0, 0}
		pb.Stylesheet.Sequence[console.PbTail] = console.Cell{'<', 0, 0}
		pb.Stylesheet.Sequence[console.PbHead] = console.Cell{'>', 0, 0}
		pb.Progress = progress
		h.Line(&progress_indicator, &pb)
		h.Line(&progress_indicator, &pb)
		h.Line(&progress_indicator, &pb)
		h.Line(&progress_indicator, &pb)

		h.Line(&spin)
	})

	// go func() {
	for x := 0; x < 20; x++ {
		time.Sleep(100 * time.Millisecond)
		console.Println("doing stuff...")
		time.Sleep(300 * time.Millisecond)
		progress = float64(x) / float64(20)
		console.SwapHud()
	}

	// console.Fatal("there was a critical error")
	// }()

	// console.RenderFunc(func(hud *console.Hud) {
	// 	var thinker

	// 	hud.Line()
	// })

	// // console.Fatal("yo thats not cool yo!!!")
	// console.Thinker("Lmao")

	// for x := 0; x < 50; x++ {
	// 	time.Sleep(300 * time.Millisecond)
	// 	console.StepThinker()
	// 	if x == 40 {
	// 		console.Thinker("Wrapping up...")
	// 	}
	// }

	console.Close()
}

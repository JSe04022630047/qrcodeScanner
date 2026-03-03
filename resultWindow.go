package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/makiuchi-d/gozxing"
	// "github.com/pkg/browser"
)

func showResultWindow(result scannedCode) {
	resultWin := fyne.CurrentApp().NewWindow("Result")
	
	titleLabel := widget.NewLabel("Result")
	// content of the qrcode
	stringEntry := widget.NewEntry()
	formatEntry := widget.NewEntry()
	timestampEntry := widget.NewEntry()

	stringEntry.SetText(result.Text)
	formatEntry.SetText(gozxing.BarcodeFormat(result.Format).String()) // hack yes but it works so I don't bother
	tm := time.Unix(result.Timestamp, 0) // we already converted the milisec to normal seconds
	timestampEntry.SetText(tm.Format(time.RFC3339))

	// disable all entries
	stringEntry.Disable(); formatEntry.Disable(); timestampEntry.Disable()

	form := widget.NewForm(
		widget.NewFormItem("Text", stringEntry),
		widget.NewFormItem("Format", formatEntry),
		widget.NewFormItem("Timestamp", timestampEntry),
	)


	mainContent := container.NewVBox(titleLabel,form)
	resultWin.SetContent(mainContent)
	resultWin.Show()
}
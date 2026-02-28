package main

import (
	// "fmt"
	// "image"

	// _ "image/png" // Register PNG decoder
	// "os"

	// "github.com/makiuchi-d/gozxing"
	// "github.com/makiuchi-d/gozxing/qrcode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var GitCommit string = "Unknown"

type scannedCode struct {
	theString string /*anything that can be scanned, only filtered to strings. binary data may be parsed into string and show garbage data*/
}

func main() {
	theApp := app.New()
	mainWindow := theApp.NewWindow("QR Scanner Tool")
	mainWindow.SetMaster()
	mainWindow.Resize(fyne.NewSize(600, 300))
	mainWindow.SetFixedSize(true)
	mainWindow.CenterOnScreen()

    // Toolbar

    mainToolBar := widget.NewToolbar(
        widget.NewToolbarAction(theme.InfoIcon(), func ()  {
            dialog.ShowInformation("This is a test", "Title", mainWindow)
        }),
    )

	// content

	// center buttons

	buttonPickFile := widget.NewButtonWithIcon("Pick file", theme.FolderOpenIcon(), func() {
		// TODO
	})
	buttonCamera := widget.NewButtonWithIcon("Open Camera", theme.MediaPhotoIcon(), func() {
		// TODO
	})
	buttonScreenCapture := widget.NewButtonWithIcon("Scan the screen", theme.ViewFullScreenIcon(), func() {
		// TODO
	})

	buttonHistory := widget.NewButtonWithIcon("History", theme.HistoryIcon(), func() {
		// TODO
	})

	buttonCreateQR := widget.NewButtonWithIcon("Create QR Code", theme.DocumentCreateIcon(), func() {
		// TODO
	})

	// main labels
	mainOptionsLabel := widget.NewLabel("Select Operation")
	mainOptionsLabel.Alignment = fyne.TextAlignCenter
	mainOptionsLabel.TextStyle.Bold = true

	mainAltOptionsLabel := widget.NewLabel("Or alternatively")
	mainAltOptionsLabel.Alignment = fyne.TextAlignCenter
	mainAltOptionsLabel.TextStyle.Bold = true

	mainFunctionButtons :=
		container.NewCenter(
			container.NewVBox(mainOptionsLabel,
				container.NewHBox(
					buttonPickFile,
					buttonCamera,
					buttonScreenCapture,
				),
				mainAltOptionsLabel,
				container.NewCenter(container.NewHBox(
					buttonHistory,
					buttonCreateQR,
				),
				),
			),
		)
	// show

	mainWindow.SetContent(container.NewVBox(
        mainToolBar,
        layout.NewSpacer(),
		mainFunctionButtons,
        layout.NewSpacer(),
		/* test stuff
		   label,
		   button,
		   content1,
		   content2,
		*/
	))

	mainWindow.ShowAndRun()
}

func changeLabel(l *widget.Label) {
	l.SetText("Scanning logic would go here!")
}

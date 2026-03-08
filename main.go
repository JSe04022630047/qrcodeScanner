package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/ncruces/zenity"
)

var GitCommit string = "Unknown"
var BuildTime string = "please compile with the script"
var Version string = "a0" // This automatically got assigned when compiling with the script, please run the script to push in production

var globalQRReader gozxing.Reader = qrcode.NewQRCodeReader()
var HistoryButton *widget.Button = nil

// variable that the progam actually used
var scannedHistory []scannedCode = nil

func printScanHistoryArr(array []scannedCode) {
	if array == nil {
		fmt.Println("can't print loaded scan history, no entry")
		return
	}

	var out strings.Builder
	out.WriteString("below is scan histroy that program loaded\n")
	for i := range array {
		out.WriteString(fmt.Sprintf("- %s\n", array[i]))
	}
	fmt.Println(out.String())
}

func printCurrScanHistory() {
	printScanHistoryArr(scannedHistory)
}

func handleFlags() {
	versionFlag := flag.Bool("version", false, "Print version information and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("QRScan Tool\n")
		fmt.Printf("Build Commit: %s\n", GitCommit)
		fmt.Printf("Build Time:   %s\n", BuildTime)
		fmt.Printf("Version:   %s\n", Version)
		os.Exit(0) // Exit early so the app doesn't run
	}

}

func main() {
	handleFlags()

	// the argument is for recursive, it returns true if successfully decoded the saved history, else return false.
	// the return may not be needed but I thought its better to indicate if loading history failed
	loadHistory(0)

	//debug print
	printCurrScanHistory()

	theApp := app.New()
	mainWindow := theApp.NewWindow("QR Scanner Tool")
	mainWindow.SetMaster()
	mainWindow.Resize(fyne.NewSize(600, 300))
	mainWindow.SetFixedSize(true)
	mainWindow.CenterOnScreen()

	// Toolbar

	mainToolBar := widget.NewToolbar(
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			showVersionDialogv2(mainWindow)
		}),
	)

	// content

	// main buttons

	buttonPickFile := widget.NewButtonWithIcon("Pick file", theme.FolderOpenIcon(), func() {
		fmt.Println("Pick file entry")
		file, errFilePicker := zenity.SelectFile(
			zenity.Title("Select image"),
			zenity.FileFilters{
				{Name: "Image files", Patterns: []string{"*.png", "*.jpg", "*.jpeg", "*.gif", "*.webp"}},
			},
		)

		if errFilePicker != nil {
			if errFilePicker == zenity.ErrCanceled {
				fmt.Println("User cancelled")
			} else {
				fmt.Println("Error:", errFilePicker)
			}
			return
		}

		// test print lol
		fmt.Println(file)

		fmt.Println("Opening the file at ", file)
		var fileOp, errOp = os.Open(file)
		if errOp != nil {
			log.Fatal(errOp)
		} else {
			imageOp, _, _ := image.Decode(fileOp)
			showCropWindow(
				mainWindow,
				imageOp,
				func(img image.Image) { // User hit accept
					fmt.Println("cropped successfully")
					bmp, _ := gozxing.NewBinaryBitmapFromImage(img) // it is expected that images are always entered here

					result, err := globalQRReader.Decode(bmp, nil)

					if err == nil {
						fmt.Println(result.String(), result.GetBarcodeFormat(), result.GetTimestamp())
						addEntryResult(*result)
						printCurrScanHistory()
						showResultWindow(scannedHistory[0], nil, 0) // sorting from addEntryResult always make the latest one the 0
					} else if _, ok := err.(gozxing.NotFoundException); ok {
						fmt.Println("qrcode scanned failed, qrcode not found", err)
						dialog.ShowError(err, mainWindow)
					} else {
						fmt.Println("qrcode scanned failed", err)
						dialog.ShowError(err, mainWindow)
					}
				},
				func() { // user hit cancled or exit
					fmt.Println("cropping canceled")
				},
			)
		}

	})
	buttonScreenCapture := widget.NewButtonWithIcon("Scan the screen", theme.ViewFullScreenIcon(),
		func() {
			showScreenCapCropWindow(mainWindow, func(img image.Image) {
				fmt.Println("cropped from ss successfully")

				bmp, _ := gozxing.NewBinaryBitmapFromImage(img)
				result, err := globalQRReader.Decode(bmp, nil)

				if err == nil {
					fmt.Println(result.String(), result.GetBarcodeFormat(), result.GetTimestamp())
					addEntryResult(*result)
					printCurrScanHistory()
					showResultWindow(scannedHistory[0], nil, 0) // sorting from addEntryResult always make the latest one the 0
				} else if _, ok := err.(gozxing.NotFoundException); ok {
					fmt.Println("qrcode scanned failed, qrcode not found", err)
					dialog.ShowError(err, mainWindow)

				} else {
					fmt.Println("qrcode scanned failed", err)
					dialog.ShowError(err, mainWindow)
				}
			},
				func() { // user hit cancled or exit
					fmt.Println("cropping canceled")
				},
			)
		})

	buttonHistory := widget.NewButtonWithIcon("History", theme.HistoryIcon(), func() {
		showHistoryWindow()
	})
	if len(scannedHistory) < 1 {
		buttonHistory.Disable()
	}
	HistoryButton = buttonHistory

	buttonCreateQR := widget.NewButtonWithIcon("Create QR Code", theme.DocumentCreateIcon(), func() {
		showCreateQRWindow(mainWindow)
	})

	// main labels
	mainOptionsLabel := widget.NewLabel("Scan QR Codes")
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
					// buttonCamera,
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
	))

	mainWindow.SetOnClosed(func() {
		updateHistoryFile()
	})

	mainWindow.ShowAndRun()

}

func changeLabel(l *widget.Label) {
	l.SetText("Scanning logic would go here!")
}

func versionText() string {
	var baseString string = `Version: %s
Build Commit: %s
Build Time: %s`
	return fmt.Sprintf(baseString, Version, GitCommit, BuildTime)
}

func showVersionDialog(win fyne.Window) {
	titleLabel := widget.NewLabel("Version")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	content := widget.NewLabel(versionText())
	content.TextStyle = fyne.TextStyle{Monospace: true}

	contentHolder := container.NewVBox(titleLabel, content)

	// Create the actual dialog
	// empty string "" is the title; we're using custom title, "OK" is the button text
	d := dialog.NewCustom("", "OK", contentHolder, win)

	// optional resize, fyne defaults to minimum size, I prefer the minimum though.
	// d.Resize(fyne.NewSize(300, 150))

	d.Show()
}

func showVersionDialogv2(win fyne.Window) {
	var popup *widget.PopUp

	// Setup the Custom Title (Bold & Styled)
	title := canvas.NewText("Version", theme.Color(theme.ColorNameForeground))
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 16
	title.Alignment = 1 // center

	// Setup the Monospace Content
	content := widget.NewLabel(versionText())
	content.TextStyle = fyne.TextStyle{Monospace: true}

	// Create the copy to clipboard button
	copyButton := widget.NewButtonWithIcon("Copy to clipboard", theme.ContentCopyIcon(), func() {
		fyne.CurrentApp().Clipboard().SetContent(versionText())
	})

	// Create the "OK" Button to close the popup
	closeButton := widget.NewButton("Dismiss", func() {
		popup.Hide()
	})
	closeButton.Importance = widget.HighImportance

	// 4. Assemble the Layout
	// We use a VBox and some padding to make it look professional
	footer := container.NewHBox(layout.NewSpacer(), copyButton, closeButton)

	mainLayout := container.NewPadded(
		container.NewVBox(
			title,
			widget.NewSeparator(),
			content,
			container.NewHBox(layout.NewSpacer()), // Push button to bottom
			footer,
		),
	)

	// 5. Create and Show the Modal
	popup = widget.NewModalPopUp(mainLayout, win.Canvas())
	// popup.Resize(fyne.NewSize(350, 200))
	popup.Show()
}

func ChangeHistoryButtonStatus() {
	
	if HistoryWindow == nil {
		historyButton := HistoryButton
		if len(scannedHistory) < 1 {
			historyButton.Disable()
		} else {
			historyButton.Enable()
		}
	} else {
		HistoryButton.Disable()
	}
}

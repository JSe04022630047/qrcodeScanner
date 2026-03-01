package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"

	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
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
var Version string = "a0" // TODO

var globalQRReader gozxing.Reader = qrcode.NewQRCodeReader()

type scannedCode struct {
	theString string /*anything that can be scanned, only filtered to strings. binary data may be parsed into string and show garbage data*/
}

// Scan QR code stuff below

// end qr code scan stuff

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
					} else if _, ok := err.(gozxing.NotFoundException); ok{
						fmt.Println("qrcode scanned failed, qrcode not found", err)
					} else {
						fmt.Println("qrcode scanned failed", err)
					}
				},
				func() { // user hit cancled or exit
					fmt.Println("cropping canceled")
				},
			)
		}

	})

	buttonCamera := widget.NewButtonWithIcon("Open Camera", theme.MediaPhotoIcon(), func() {
		// TODO
	})
	buttonScreenCapture := widget.NewButtonWithIcon("Scan the screen", theme.ViewFullScreenIcon(),
		func() {
			// getting mouse location
			mouseX, mouseY := robotgo.Location()

			// getting display
			displayIndex := 0
			numDisplays := screenshot.NumActiveDisplays()
			for i := 0; i < numDisplays; i++ {
				bounds := screenshot.GetDisplayBounds(i)
				// Check if mouse is within this monitor's rectangle
				if mouseX >= bounds.Min.X && mouseX <= bounds.Max.X &&
					mouseY >= bounds.Min.Y && mouseY <= bounds.Max.Y {
					displayIndex = i
					break
				}
			}

			// prep screenshot
			bounds := screenshot.GetDisplayBounds(displayIndex)
			imagetopass, err := screenshot.CaptureRect(bounds)
			if err != nil {
				fmt.Println("failed to prep background:", err)
				return
			}
			showCropWindow(mainWindow, imagetopass, func(img image.Image) {
				if img != nil {
					showResultWindow(img)
					bmp, _ := gozxing.NewBinaryBitmapFromImage(img)
					result, err := globalQRReader.Decode(bmp, nil)
					if err == nil {
						fmt.Println(result.String(), result.GetBarcodeFormat(), result.GetTimestamp())
					} else if _, ok := err.(gozxing.NotFoundException); ok{
						fmt.Println("qrcode scanned failed, qrcode not found", err)
					} else {
						fmt.Println("qrcode scanned failed", err)
					}
				}
			},
				func() {
					fmt.Println("screenshot canceled")
				})
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

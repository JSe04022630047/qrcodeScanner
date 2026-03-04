package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/ncruces/zenity"
)

func showCreateQRWindow(_ fyne.Window) {
	formats := map[string]gozxing.BarcodeFormat{
		"QR Code":     gozxing.BarcodeFormat_QR_CODE,
		"Data Matrix": gozxing.BarcodeFormat_DATA_MATRIX,
		"Code 128":    gozxing.BarcodeFormat_CODE_128,
		"Code 39":     gozxing.BarcodeFormat_CODE_39,
		"EAN 13":      gozxing.BarcodeFormat_EAN_13,
	}

	var options []string
	for k := range formats {
		options = append(options, k)
	}
	sort.Strings(options) // to lazy to make sorting algro, I hope this is good enough for small subset

	createQRWindow := fyne.CurrentApp().NewWindow("Result")

	var currentImg image.Image
	selector := widget.NewSelect(options, nil)
	selector.SetSelected("QR Code")

	// MultiLine Entry allows for "Enter" to create new lines
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Enter text here...")
	
	// Create a scroll container for the input so it doesn't break the layout
	inputScroll := container.NewVScroll(input)
	inputScroll.SetMinSize(fyne.NewSize(0, 150))

	imgDisplay := canvas.NewImageFromImage(nil)
	imgDisplay.FillMode = canvas.ImageFillContain
	imgDisplay.SetMinSize(fyne.NewSquareSize(250))

	errorLabel := widget.NewLabel("")
	errorLabel.Wrapping = fyne.TextWrapWord
	errorLabel.Alignment = fyne.TextAlignCenter
	// use a red color for the error text
	errorLabel.TextStyle = fyne.TextStyle{Monospace: true}
	errorContainer := container.NewGridWrap(fyne.NewSize(300, 300), errorLabel)

	// inititally hidden
	imgDisplay.Hide()
	errorLabel.Hide()

	buttonExport := widget.NewButton("Export PNG", func() {
		if currentImg == nil { return }
		path, _ := zenity.SelectFileSave(
			zenity.Title("Save code: "+selector.Selected),
			zenity.Filename("code.png"),
			zenity.FileFilter{Name: "PNG Image", Patterns: []string{"*.png"}},
		)
		if path == "" { return }

		f, _ := os.Create(path)
		defer f.Close()
		png.Encode(f, currentImg)
	})
	buttonExport.Disable()

	buttonGenerate := widget.NewButton("Generate", func() {
		text := input.Text
		if text == "" {
			return
		}

		// format that user has selected
		format := formats[selector.Selected]
		bitMatrix, err := encodeBasedOnFormatWithText(text, format)
		if err != nil {
			fmt.Println("Encoding failed:", err)
			currentImg = nil
			errorLabel.SetText(fmt.Sprintf("❌ Error: %v", err))
			buttonExport.Disable()
			errorLabel.Show()
			imgDisplay.Hide()
		} else {
			currentImg = bitMatrix
			imgDisplay.Image = bitMatrix
			buttonExport.Enable()
			imgDisplay.Show()
			errorLabel.Hide()
			addEntry(scannedCode{Text: text, Timestamp: time.Now().Unix(), Format: int(format)})
			if HistoryWindow != nil{
				HistoryWindow.SetContent(generateButtonsWithScroll(HistoryWindow))
			}
		}

		imgDisplay.Refresh()
		errorLabel.Refresh()
	})
	buttonGenerate.Importance = widget.HighImportance

	// layout
	resultStack := container.NewStack(imgDisplay, errorContainer)
	resultContainer := container.NewCenter(resultStack)
	content := container.NewVBox(
		widget.NewLabel("1. Select Format"),
		selector,
		widget.NewLabel("2. Input Data"),
		inputScroll,
		container.NewGridWithColumns(2, buttonExport, buttonGenerate),
		resultContainer,
	)

	createQRWindow.SetContent(container.NewPadded(content))
	createQRWindow.Resize(fyne.NewSize(450, 600))
	createQRWindow.Show()
}

// bascially extension of func encodeBasedOnFormat() but now actually accepts normal string instead of result from another QRCode
func encodeBasedOnFormatWithText[T gozxing.BarcodeFormat | int](st string, format T) (image.Image, error){
	var encoder gozxing.Writer
	var actuallyFormat gozxing.BarcodeFormat
	switch v := any(format).(type) {
	case int:
		actuallyFormat = gozxing.BarcodeFormat(v)
	case gozxing.BarcodeFormat:
		actuallyFormat = v
	}
	if !isEncodable(actuallyFormat) {return nil, errors.New("unsupported encoding format: " + actuallyFormat.String())} // cannot encode.
	switch actuallyFormat {
	case gozxing.BarcodeFormat_QR_CODE:
		encoder = qrcode.NewQRCodeWriter()
	case gozxing.BarcodeFormat_DATA_MATRIX:
		encoder = datamatrix.NewDataMatrixWriter()
	case gozxing.BarcodeFormat_UPC_A:
		encoder = oned.NewUPCAWriter()
	case gozxing.BarcodeFormat_UPC_E:
		encoder = oned.NewUPCEWriter()
	case gozxing.BarcodeFormat_EAN_8:
		encoder = oned.NewEAN8Writer()
	case gozxing.BarcodeFormat_EAN_13:
		encoder = oned.NewEAN13Writer()
	case gozxing.BarcodeFormat_CODE_39:
		encoder = oned.NewCode39Writer()
	case gozxing.BarcodeFormat_CODE_93:
		encoder = oned.NewCode93Writer()
	case gozxing.BarcodeFormat_CODE_128:
		encoder = oned.NewCode128Writer()
	case gozxing.BarcodeFormat_CODABAR:
		encoder = oned.NewCodaBarWriter()
	case gozxing.BarcodeFormat_ITF:
		encoder = oned.NewITFWriter()
	default:
		return nil, errors.New("unsupported encoding format: " + actuallyFormat.String())
	}

	matrix, errMatrix := encoder.Encode(st, actuallyFormat, 256,256,nil)
	if errMatrix != nil {
		return nil, errMatrix
	}

	return matrix, nil
}
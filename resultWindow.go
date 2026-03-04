package main

import (
	"errors"
	"image"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/datamatrix"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/pkg/browser"
)

// new entry widget, appearing as normal but content can't change
type ReadOnlyEntry struct {
	widget.Entry
}

// NewReadOnlyEntry creates a version that looks normal but ignores input
func NewReadOnlyEntry() *ReadOnlyEntry {
	e := &ReadOnlyEntry{}
	e.ExtendBaseWidget(e) // Required for custom widgets
	return e
}

// TypedRune intercepts character input and does nothing
func (e *ReadOnlyEntry) TypedRune(r rune) {}

// TypedKey intercepts special keys (Backspace, Enter, etc.) and does nothing
// We allow shortcuts (like Ctrl+C) so users can still copy the text
func (e *ReadOnlyEntry) TypedKey(k *fyne.KeyEvent) {
	if k.Name == fyne.KeyName(fyne.KeyboardShortcut.ShortcutName(&fyne.ShortcutCopy{})) {
		e.Entry.TypedKey(k)
	}
}

// DoubleTap/SecondaryTap/Paste can also be overridden if you want to block context menus
func (e *ReadOnlyEntry) Paste(s string) {}

func showResultWindow(result scannedCode) {
	resultWin := fyne.CurrentApp().NewWindow("Result")
	
	titleLabel := widget.NewLabel("Result")
	// content of the qrcode
	stringEntry := NewReadOnlyEntry()
	formatEntry := NewReadOnlyEntry()
	timestampEntry := NewReadOnlyEntry()

	stringEntry.SetText(result.Text)
	formatEntry.SetText(gozxing.BarcodeFormat(result.Format).String()) // hack yes but it works so I don't bother
	tm := time.Unix(result.Timestamp, 0) // we already converted the milisec to normal seconds
	timestampEntry.SetText(tm.Format(time.RFC3339))

	
	// image stuff here, I wish this works better
	labelNotEncodeable := widget.NewLabel("")

	img, err := encodeBasedOnFormat(result) 
	var qrCanvas *canvas.Image
	if err != nil {
		labelNotEncodeable.SetText("No encoder format found for "+formatEntry.Text)
		labelNotEncodeable.SizeName = theme.SizeNameHeadingText
	} else {
		qrCanvas = canvas.NewImageFromImage(img)
		qrCanvas.FillMode = canvas.ImageFillContain
		qrCanvas.SetMinSize(fyne.NewSize(200, 200))
	}

	form := widget.NewForm(
		widget.NewFormItem("Text", stringEntry),
		widget.NewFormItem("Format", formatEntry),
		widget.NewFormItem("Timestamp", timestampEntry),
	)

	// buttons
	openUrlButton := widget.NewButtonWithIcon("Open in web browser", theme.SearchIcon(), func() {
		browser.OpenURL(result.Text)
	})

	if !isValidURI(result.Text, []string{"http", "https"}){
		openUrlButton.Hide()
	}

	copyTextButton := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		fyne.CurrentApp().Clipboard().SetContent(result.Text)
	})

	closeButton := widget.NewButtonWithIcon("Close", theme.WindowCloseIcon(), func() {
		resultWin.Close()
	})
	closeButton.Importance = widget.HighImportance

	buttonStrip := container.NewHBox(
		openUrlButton,
		layout.NewSpacer(),
		copyTextButton,
		closeButton,
	)

	mainContent := container.NewVBox(titleLabel,form,labelNotEncodeable,qrCanvas)
	mainContentWrapper := container.NewVScroll(mainContent)
	mainContentWrapper.SetMinSize(mainContent.MinSize())
	mainContentWithButton := container.NewVBox(mainContentWrapper, buttonStrip)
	resultWin.SetContent(mainContentWithButton)
	resultWin.Resize(fyne.NewSize(400, mainContent.MinSize().Height+buttonStrip.MinSize().Height))
	resultWin.Show()
}

// history window stuff
type SubtitleButton struct { // button with new subtitle, the normal button doesn't have subtitle option.
	widget.BaseWidget
	Title    string
	Subtitle string
	OnTapped func()
}

func NewSubtitleButton(title, subtitle string, tapped func()) *SubtitleButton {
	b := &SubtitleButton{Title: title, Subtitle: subtitle, OnTapped: tapped}
	b.ExtendBaseWidget(b)
	return b
}

func (b *SubtitleButton) CreateRenderer() fyne.WidgetRenderer {
	title := widget.NewLabel(b.Title)
	title.TextStyle.Bold = true
	title.SizeName = theme.SizeNameHeadingText
	title.Truncation = fyne.TextTruncateEllipsis
	
	sub := widget.NewLabel(b.Subtitle)
	sub.SizeName = theme.SizeNameText
	sub.Truncation = fyne.TextTruncateEllipsis
	
	// Use a container to stack them
	content := container.NewVBox(title, sub)
	
	// Wrap in a standard button to get the hover/click effects
	btn := widget.NewButton("", b.OnTapped)

	return widget.NewSimpleRenderer(container.NewStack(btn, content))
}

func showHistoryWindow() {
	histroyWin := fyne.CurrentApp().NewWindow("History")
	buttonList := container.NewVBox()

	for i := 0; i < len(scannedHistory); i++ {
		index := i // capture index
		var item scannedCode = scannedHistory[index] 
		button := NewSubtitleButton(item.Text, strconv.FormatInt(item.Timestamp, 10), func ()  {
			showResultWindow(item)
		})
		buttonList.Add(button)
	}
	scroll := container.NewVScroll(buttonList)
	histroyWin.SetContent(scroll)
	histroyWin.Resize(fyne.NewSquareSize(600))
	histroyWin.CenterOnScreen()
	histroyWin.Show()
}

func isEncodable[T int | gozxing.BarcodeFormat](format T) bool {
	switch format{
	case 1:
		return true
	case 2:
		return true
	case 3:
		return true
	case 4:
		return true
	case 5:
		return true
	case 6:
		return true
	case 7:
		return true
	case 8:
		return true
	case 11:
		return true
	case 12:
		return true
	case 14:
		return true
	case 15:
		return true
	default:
		return false
	}
}

// this is memory intensive but I need a multi writer essentially, I wish there is a better way to do this but I am not smart enough and this is due in a week, I should finish this yesterday.
func encodeBasedOnFormat(result scannedCode) (image.Image, error) {
	var encoder gozxing.Writer
	format := gozxing.BarcodeFormat(result.Format)
	if !isEncodable(format) {return nil, errors.New("unsupported encoding format: " + format.String())} // cannot encode.
	switch format {
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
		return nil, errors.New("unsupported encoding format: " + format.String())
	}

	matrix, errMatrix := encoder.Encode(result.Text, format, 256,256,nil)
	if errMatrix != nil {
		return nil, errMatrix
	}

	return matrix, nil
}

func isValidURI(input string, allowedProtocols []string) bool {
	u, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}

	// Ensure there is a scheme and a host (prevents strings like "mailto:test@test.com")
	if u.Scheme == "" || u.Host == "" {
		return false
	}

	// Check against your allowed list
	for _, proto := range allowedProtocols {
		if strings.EqualFold(u.Scheme, proto) {
			return true
		}
	}

	return false
}
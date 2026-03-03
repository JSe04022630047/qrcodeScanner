package main

import (
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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
	resultWin.Resize(fyne.NewSize(400, mainContent.MinSize().Height))
	resultWin.Show()
}

type SubtitleButton struct { // button with new subtitle
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
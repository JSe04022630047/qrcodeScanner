package main

import (
	"fmt"
	"image"
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// image cropper, AI magic (vibe coded) I don't know fyne backend stuff. probably going to learn when this is all over.
// TODO refactor after done with school and coop
type handle struct {
	widget.BaseWidget
	circle *canvas.Circle
	onDrag func(fyne.Delta)
}

func newHandle(onDrag func(fyne.Delta)) *handle {
	h := &handle{onDrag: onDrag}
	h.circle = canvas.NewCircle(theme.PrimaryColor())
	h.circle.StrokeWidth = 1
	h.circle.StrokeColor = theme.ForegroundColor()
	h.ExtendBaseWidget(h)
	return h
}

func (h *handle) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.circle)
}

func (h *handle) Dragged(e *fyne.DragEvent) { h.onDrag(e.Dragged) }
func (h *handle) DragEnd() {}

// The Main Selector Widget
type cropSelector struct {
	widget.BaseWidget
	rect   *canvas.Rectangle
	handle *handle
}

func newCropSelector() *cropSelector {
	cs := &cropSelector{rect: canvas.NewRectangle(image.Transparent)}
	cs.rect.StrokeWidth = 2
	cs.rect.StrokeColor = theme.PrimaryColor()

	cs.handle = newHandle(func(delta fyne.Delta) {
		newSize := cs.Size().Add(delta)
		if newSize.Width > 30 && newSize.Height > 30 {
			cs.Resize(newSize)
		}
	})

	cs.ExtendBaseWidget(cs)
	return cs
}

func (cs *cropSelector) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewWithoutLayout(cs.rect, cs.handle))
}

func (cs *cropSelector) Resize(s fyne.Size) {
	cs.BaseWidget.Resize(s)
	cs.rect.Resize(s)
	hSize := fyne.NewSize(20, 20)
	cs.handle.Resize(hSize)
	cs.handle.Move(fyne.NewPos(s.Width-hSize.Width/2, s.Height-hSize.Height/2))
}

func (cs *cropSelector) GetCropRect(imgWidget *canvas.Image, originalImg image.Image) image.Rectangle {
    // 1. Calculate the scale ratio
    // Fyne uses "Points", the image uses "Pixels"
    viewSize := imgWidget.Size()
    imgSize := originalImg.Bounds().Size()
    
    ratioX := float32(imgSize.X) / viewSize.Width
    ratioY := float32(imgSize.Y) / viewSize.Height

    // 2. Calculate position relative to the image widget
    // Subtracting imgWidget.Position() handles cases where the image is centered
    relPos := cs.Position().Subtract(imgWidget.Position())
    
    x1 := int(relPos.X * ratioX)
    y1 := int(relPos.Y * ratioY)
    x2 := x1 + int(cs.Size().Width * ratioX)
    y2 := y1 + int(cs.Size().Height * ratioY)

    return image.Rect(x1, y1, x2, y2)
}

func (cs *cropSelector) Dragged(e *fyne.DragEvent) {
	cs.Move(cs.Position().Add(e.Dragged))
}

// Main show
func showCropWindow(
	_ fyne.Window, // skipped for now
	src image.Image,
	onCrop func(image.Image), // Success callback
	onCancel func(), // Cancel callback
) {
	cropWin := fyne.CurrentApp().NewWindow("Crop Image")

	// 1. The Image and Draggable Selector
	imgWidget := canvas.NewImageFromImage(src)
	imgWidget.FillMode = canvas.ImageFillContain
	selector := newCropSelector()
	selector.Resize(fyne.NewSize(150, 150))

	centerStack := container.NewStack(
		imgWidget,
		container.NewWithoutLayout(selector),
	)

	// 2. The Buttons
	cancelBtn := widget.NewButton("Cancel", func() {
		cropWin.Close()
	})

	confirmBtn := widget.NewButtonWithIcon("Crop & Scan", theme.ConfirmIcon(), func() {
		// 1. Get the pixel coordinates from our widget
		cropRect := selector.GetCropRect(imgWidget, src)

		// 2. Perform the actual crop
		// We create a new RGBA image and draw the section of 'src' onto it
		subImg := image.NewRGBA(image.Rect(0, 0, cropRect.Dx(), cropRect.Dy()))
		draw.Draw(subImg, subImg.Bounds(), src, cropRect.Min, draw.Src)

		// 3. Send the cropped image back to the caller
		onCrop(subImg)

		cropWin.Close()
	})
	confirmBtn.Importance = widget.HighImportance

	// 3. Layout: Stick to Bottom Right
	// Spacer pushes everything after it to the right
	buttonStrip := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		confirmBtn,
	)

	// Border layout: Center is the image, Bottom is the button strip
	mainLayout := container.NewBorder(nil, buttonStrip, nil, nil, centerStack)

	// listens to window close
	cropWin.SetOnClosed(func() {
		fmt.Println("Closing window or pressing cancel")
		onCancel()
	})
	cropWin.SetContent(mainLayout)
	cropWin.Resize(fyne.NewSize(600, 500))
	cropWin.Show()

}

// end crop stuff

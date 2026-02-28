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

// handle for the cropper
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
	container fyne.CanvasObject
}

func newCropSelector(bounds fyne.CanvasObject) *cropSelector {
	cs := &cropSelector{rect: canvas.NewRectangle(image.Transparent), container: bounds}
	cs.rect.StrokeWidth = 2
	cs.rect.StrokeColor = theme.Color(theme.ColorNamePrimary)

	cs.handle = newHandle(func(delta fyne.Delta) {
    parent := cs.container
    if parent == nil { return }
    
    parentSize := parent.Size()
    currentPos := cs.Position()
    
    // Calculate intended new size
    newSize := cs.Size().Add(delta)

    // Clamp Width: cannot be smaller than 30 or larger than (ParentWidth - CurrentX)
    if newSize.Width < 30 {
        newSize.Width = 30
    } else if currentPos.X + newSize.Width > parentSize.Width {
        newSize.Width = parentSize.Width - currentPos.X
    }

    // Clamp Height: cannot be smaller than 30 or larger than (ParentHeight - CurrentY)
    if newSize.Height < 30 {
        newSize.Height = 30
    } else if currentPos.Y + newSize.Height > parentSize.Height {
        newSize.Height = parentSize.Height - currentPos.Y
    }

    cs.Resize(newSize)
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
    if cs.container == nil {
        return
    }

    // Get the size of the image widget
    boundarySize := cs.container.Size()
    currentSize := cs.Size()
    
    // Calculate new position
    newPos := cs.Position().Add(e.Dragged)

    // Clamp X to [0, ImageWidth - SelectorWidth]
    if newPos.X < 0 {
        newPos.X = 0
    } else if newPos.X+currentSize.Width > boundarySize.Width {
        newPos.X = boundarySize.Width - currentSize.Width
    }

    // Clamp Y to [0, ImageHeight - SelectorHeight]
    if newPos.Y < 0 {
        newPos.Y = 0
    } else if newPos.Y+currentSize.Height > boundarySize.Height {
        newPos.Y = boundarySize.Height - currentSize.Height
    }

    cs.Move(newPos)
}

func (cs *cropSelector) DragEnd() {
	fmt.Println("Finished moving to:", cs.Position())
}

// Main show
func showCropWindow(
	_ fyne.Window, // skipped for now
	src image.Image,
	onCrop func(image.Image), // Success callback
	onCancel func(), // Cancel callback
) {
	var confirmed bool = false
	cropWin := fyne.CurrentApp().NewWindow("Crop Image")

	// The Image 
	imgWidget := canvas.NewImageFromImage(src)
	imgWidget.FillMode = canvas.ImageFillStretch
	imgWidget.SetMinSize(fyne.NewSize(500,500))



	// selector
	selector := newCropSelector(imgWidget)
	selector.Resize(fyne.NewSize(100, 100))

	centerStack := container.NewStack(
		imgWidget,
		container.NewWithoutLayout(selector),
	)

	// 2. The Buttons
	cancelButton := widget.NewButton("Cancel", func() {
		cropWin.Close()
	})

	confirmButton := widget.NewButtonWithIcon("Crop & Scan", theme.ConfirmIcon(), func() {
		confirmed = true
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
	confirmButton.Importance = widget.HighImportance

	// 3. Layout: Stick to Bottom Right
	// Spacer pushes everything after it to the right
	buttonStrip := container.NewHBox(
		layout.NewSpacer(),
		cancelButton,
		confirmButton,
	)

	// Border layout: Center is the image, Bottom is the button strip
	mainLayout := container.NewBorder(nil, buttonStrip, nil, nil, centerStack)

	// listens to window close
	cropWin.SetOnClosed(func() {
		if !confirmed {
			fmt.Println("Closing window or pressing cancel")
			onCancel()
		}
	})
	cropWin.SetContent(mainLayout)
	cropWin.Resize(fyne.NewSize(600, 500))
	cropWin.Show()

}

// end crop stuff

// test
func showResultWindow(cropped image.Image) {
	// 1. Create the new window
	resWin := fyne.CurrentApp().NewWindow("Cropped Result")
	
	// 2. Convert standard image.Image to a Fyne Canvas Image
	uiImg := canvas.NewImageFromImage(cropped)
	
	// Use Original mode so it doesn't stretch and looks sharp
	uiImg.FillMode = canvas.ImageFillOriginal
	
	// 3. Add a "Close" button for convenience
	closeBtn := widget.NewButton("Close", func() {
		resWin.Close()
	})

	// 4. Layout the window
	// Using NewCenter ensures the image stays in the middle
	content := container.NewBorder(nil, closeBtn, nil, nil, container.NewCenter(uiImg))
	
	resWin.SetContent(content)
	resWin.Resize(fyne.NewSize(float32(cropped.Bounds().Dx())+40, float32(cropped.Bounds().Dy())+80))
	resWin.Show()
}
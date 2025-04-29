package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	window := myApp.NewWindow("FHIR Guard")

	// Create a label widget
	label := widget.NewLabel("Welcome to FHIR Guard Desktop")

	// Create a button widget
	button := widget.NewButton("Click Me", func() {
		label.SetText("Button clicked!")
	})

	// Create a container with the widgets
	content := container.NewVBox(
		label,
		button,
	)

	// Set the window content
	window.SetContent(content)

	// Set the window size
	window.Resize(fyne.NewSize(400, 300))

	// Show and run the application
	window.ShowAndRun()
} 
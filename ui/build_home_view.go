package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// BuildHomeView restituisce la vista Home.
func BuildHomeView() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Astro-Lair Go Edition", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	sub := widget.NewLabel("Versione di prova in Go + Fyne.\nQui potrai avere riepiloghi rapidi della serata, setup strumenti, ecc.")
	sub.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		layout.NewSpacer(),
		title,
		sub,
		layout.NewSpacer(),
	)
}

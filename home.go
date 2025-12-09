package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func buildHomeView() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Astro-Lair Go Edition", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	sub := widget.NewLabel("Versione di prova in Go + Fyne.\nQui potrai avere riepiloghi rapidi della serata, setup strumenti, ecc.")
	sub.Wrapping = fyne.TextWrapWord

	tabs := container.NewAppTabs(
		container.NewTabItem("Home", buildHomeView()),
		container.NewTabItem("Moon", buildMoonView()),
		container.NewTabItem("Weather", buildWeatherView()),
		container.NewTabItem("SpaceWeather", buildSpaceWeatherView()),
		container.NewTabItem("SOHO", buildSohoView()),
		container.NewTabItem("Satellites", satView),
		container.NewTabItem("Targets", targetsView.Widget()),
		container.NewTabItem("Tools", buildToolsView()),
	)
	return container.NewVBox(
		layout.NewSpacer(),
		title,
		sub,
		layout.NewSpacer(),
	)
}

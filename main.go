package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.NewWithID("com.cr4sh.astrolair.go")
	w := a.NewWindow("Astro-Lair (Go Edition)")

	// Modalità fullscreen (su Android nasconde la status bar)
	w.SetFullScreen(true)

	title := canvas.NewText("Astro-Lair", theme.ForegroundColor())
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("Go Edition", color.NRGBA{R: 150, G: 200, B: 255, A: 255})
	subtitle.TextSize = 12

	titleBox := container.NewVBox(title, subtitle)

	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		showEquipmentDialog(w)
	})

	topBar := container.NewBorder(
		nil,
		nil,
		titleBox,
		settingsBtn,
	)

	targetsView := NewTargetsView(buildTargetsCatalog())

	// Satellites ora ritorna anche una funzione di refresh
	satView, satRefresh := buildSatellitesView()

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
	tabs.SetTabLocation(container.TabLocationBottom)

	// Quando entri nella tab Satellites → simula "Adesso"
	tabs.OnChanged = func(ti *container.TabItem) {
		if ti.Text == "Satellites" {
			satRefresh()
		}
	}

	tabs.SetTabLocation(container.TabLocationBottom)

	root := container.NewBorder(
		topBar,
		nil,
		nil,
		nil,
		tabs,
	)

	w.SetContent(root)
	w.Resize(fyne.NewSize(480, 800))
	w.ShowAndRun()
}

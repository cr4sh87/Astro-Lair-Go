package main

import (
	"image/color"

	"github.com/cr4sh87/astro-lair-go/services"
	"github.com/cr4sh87/astro-lair-go/ui"

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

	// 🔭 Prima cosa: aggiorna il catalogo DSO dal repo GitHub
	// Se fallisce, buildTargetsCatalog() userà comunque il file esistente (se c'è).
	services.UpdateDSOCatalogFromGitHub()

	// Inizializza la configurazione dell'equipaggio nel package ui
	ui.SetEquipmentConfig(ui.NewDefaultEquipmentConfig())

	// Inizializza il provider dei sprite della luna
	ui.InitMoonProviderForUI()

	// Modalità fullscreen (su Android nasconde la status bar)
	//w.SetFullScreen(true)

	title := canvas.NewText("Astro-Lair", theme.ForegroundColor())
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("Go Edition", color.NRGBA{R: 150, G: 200, B: 255, A: 255})
	subtitle.TextSize = 12

	titleBox := container.NewVBox(title, subtitle)

	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		ui.ShowEquipmentDialog(w)
	})

	topBar := container.NewBorder(
		nil,
		nil,
		titleBox,
		settingsBtn,
	)

	// Qui buildTargetsCatalog() leggerà il file aggiornato in catalog/dso_catalog.json
	targetsView := ui.BuildTargetsViewPublic(ui.BuildCatalog())

	// Satellites ora ritorna anche una funzione di refresh
	satView, satRefresh := ui.BuildSatellitesViewPublic()

	tabs := container.NewAppTabs(
		container.NewTabItem("Home", ui.BuildHomeView()),
		container.NewTabItem("Moon", ui.BuildMoonViewPublic()),
		container.NewTabItem("Weather", ui.BuildWeatherViewPublic()),
		container.NewTabItem("SpaceWeather", ui.BuildSpaceWeatherView()),
		container.NewTabItem("SOHO", ui.BuildSohoView()),
		container.NewTabItem("Satellites", satView),
		container.NewTabItem("Targets", targetsView.Widget()),
		container.NewTabItem("Tools", ui.BuildToolsView()),
	)
	tabs.SetTabLocation(container.TabLocationBottom)

	// Quando entri nella tab Satellites → simula "Adesso"
	tabs.OnChanged = func(ti *container.TabItem) {
		if ti.Text == "Satellites" {
			satRefresh()
		}
	}

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

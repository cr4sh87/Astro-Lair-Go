package ui

import (
	"image/color"

	"github.com/cr4sh87/astro-lair-go/services"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// URL NOAA
const (
	auroraNorthURL = "https://services.swpc.noaa.gov/images/aurora-forecast-northern-hemisphere.jpg"
	auroraSouthURL = "https://services.swpc.noaa.gov/images/aurora-forecast-southern-hemisphere.jpg"

	spaceWeatherOverviewGIF = "https://services.swpc.noaa.gov/images/swx-overview-large.gif"
)

// BuildSpaceWeatherView — Space Weather UI
func BuildSpaceWeatherView() fyne.CanvasObject {
	return buildSpaceWeatherView()
}

func buildSpaceWeatherView() fyne.CanvasObject {
	// Titolo
	title := widget.NewLabelWithStyle(
		"Space Weather – Aurora & Overview",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	subtitle := widget.NewLabel(
		"Immagini in tempo quasi-reale dall'NOAA Space Weather Prediction Center.\n" +
			"Aurora boreale/australe e panoramica animata del meteo spaziale.",
	)

	statusLabel := widget.NewLabel("")

	// Aurora Nord / Sud
	northImg := canvas.NewImageFromResource(nil)
	northImg.FillMode = canvas.ImageFillContain
	northImg.SetMinSize(fyne.NewSize(260, 260))

	southImg := canvas.NewImageFromResource(nil)
	southImg.FillMode = canvas.ImageFillContain
	southImg.SetMinSize(fyne.NewSize(260, 260))

	btnLoadNorth := widget.NewButton("Aggiorna Aurora Nord", func() {
		services.LoadRemoteImage(auroraNorthURL, "Aurora Nord", northImg, statusLabel)
	})

	btnLoadSouth := widget.NewButton("Aggiorna Aurora Sud", func() {
		services.LoadRemoteImage(auroraSouthURL, "Aurora Sud", southImg, statusLabel)
	})

	auroraNorthBox := container.NewVBox(
		widget.NewLabelWithStyle("Aurora – Emisfero Nord", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		northImg,
		btnLoadNorth,
	)

	auroraSouthBox := container.NewVBox(
		widget.NewLabelWithStyle("Aurora – Emisfero Sud", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		southImg,
		btnLoadSouth,
	)

	auroraRow := container.NewAdaptiveGrid(2,
		auroraNorthBox,
		auroraSouthBox,
	)

	// Overview animata (GIF NOAA)
	overviewTitle := widget.NewLabelWithStyle(
		"Space Weather Overview (GIF NOAA)",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	overviewStatus := widget.NewLabel("")

	// Oggetto fittizio usato come parent per la popup
	overviewPlaceholder := canvas.NewRectangle(color.Transparent)

	btnShowOverview := widget.NewButton("Mostra GIF NOAA", func() {
		services.LoadRemoteAnimation(spaceWeatherOverviewGIF, "Space Weather Overview", overviewStatus, func(frames []fyne.Resource) {
			if len(frames) == 0 {
				overviewStatus.SetText("Nessun dato GIF disponibile.")
				return
			}
			overviewStatus.SetText("GIF NOAA pronta.")
			showResourcesDialog(overviewPlaceholder, frames, "Space Weather Overview – NOAA SWPC")
		})
	})

	overviewBox := container.NewVBox(
		overviewTitle,
		btnShowOverview,
		overviewStatus,
	)

	credits := widget.NewLabel("Dati e immagini: NOAA / NWS Space Weather Prediction Center")

	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		auroraRow,
		widget.NewSeparator(),
		overviewBox,
		widget.NewSeparator(),
		statusLabel,
		credits,
	)

	// Caricamento iniziale delle mappe aurorali
	services.LoadRemoteImage(auroraNorthURL, "Aurora Nord", northImg, statusLabel)
	services.LoadRemoteImage(auroraSouthURL, "Aurora Sud", southImg, statusLabel)

	return container.NewVScroll(content)
}

package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cr4sh87/astro-lair-go/services"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// =============================================================
//  STORAGE POSIZIONE METEO (Preferences)
// =============================================================

const (
	prefWeatherLat = "weather.lat"
	prefWeatherLon = "weather.lon"
)

// loadStoredCoords carica lat/lon se presenti nelle Preferences
func loadStoredCoords() (float64, float64, bool) {
	app := fyne.CurrentApp()
	if app == nil {
		return 0, 0, false
	}
	p := app.Preferences()

	latStr := p.String(prefWeatherLat)
	lonStr := p.String(prefWeatherLon)
	if latStr == "" || lonStr == "" {
		return 0, 0, false
	}

	lat, err1 := strconv.ParseFloat(latStr, 64)
	lon, err2 := strconv.ParseFloat(lonStr, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}

	return lat, lon, true
}

// saveStoredCoords salva la posizione meteo scelta dall'utente
func saveStoredCoords(lat, lon float64) {
	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	p := app.Preferences()
	p.SetString(prefWeatherLat, fmt.Sprintf("%f", lat))
	p.SetString(prefWeatherLon, fmt.Sprintf("%f", lon))
}

// =============================================================
//  WEATHER VIEW
// =============================================================

func buildWeatherView() fyne.CanvasObject {
	// valori di default / fallback
	latDefault := 37.65
	lonDefault := 15.17

	// se esiste una posizione salvata → usala
	if lat, lon, ok := loadStoredCoords(); ok {
		latDefault = lat
		lonDefault = lon
	}

	latEntry := widget.NewEntry()
	latEntry.SetText(fmt.Sprintf("%.4f", latDefault))

	lonEntry := widget.NewEntry()
	lonEntry.SetText(fmt.Sprintf("%.4f", lonDefault))

	result := widget.NewLabel("Inserisci coordinate o usa la posizione del dispositivo, poi premi \"Aggiorna Meteo\".")
	result.Wrapping = fyne.TextWrapWord

	loading := widget.NewProgressBarInfinite()
	loading.Hide()

	// funzione condivisa che scarica e aggiorna il meteo
	updateWeather := func(lat, lon float64) {
		// salva per uso futuro
		saveStoredCoords(lat, lon)

		loading.Show()
		result.SetText("Scarico previsioni…")

		go func() {
			meteo, err := services.FetchWeather(lat, lon)
			if err != nil {
				fyne.Do(func() {
					loading.Hide()
					result.SetText("Errore durante il download del meteo:\n" + err.Error())
				})
				return
			}

			fyne.Do(func() {
				loading.Hide()

				if len(meteo.Hourly.Time) == 0 ||
					len(meteo.Hourly.CloudCover) == 0 ||
					len(meteo.Hourly.Humidity) == 0 ||
					len(meteo.Hourly.WindSpeed) == 0 {
					result.SetText("Dati meteo non disponibili.")
					return
				}

				cc := meteo.Hourly.CloudCover[0]
				hum := meteo.Hourly.Humidity[0]
				wnd := meteo.Hourly.WindSpeed[0]

				text := fmt.Sprintf(
					"🌤️ Copertura nuvolosa: %.0f%%\n💧 Umidità: %.0f%%\n🌬️ Vento: %.1f m/s",
					cc, hum, wnd,
				)

				var quality string
				switch {
				case cc < 20 && wnd < 4:
					quality = "🔵 Cielo ottimo per foto"
				case cc < 40 && wnd < 6:
					quality = "🟢 Cielo buono"
				case cc < 70:
					quality = "🟡 Cielo mediocre"
				default:
					quality = "🔴 Cielo poco adatto"
				}

				result.SetText(text + "\n\n" + quality)
			})
		}()
	}

	updateBtn := widget.NewButton("Aggiorna Meteo", func() {
		lat, err1 := strconv.ParseFloat(strings.ReplaceAll(latEntry.Text, ",", "."), 64)
		lon, err2 := strconv.ParseFloat(strings.ReplaceAll(lonEntry.Text, ",", "."), 64)
		if err1 != nil || err2 != nil {
			result.SetText("Coordinate non valide.")
			return
		}
		updateWeather(lat, lon)
	})

	// 🔘 Pulsante che forza la richiesta di posizione (via IP)
	useDeviceLocationBtn := widget.NewButton("Usa posizione dispositivo", func() {
		loading.Show()
		result.SetText("Rilevo la posizione del dispositivo…")

		go func() {
			loc, err := services.AutoLocateDevice()
			if err != nil {
				fyne.Do(func() {
					loading.Hide()
					result.SetText("Impossibile rilevare la posizione:\n" + err.Error())
				})
				return
			}

			fyne.Do(func() {
				// aggiorna campi
				latEntry.SetText(fmt.Sprintf("%.4f", loc.Lat))
				lonEntry.SetText(fmt.Sprintf("%.4f", loc.Lon))

				// e aggiorna direttamente il meteo
				updateWeather(loc.Lat, loc.Lon)
			})
		}()
	})

	form := widget.NewForm(
		widget.NewFormItem("Latitudine", latEntry),
		widget.NewFormItem("Longitudine", lonEntry),
	)

	header := container.NewVBox(
		widget.NewLabelWithStyle("Meteo Astronomico", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		useDeviceLocationBtn,
	)

	return container.NewVBox(
		header,
		widget.NewSeparator(),
		form,
		updateBtn,
		loading,
		widget.NewSeparator(),
		result,
	)
}

// BuildWeatherView ritorna la vista meteo
func BuildWeatherView() fyne.CanvasObject {
	return buildWeatherView()
}

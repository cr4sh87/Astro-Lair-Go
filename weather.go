package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

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

// carica lat/lon se presenti nelle Preferences
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

// salva la posizione meteo scelta dall’utente
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
//  AUTO-LOCALIZZAZIONE VIA IP (network-based)
// =============================================================

type DeviceLocation struct {
	Lat float64
	Lon float64
}

type ipAPIResponse struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// autoLocateDevice prova a stimare la posizione dal IP pubblico
// usando il servizio ip-api.com.
// NON è GPS, ma su Android di solito ti dà una posizione abbastanza vicina.
func autoLocateDevice() (*DeviceLocation, error) {
	const url = "http://ip-api.com/json/"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("errore richiesta IP-geo: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("errore lettura risposta IP-geo: %w", err)
	}

	var data ipAPIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("errore decode JSON IP-geo: %w", err)
	}

	if data.Status != "success" {
		if data.Message != "" {
			return nil, fmt.Errorf("IP-geo fallita: %s", data.Message)
		}
		return nil, fmt.Errorf("IP-geo fallita con stato: %s", data.Status)
	}

	return &DeviceLocation{
		Lat: data.Lat,
		Lon: data.Lon,
	}, nil
}

// =============================================================
//  Open-Meteo Fetch
// =============================================================

type WeatherResponse struct {
	Hourly struct {
		Time       []string  `json:"time"`
		CloudCover []float64 `json:"cloud_cover"`
		Humidity   []float64 `json:"relative_humidity_2m"`
		WindSpeed  []float64 `json:"wind_speed_10m"`
	} `json:"hourly"`
}

func fetchWeather(lat, lon float64) (*WeatherResponse, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=cloud_cover,relative_humidity_2m,wind_speed_10m&forecast_days=1&timezone=auto",
		lat, lon,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wr WeatherResponse
	if err := json.Unmarshal(data, &wr); err != nil {
		return nil, err
	}

	fmt.Printf("Meteo ricevuto: time=%d cloud=%d hum=%d wind=%d\n",
		len(wr.Hourly.Time),
		len(wr.Hourly.CloudCover),
		len(wr.Hourly.Humidity),
		len(wr.Hourly.WindSpeed),
	)

	return &wr, nil
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
			meteo, err := fetchWeather(lat, lon)
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

	// 🔘 Pulsante che forz a la richiesta di posizione (via IP)
	useDeviceLocationBtn := widget.NewButton("Usa posizione dispositivo", func() {
		loading.Show()
		result.SetText("Rilevo la posizione del dispositivo…")

		go func() {
			loc, err := autoLocateDevice()
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

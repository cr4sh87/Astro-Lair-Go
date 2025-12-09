package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DeviceLocation rappresenta la posizione geografica del dispositivo
type DeviceLocation struct {
	Lat float64
	Lon float64
}

// ipAPIResponse è la risposta dal servizio ip-api.com
type ipAPIResponse struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// AutoLocateDevice prova a stimare la posizione dal IP pubblico
// usando il servizio ip-api.com.
// NON è GPS, ma su Android di solito ti dà una posizione abbastanza vicina.
func AutoLocateDevice() (*DeviceLocation, error) {
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

// WeatherResponse è la risposta dal servizio Open-Meteo
type WeatherResponse struct {
	Hourly struct {
		Time       []string  `json:"time"`
		CloudCover []float64 `json:"cloud_cover"`
		Humidity   []float64 `json:"relative_humidity_2m"`
		WindSpeed  []float64 `json:"wind_speed_10m"`
	} `json:"hourly"`
}

// FetchWeather scarica le previsioni meteo da Open-Meteo
func FetchWeather(lat, lon float64) (*WeatherResponse, error) {
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

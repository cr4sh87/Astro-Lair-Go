package ui

import (
	"github.com/cr4sh87/astro-lair-go/models"

	"fyne.io/fyne/v2"
)

// BuildTargetsView ritorna la vista targets
func BuildTargetsViewPublic(byCatalog map[string][]models.TargetObject) *TargetsView {
	return BuildTargetsView(byCatalog)
}

// BuildCatalog ritorna il catalogo dei target dal file o dal fallback
func BuildCatalog() map[string][]models.TargetObject {
	return buildTargetsCatalog()
}

// BuildSatellitesView ritorna la vista dei satelliti e un refresh function
func BuildSatellitesViewPublic() (fyne.CanvasObject, func()) {
	return buildSatellitesView()
}

// BuildMoonView ritorna la vista della luna
func BuildMoonViewPublic() fyne.CanvasObject {
	return buildMoonView()
}

// BuildWeatherView ritorna la vista del meteo
func BuildWeatherViewPublic() fyne.CanvasObject {
	return buildWeatherView()
}

package ui

import (
	"fyne.io/fyne/v2"

	"github.com/cr4sh87/astro-lair-go/services"
)

// moonSpriteProviderImpl implementa MoonSpriteProvider usando il servizio delle risorse embeddate.
type moonSpriteProviderImpl struct{}

func (m *moonSpriteProviderImpl) GetMoonSprite(idx int) fyne.Resource {
	return services.GetMoonSpriteByIndex(idx)
}

// InitMoonProviderForUI inizializza il provider della luna per il package ui.
func InitMoonProviderForUI() {
	SetMoonSpriteProvider(&moonSpriteProviderImpl{})
}

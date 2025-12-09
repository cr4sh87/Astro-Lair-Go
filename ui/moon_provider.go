package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
)

// MoonSpriteProvider è un'interfaccia per fornire gli sprite della luna
type MoonSpriteProvider interface {
	GetMoonSprite(idx int) fyne.Resource
}

var moonProvider MoonSpriteProvider

// SetMoonSpriteProvider imposta il provider per gli sprite della luna
func SetMoonSpriteProvider(provider MoonSpriteProvider) {
	moonProvider = provider
}

// GetMoonSpriteByIndex wrapper che chiama il provider
func GetMoonSpriteByIndex(idx int) fyne.Resource {
	if moonProvider == nil {
		fmt.Println("[ERROR] MoonSpriteProvider non è stato inizializzato")
		return nil
	}
	return moonProvider.GetMoonSprite(idx)
}

package ui

import "fyne.io/fyne/v2"

// EquipmentConfig struct (moved from main to avoid duplication)
type EquipmentConfig struct {
	PrimaryName          string
	PrimaryFocalLengthMm float64
	PrimaryFocalRatio    float64
	SecondaryName        string
	ImagingCamera        string
	GuideCamera          string
}

// Global reference to equipment config (set from main)
var equipmentConfigRef *EquipmentConfig

// SetEquipmentConfig sets the global equipment config reference
func SetEquipmentConfig(cfg *EquipmentConfig) {
	equipmentConfigRef = cfg
}

// getEquipmentConfig returns the current equipment config
func getEquipmentConfig() *EquipmentConfig {
	if equipmentConfigRef == nil {
		return &EquipmentConfig{}
	}
	return equipmentConfigRef
}

// WindowAccessor interface for testability
type WindowAccessor interface {
	GetWindow() fyne.Window
}

package ui

// NewDefaultEquipmentConfig restituisce una configurazione predefinita.
func NewDefaultEquipmentConfig() *EquipmentConfig {
	return &EquipmentConfig{
		PrimaryName:          "Newton 200/800",
		PrimaryFocalLengthMm: 800.0,
		PrimaryFocalRatio:    4.0,
		SecondaryName:        "",
		ImagingCamera:        "CCD Moravian",
		GuideCamera:          "ZWO ASI120MM",
	}
}

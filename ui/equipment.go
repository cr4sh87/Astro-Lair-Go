package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// EquipmentConfig definisce i parametri della strumentazione.
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

// WindowAccessor interface for testability
type WindowAccessor interface {
	GetWindow() fyne.Window
}

// ShowEquipmentDialog dialogo configurazione equipaggiamento
func ShowEquipmentDialog(win fyne.Window) {
	showEquipmentDialog(win)
}

func showEquipmentDialog(win fyne.Window) {
	eq := getEquipmentConfig()

	primaryName := widget.NewEntry()
	primaryName.SetText(eq.PrimaryName)

	primaryFocal := widget.NewEntry()
	primaryFocal.SetText(fmt.Sprintf("%.1f", eq.PrimaryFocalLengthMm))

	primaryRatio := widget.NewEntry()
	primaryRatio.SetText(fmt.Sprintf("%.1f", eq.PrimaryFocalRatio))

	secondaryName := widget.NewEntry()
	secondaryName.SetText(eq.SecondaryName)

	imagingCam := widget.NewEntry()
	imagingCam.SetText(eq.ImagingCamera)

	guideCam := widget.NewEntry()
	guideCam.SetText(eq.GuideCamera)

	form := container.NewVBox(
		widget.NewLabelWithStyle("Strumento Primario", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Nome", primaryName),
			widget.NewFormItem("Focale (mm)", primaryFocal),
			widget.NewFormItem("Rapporto focale (f/)", primaryRatio),
		),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Strumento Secondario", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Nome", secondaryName),
		),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Camere", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(
			widget.NewFormItem("Imaging", imagingCam),
			widget.NewFormItem("Guida", guideCam),
		),
	)

	dialog.NewCustomConfirm(
		"Configurazione Strumentazione",
		"Salva",
		"Annulla",
		form,
		func(ok bool) {
			if !ok {
				return
			}

			eq.PrimaryName = primaryName.Text
			eq.SecondaryName = secondaryName.Text
			eq.ImagingCamera = imagingCam.Text
			eq.GuideCamera = guideCam.Text

			if v, err := strconv.ParseFloat(strings.ReplaceAll(primaryFocal.Text, ",", "."), 64); err == nil {
				eq.PrimaryFocalLengthMm = v
			}
			if v, err := strconv.ParseFloat(strings.ReplaceAll(primaryRatio.Text, ",", "."), 64); err == nil {
				eq.PrimaryFocalRatio = v
			}
		},
		win,
	).Show()
}

// BuildToolsView — Tools UI
func BuildToolsView() fyne.CanvasObject {
	return buildToolsView()
}

func buildToolsView() fyne.CanvasObject {
	pixelSizeEntry := widget.NewEntry()
	pixelSizeEntry.SetText("4.3") // µm

	widthPxEntry := widget.NewEntry()
	widthPxEntry.SetText("4656")

	heightPxEntry := widget.NewEntry()
	heightPxEntry.SetText("3520")

	resultLabel := widget.NewLabel("Inserisci i dati e premi \"Calcola\".")
	resultLabel.Wrapping = fyne.TextWrapWord

	calcBtn := widget.NewButton("Calcola", func() {
		eq := getEquipmentConfig()
		fLen := eq.PrimaryFocalLengthMm
		if fLen <= 0 {
			resultLabel.SetText("Configura prima la focale dello strumento primario nelle impostazioni (⚙️).")
			return
		}

		pxSizeStr := strings.ReplaceAll(strings.TrimSpace(pixelSizeEntry.Text), ",", ".")
		pxSizeVal, err1 := strconv.ParseFloat(pxSizeStr, 64)

		wStr := strings.TrimSpace(widthPxEntry.Text)
		hStr := strings.TrimSpace(heightPxEntry.Text)
		wPx, err2 := strconv.Atoi(wStr)
		hPx, err3 := strconv.Atoi(hStr)

		if err1 != nil || err2 != nil || err3 != nil || pxSizeVal <= 0 || wPx <= 0 || hPx <= 0 {
			resultLabel.SetText("Valori non validi. Controlla pixel size e risoluzione in pixel.")
			return
		}

		// Scala in arcsec/px
		scale := 206.265 * pxSizeVal / fLen

		// Dimensioni sensore in mm
		sensorWidthMm := float64(wPx) * pxSizeVal / 1000.0
		sensorHeightMm := float64(hPx) * pxSizeVal / 1000.0

		// FOV in gradi
		fovWidthDeg := 57.2958 * sensorWidthMm / fLen
		fovHeightDeg := 57.2958 * sensorHeightMm / fLen

		resultLabel.SetText(fmt.Sprintf(
			"Scala di campionamento: %.2f\"/px\nFOV: %.2f° × %.2f° (≈ %.1f' × %.1f')",
			scale,
			fovWidthDeg, fovHeightDeg,
			fovWidthDeg*60.0, fovHeightDeg*60.0,
		))
	})

	form := widget.NewForm(
		widget.NewFormItem("Pixel size (µm)", pixelSizeEntry),
		widget.NewFormItem("Larghezza sensore (px)", widthPxEntry),
		widget.NewFormItem("Altezza sensore (px)", heightPxEntry),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("Strumenti", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Calcolo Scala & FOV\n(usando la focale dello strumento primario configurato)"),
		form,
		calcBtn,
		widget.NewSeparator(),
		resultLabel,
	)
}

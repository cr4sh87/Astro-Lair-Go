package main

import (
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Rappresenta la posizione di un satellite lungo il piano equatoriale visto di taglio.
// X è in "raggi planetari" (unità fittizie ma proporzionali).
// Behind = true significa "dietro il pianeta" (semplificazione geometrica).
type SatellitePos struct {
	Name   string
	X      float64
	Behind bool
}

// =======================
//  Modello semplificato – Giove
// =======================

var jRefEpoch = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

// Dati semplificati per i 4 galileiani (periodi e distanze in raggi gioviani circa)
type jSatParam struct {
	Name        string
	PeriodDays  float64 // periodo orbitale
	RadiusRj    float64 // distanza media dal pianeta (raggi gioviani)
	PhaseOffset float64 // fase iniziale [0..1), per sfalsare un minimo le posizioni
}

var jupiterSats = []jSatParam{
	{"Io", 1.769, 5.9, 0.10},
	{"Europa", 3.551, 9.4, 0.30},
	{"Ganimede", 7.155, 15.0, 0.55},
	{"Callisto", 16.689, 26.3, 0.80},
}

// Modello super-semplificato: orbite circolari sul piano, moto uniforme.
// Interpretiamo:
//
//	x = r * sin(angle)  (posizione sul piano del cielo, sinistra/destra)
//	z = r * cos(angle)  (profondità: z<0 = dietro il pianeta)
func computeJupiterSatellites(t time.Time) []SatellitePos {
	dtDays := t.Sub(jRefEpoch).Hours() / 24.0
	var out []SatellitePos

	for _, s := range jupiterSats {
		angle := 2.0 * math.Pi * (dtDays/s.PeriodDays + s.PhaseOffset)
		x := s.RadiusRj * math.Sin(angle)
		z := s.RadiusRj * math.Cos(angle)
		behind := z < 0 // metà orbita "dietro"

		out = append(out, SatellitePos{
			Name:   s.Name,
			X:      x,
			Behind: behind,
		})
	}
	return out
}

// =======================
//  Modello semplificato – Saturno
// =======================

var sRefEpoch = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

type sSatParam struct {
	Name        string
	PeriodDays  float64
	RadiusRs    float64 // raggi saturniani
	PhaseOffset float64
}

var saturnSats = []sSatParam{
	{"Mimas", 0.942, 3.1, 0.10},
	{"Encelado", 1.370, 3.9, 0.25},
	{"Teti", 1.888, 4.9, 0.40},
	{"Dione", 2.737, 6.3, 0.55},
	{"Rea", 4.518, 8.8, 0.70},
	{"Titano", 15.945, 20.3, 0.20},
	{"Giapeto", 79.321, 59.0, 0.85},
}

func computeSaturnSatellites(t time.Time) []SatellitePos {
	dtDays := t.Sub(sRefEpoch).Hours() / 24.0
	var out []SatellitePos

	for _, s := range saturnSats {
		angle := 2.0 * math.Pi * (dtDays/s.PeriodDays + s.PhaseOffset)
		x := s.RadiusRs * math.Sin(angle)
		z := s.RadiusRs * math.Cos(angle)
		behind := z < 0

		out = append(out, SatellitePos{
			Name:   s.Name,
			X:      x,
			Behind: behind,
		})
	}
	return out
}

// =======================
//  Rendering schema tipo TheSkyLive
// =======================

// Disegna un diagramma orizzontale: pianeta al centro, satelliti a sinistra/destra.
// widthPx è la larghezza disponibile (prendiamo quella del contenitore).
func buildSatDiagram(planetName string, sats []SatellitePos, widthPx float32) fyne.CanvasObject {
	if widthPx < 200 {
		widthPx = 400 // fallback la prima volta, prima che il layout dia una size reale
	}
	height := float32(220)
	centerX := widthPx / 2
	centerY := height / 2

	// Scala: quanti pixel per "raggio planetario"
	scale := float32(4.0)

	// Pianeta
	planetRadiusPx := float32(22)
	var planetColor color.NRGBA
	switch planetName {
	case "Giove":
		planetColor = color.NRGBA{R: 210, G: 180, B: 140, A: 255}
	case "Saturno":
		planetColor = color.NRGBA{R: 220, G: 210, B: 160, A: 255}
	default:
		planetColor = color.NRGBA{R: 180, G: 180, B: 180, A: 255}
	}
	planet := canvas.NewCircle(planetColor)
	planet.Resize(fyne.NewSize(planetRadiusPx*2, planetRadiusPx*2))
	planet.Move(fyne.NewPos(centerX-planetRadiusPx, centerY-planetRadiusPx))

	// "Linea" del piano equatoriale
	line := canvas.NewLine(color.NRGBA{R: 120, G: 120, B: 120, A: 255})
	line.StrokeWidth = 1
	line.Position1 = fyne.NewPos(10, centerY)
	line.Position2 = fyne.NewPos(widthPx-10, centerY)

	// Costruiamo le liste separate per "dietro" e "davanti"
	var behindObjs []fyne.CanvasObject
	var frontObjs []fyne.CanvasObject

	for i, s := range sats {
		x := centerX + float32(s.X)*scale

		// alterniamo leggermente sopra/sotto per non sovrapporre le label
		offsetY := float32(0)
		if i%2 == 0 {
			offsetY = -24
		} else {
			offsetY = +24
		}

		// colore: chiaro davanti, grigio scuro dietro
		var satColor color.NRGBA
		if s.Behind {
			satColor = color.NRGBA{R: 70, G: 70, B: 70, A: 255}
		} else {
			satColor = color.NRGBA{R: 180, G: 220, B: 255, A: 255}
		}

		// pallino satellite
		satCircle := canvas.NewCircle(satColor)
		satCircle.Resize(fyne.NewSize(10, 10))
		satCircle.Move(fyne.NewPos(x-5, centerY-5))

		// label col nome
		label := widget.NewLabel(s.Name)
		label.Alignment = fyne.TextAlignCenter
		labelContainer := container.NewWithoutLayout(label)
		label.Resize(fyne.NewSize(70, 16))
		label.Move(fyne.NewPos(x-35, centerY+offsetY))

		if s.Behind {
			behindObjs = append(behindObjs, satCircle, labelContainer)
		} else {
			frontObjs = append(frontObjs, satCircle, labelContainer)
		}
	}

	// Ordine di disegno:
	// 1) linea
	// 2) satelliti dietro
	// 3) pianeta
	// 4) satelliti davanti
	objs := []fyne.CanvasObject{line}
	objs = append(objs, behindObjs...)
	objs = append(objs, planet)
	objs = append(objs, frontObjs...)
	root := container.NewWithoutLayout(objs...)

	// wrapper che impone la min-size restituendola al layout system
	return container.New(layout.NewMaxLayout(), root)
}

// Costruisce la vista per un singolo pianeta, con controlli tempo + schema.
// -----------------------
// pagina per singolo pianeta
// -----------------------

type planetSatPage struct {
	root    fyne.CanvasObject
	refresh func()
}

func buildPlanetSatPage(planetName string, compute func(time.Time) []SatellitePos) planetSatPage {
	current := time.Now()

	timeLabel := widget.NewLabel("")
	timeLabel.Alignment = fyne.TextAlignCenter

	diagramHolder := container.New(layout.NewMaxLayout())

	updateUI := func() {
		timeLabel.SetText(current.Format("2006-01-02 15:04 MST"))

		// Usa la larghezza reale assegnata dal layout
		w := diagramHolder.Size().Width
		if w < 200 {
			if c := fyne.CurrentApp().Driver().CanvasForObject(diagramHolder); c != nil {
				w = c.Size().Width
			}
		}
		if w < 200 {
			w = 400
		}

		diagram := buildSatDiagram(planetName, compute(current), w)
		diagramHolder.Objects = []fyne.CanvasObject{diagram}
		diagramHolder.Refresh()
	}

	nowBtn := widget.NewButton("Adesso", func() {
		current = time.Now()
		updateUI()
	})
	minus1h := widget.NewButton("-1h", func() {
		current = current.Add(-1 * time.Hour)
		updateUI()
	})
	plus1h := widget.NewButton("+1h", func() {
		current = current.Add(1 * time.Hour)
		updateUI()
	})
	minus10m := widget.NewButton("-10m", func() {
		current = current.Add(-10 * time.Minute)
		updateUI()
	})
	plus10m := widget.NewButton("+10m", func() {
		current = current.Add(10 * time.Minute)
		updateUI()
	})

	buttonRow := container.NewHBox(
		layout.NewSpacer(),
		minus1h, minus10m, nowBtn, plus10m, plus1h,
		layout.NewSpacer(),
	)

	infoLabel := widget.NewLabel("Schema semplificato tipo vista equatoriale: non è una efemeride precisa,\nma rende bene la disposizione generale dei satelliti attorno al pianeta.")
	infoLabel.Wrapping = fyne.TextWrapWord

	root := container.NewVBox(
		widget.NewLabelWithStyle("Satelliti di "+planetName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		timeLabel,
		buttonRow,
		widget.NewSeparator(),
		diagramHolder,
		widget.NewSeparator(),
		infoLabel,
	)

	// NOTA: qui NON chiamiamo updateUI();
	// verrà chiamato dall'esterno quando la tab è davvero visibile.

	return planetSatPage{
		root:    root,
		refresh: updateUI,
	}
}

// -----------------------
// tab Satellites (Giove + Saturno)
// -----------------------

func buildSatellitesView() (fyne.CanvasObject, func()) {
	j := buildPlanetSatPage("Giove", computeJupiterSatellites)
	s := buildPlanetSatPage("Saturno", computeSaturnSatellites)

	tabs := container.NewAppTabs(
		container.NewTabItem("Giove", j.root),
		container.NewTabItem("Saturno", s.root),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// Quando cambi sottotab (Giove / Saturno) → rinfresca la pagina visibile
	tabs.OnChanged = func(ti *container.TabItem) {
		switch ti.Text {
		case "Giove":
			j.refresh()
		case "Saturno":
			s.refresh()
		}
	}

	view := container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		tabs,
	)

	// refreshAll simula "clic su Adesso" per entrambe quando entri nella tab Satellites
	refreshAll := func() {
		j.refresh()
		s.refresh()
	}

	return view, refreshAll
}

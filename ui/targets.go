package ui

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/cr4sh87/astro-lair-go/models"
	"github.com/cr4sh87/astro-lair-go/services"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// =======================
//  DSO Catalog (JSON)
// =======================

// Utilizzare i tipi definiti in models package
type DsoCatalog = models.DsoCatalog
type DsoObject = models.DsoObject
type TargetObject = models.TargetObject

var targetsByCatalog = models.TargetsByaCatalog

func floatPtr(f float64) *float64 { return models.FloatPtr(f) }

// =======================
//  Helpers RA/Dec
// =======================

func formatRAFromDeg(raDeg float64) string {
	// RA in ore: gradi / 15
	totalSeconds := raDeg / 15.0 * 3600.0
	if totalSeconds < 0 {
		totalSeconds = math.Mod(totalSeconds, 24*3600)
		if totalSeconds < 0 {
			totalSeconds += 24 * 3600
		}
	}
	h := int(totalSeconds / 3600)
	m := int(totalSeconds/60) % 60
	s := totalSeconds - float64(h*3600+m*60)
	return fmt.Sprintf("%02dh %02dm %04.1fs", h, m, s)
}

func formatDecFromDeg(decDeg float64) string {
	sign := '+'
	if decDeg < 0 {
		sign = '-'
	}
	absDeg := math.Abs(decDeg)
	totalArcsec := absDeg * 3600.0
	d := int(totalArcsec / 3600)
	m := int(totalArcsec/60) % 60
	s := totalArcsec - float64(d*3600+m*60)
	return fmt.Sprintf("%c%02d° %02d' %04.1f\"", sign, d, m, s)
}

// =======================
//  Catalog loading
// =======================

func buildTargetsCatalog() map[string][]TargetObject {
	data, err := os.ReadFile(services.DSOCatalogLocalPath)
	if err == nil {
		m, err := parseDsoTargetsJSON(data)
		if err == nil && len(m) > 0 {
			fmt.Printf("[Targets] Caricato catalogo dinamico (%d cataloghi)\n", len(m))
			return m
		}
		fmt.Printf("[Targets] Catalogo JSON non valido o vuoto, uso fallback. err=%v len=%d\n", err, len(m))
	} else {
		fmt.Printf("[Targets] Nessun catalogo locale (%s): %v. Uso fallback.\n", services.DSOCatalogLocalPath, err)
	}

	// fallback: catalogo hardcoded
	return targetsByCatalog
}

func parseDsoTargetsJSON(data []byte) (map[string][]TargetObject, error) {
	var cat DsoCatalog
	if err := json.Unmarshal(data, &cat); err != nil {
		return nil, err
	}

	out := make(map[string][]TargetObject)

	for _, o := range cat.Objects {
		if strings.TrimSpace(o.Catalog) == "" {
			continue
		}

		code := o.Code
		if code == "" && o.Number != nil {
			code = fmt.Sprintf("%d", *o.Number)
		}
		if code == "" {
			code = o.ID
		}

		var mag float64
		if o.Mag != nil {
			mag = *o.Mag
		} else {
			mag = 0
		}

		raStr := ""
		if o.RADeg != nil {
			raStr = formatRAFromDeg(*o.RADeg)
		}
		decStr := ""
		if o.DecDeg != nil {
			decStr = formatDecFromDeg(*o.DecDeg)
		}

		t := TargetObject{
			Catalog:       o.Catalog,
			Code:          code,
			Name:          o.Name,
			Type:          o.Type,
			Magnitude:     mag,
			SurfaceBright: o.SurfaceBrightness,
			RA:            raStr,
			Dec:           decStr,
			Constellation: o.Constellation,
		}

		out[o.Catalog] = append(out[o.Catalog], t)
	}

	for k, list := range out {
		sortTargetsByNumericCode(list)
		out[k] = list
	}

	return out, nil
}

// helper per ordinare i target in base alla parte numerica del codice
func sortTargetsByNumericCode(list []TargetObject) {
	sort.Slice(list, func(i, j int) bool {
		return extractNumericCode(list[i].Code) < extractNumericCode(list[j].Code)
	})
}

func extractNumericCode(code string) int {
	var digits []rune
	for _, ch := range code {
		if ch >= '0' && ch <= '9' {
			digits = append(digits, ch)
		}
	}
	if len(digits) == 0 {
		return math.MaxInt32
	}
	v, err := strconv.Atoi(string(digits))
	if err != nil {
		return math.MaxInt32
	}
	return v
}

type TargetsView struct {
	allByCatalog    map[string][]TargetObject
	catalogNames    []string
	currentCatalog  string
	allTargets      []TargetObject
	filtered        []TargetObject
	root            fyne.CanvasObject
	list            *widget.List
	searchEntry     *widget.Entry
	detailLabel     *widget.Label
	catalogSelector *widget.Select
}

// 👇 nome standardizzato
func BuildTargetsView(byCatalog map[string][]TargetObject) *TargetsView {
	tv := &TargetsView{
		allByCatalog: byCatalog,
	}

	for name := range byCatalog {
		tv.catalogNames = append(tv.catalogNames, name)
	}
	sort.Strings(tv.catalogNames)

	if len(tv.catalogNames) > 0 {
		tv.currentCatalog = tv.catalogNames[0]
	}

	for k, v := range tv.allByCatalog {
		copyList := append([]TargetObject(nil), v...)
		sortTargetsByNumericCode(copyList)
		tv.allByCatalog[k] = copyList
	}

	tv.allTargets = append(tv.allTargets, tv.allByCatalog[tv.currentCatalog]...)
	tv.filtered = append(tv.filtered, tv.allTargets...)

	tv.searchEntry = widget.NewEntry()
	tv.searchEntry.SetPlaceHolder("Cerca per nome, codice o costellazione…")
	tv.searchEntry.OnChanged = func(s string) {
		tv.applyFilter(s)
	}

	tv.catalogSelector = widget.NewSelect(tv.catalogNames, func(selected string) {
		if selected == "" {
			return
		}
		tv.switchCatalog(selected)
	})
	tv.catalogSelector.PlaceHolder = "Seleziona catalogo"
	if tv.currentCatalog != "" {
		tv.catalogSelector.Selected = tv.currentCatalog
	}

	tv.detailLabel = widget.NewLabel("Seleziona un target per vedere i dettagli.")
	tv.detailLabel.Wrapping = fyne.TextWrapWord

	tv.list = widget.NewList(
		func() int {
			return len(tv.filtered)
		},
		func() fyne.CanvasObject {
			title := widget.NewLabel("Nome target")
			sub := widget.NewLabel("Catalogo • Costellazione • Magnitudine")
			sub.TextStyle = fyne.TextStyle{Italic: true}
			return container.NewVBox(title, sub)
		},
		func(id widget.ListItemID, co fyne.CanvasObject) {
			if id < 0 || id >= len(tv.filtered) {
				return
			}
			t := tv.filtered[id]
			box := co.(*fyne.Container)
			title := box.Objects[0].(*widget.Label)
			sub := box.Objects[1].(*widget.Label)

			title.SetText(t.Name)
			sub.SetText(fmt.Sprintf("%s • %s • mag %.1f",
				fmt.Sprintf("%s %s", t.Catalog, t.Code),
				t.Constellation,
				t.Magnitude,
			))
		},
	)

	tv.list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(tv.filtered) {
			return
		}
		t := tv.filtered[id]
		tv.detailLabel.SetText(tv.formatDetails(t))
	}

	topControls := container.NewVBox(
		tv.catalogSelector,
		tv.searchEntry,
	)

	listCard := container.NewBorder(
		topControls,
		nil,
		nil,
		nil,
		tv.list,
	)

	detailCard := container.NewVBox(
		widget.NewLabelWithStyle("Dettagli target", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		tv.detailLabel,
	)

	detailScroll := container.NewVScroll(detailCard)
	detailScroll.SetMinSize(fyne.NewSize(220, 0))

	split := container.NewHSplit(listCard, detailScroll)
	split.Offset = 0.45

	tv.root = split
	return tv
}

func (tv *TargetsView) switchCatalog(catalog string) {
	tv.currentCatalog = catalog
	tv.allTargets = tv.allTargets[:0]
	tv.allTargets = append(tv.allTargets, tv.allByCatalog[catalog]...)
	tv.applyFilter(tv.searchEntry.Text)
}

func (tv *TargetsView) applyFilter(q string) {
	q = strings.TrimSpace(strings.ToLower(q))
	tv.filtered = tv.filtered[:0]

	if q == "" {
		tv.filtered = append(tv.filtered, tv.allTargets...)
		tv.list.Refresh()
		return
	}

	for _, t := range tv.allTargets {
		if strings.Contains(strings.ToLower(t.Name), q) ||
			strings.Contains(strings.ToLower(t.Constellation), q) ||
			strings.Contains(strings.ToLower(t.Code), q) ||
			strings.Contains(strings.ToLower(t.Catalog), q) {
			tv.filtered = append(tv.filtered, t)
		}
	}
	tv.list.Refresh()
}

func (tv *TargetsView) formatDetails(t TargetObject) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Nome: %s\n", t.Name)
	fmt.Fprintf(sb, "Catalogo: %s %s\n", t.Catalog, t.Code)
	fmt.Fprintf(sb, "Tipo: %s\n", t.Type)
	fmt.Fprintf(sb, "Costellazione: %s\n", t.Constellation)
	fmt.Fprintf(sb, "Magnitudine: %.1f\n", t.Magnitude)
	if t.SurfaceBright != nil {
		fmt.Fprintf(sb, "Luminosità superficiale: %.2f mag/arcsec²\n", *t.SurfaceBright)
	}
	fmt.Fprintf(sb, "\nCoordinate:\n  RA:  %s\n  Dec: %s\n", t.RA, t.Dec)
	return sb.String()
}

func (tv *TargetsView) Widget() fyne.CanvasObject {
	return tv.root
}

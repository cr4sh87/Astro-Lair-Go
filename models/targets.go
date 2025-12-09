package models

// =======================
//  Targets (astronomical objects)
// =======================

type TargetObject struct {
	Catalog       string
	Code          string
	Name          string
	Type          string
	Magnitude     float64
	SurfaceBright *float64
	RA            string
	Dec           string
	Constellation string
}

func FloatPtr(f float64) *float64 { return &f }

var TargetsByaCatalog = map[string][]TargetObject{
	"Messier": {
		{
			Catalog:       "Messier",
			Code:          "M31",
			Name:          "Galassia di Andromeda",
			Type:          "Galassia a spirale",
			Magnitude:     3.4,
			SurfaceBright: FloatPtr(22.0),
			RA:            "00h 42m 44s",
			Dec:           "+41° 16′ 09″",
			Constellation: "Andromeda",
		},
		{
			Catalog:       "Messier",
			Code:          "M42",
			Name:          "Nebulosa di Orione",
			Type:          "Nebulosa a emissione",
			Magnitude:     4.0,
			SurfaceBright: FloatPtr(21.0),
			RA:            "05h 35m 17s",
			Dec:           "−05° 23′ 28″",
			Constellation: "Orione",
		},
		{
			Catalog:       "Messier",
			Code:          "M45",
			Name:          "Pleiadi",
			Type:          "Ammasso aperto",
			Magnitude:     1.6,
			SurfaceBright: nil,
			RA:            "03h 47m 24s",
			Dec:           "+24° 07′ 00″",
			Constellation: "Toro",
		},
	},
	"NGC": {
		{
			Catalog:       "NGC",
			Code:          "NGC 7000",
			Name:          "Nebulosa Nord America",
			Type:          "Nebulosa a emissione",
			Magnitude:     4.0,
			SurfaceBright: nil,
			RA:            "20h 58m",
			Dec:           "+44° 20′",
			Constellation: "Cigno",
		},
		{
			Catalog:       "NGC",
			Code:          "NGC 253",
			Name:          "Galassia dello Scultore",
			Type:          "Galassia a spirale",
			Magnitude:     8.0,
			SurfaceBright: FloatPtr(22.5),
			RA:            "00h 47m 33s",
			Dec:           "−25° 17′ 18″",
			Constellation: "Scultore",
		},
	},
}

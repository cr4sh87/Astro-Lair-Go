package services

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// =======================
//  Remote Image Loading (SOHO/SUVI)
// =======================

// LoadRemoteImage scarica un'immagine da url, usa cache e aggiorna il canvas.Image.
func LoadRemoteImage(url, name string, img *canvas.Image, status *widget.Label) {
	if res, ok := GetImageFromCache(url); ok {
		img.Resource = res
		img.Refresh()
		if status != nil {
			status.SetText(name + " caricata dalla cache.")
		}
		return
	}

	if status != nil {
		status.SetText(fmt.Sprintf("Scarico immagine… (%s)", name))
	}

	go func() {
		resp, err := http.Get(url)
		if err != nil {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Errore download " + name + ":\n" + err.Error())
				}
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fyne.Do(func() {
				if status != nil {
					status.SetText(fmt.Sprintf("HTTP %d per %s", resp.StatusCode, name))
				}
			})
			return
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil || len(data) == 0 {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Dati non validi per " + name)
				}
			})
			return
		}

		res := fyne.NewStaticResource(name, data)

		fyne.Do(func() {
			SetImageInCache(url, res)
			img.Resource = res
			img.Refresh()
			if status != nil {
				status.SetText(name + " aggiornata.")
			}
		})
	}()
}

// =======================
//  Remote Animation Loading (SOHO/SUVI)
// =======================

// LoadRemoteAnimation:
//   - se source termina con .gif / .avi → scarica il file e ritorna []Resource
//   - altrimenti considera source una directory e scarica i frame (png/jpg) in ordine.
//
// Usa cache e chiama callback(frames) quando è pronto.
func LoadRemoteAnimation(source, name string, status *widget.Label, callback func([]fyne.Resource)) {
	if frames, ok := GetAnimationFromCache(source); ok {
		if status != nil {
			status.SetText(name + " caricata dalla cache.")
		}
		if callback != nil {
			fyne.Do(func() {
				callback(frames)
			})
		}
		return
	}

	if status != nil {
		status.SetText(fmt.Sprintf("Scarico animazione… (%s)", name))
	}

	lower := strings.ToLower(source)
	if strings.HasSuffix(lower, ".gif") || strings.HasSuffix(lower, ".avi") {
		go func() {
			resp, err := http.Get(source)
			if err != nil {
				fyne.Do(func() {
					if status != nil {
						status.SetText("Errore download " + name + ":\n" + err.Error())
					}
				})
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fyne.Do(func() {
					if status != nil {
						status.SetText(fmt.Sprintf("HTTP %d per %s", resp.StatusCode, name))
					}
				})
				return
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil || len(data) == 0 {
				fyne.Do(func() {
					if status != nil {
						status.SetText("Dati non validi per " + name)
					}
				})
				return
			}

			resName := filepath.Base(source)
			if resName == "" || !strings.Contains(resName, ".") {
				resName = name
			}

			var frames []fyne.Resource
			lowerResName := strings.ToLower(resName)
			if strings.HasSuffix(lowerResName, ".gif") {
				if g, err := gif.DecodeAll(bytes.NewReader(data)); err == nil && len(g.Image) > 0 {
					for i, pal := range g.Image {
						rgba := image.NewRGBA(pal.Bounds())
						draw.Draw(rgba, rgba.Bounds(), pal, pal.Bounds().Min, draw.Over)

						var buf bytes.Buffer
						if err := png.Encode(&buf, rgba); err != nil {
							continue
						}
						fname := fmt.Sprintf("%s_frame_%03d.png", strings.TrimSuffix(resName, ".gif"), i)
						frames = append(frames, fyne.NewStaticResource(fname, buf.Bytes()))
					}
				}
			}

			if len(frames) == 0 {
				frames = []fyne.Resource{fyne.NewStaticResource(resName, data)}
			}

			fyne.Do(func() {
				SetAnimationInCache(source, frames)
				if status != nil {
					status.SetText(name + " animazione pronta.")
				}
				if callback != nil {
					callback(frames)
				}
			})
		}()
		return
	}

	go func() {
		resp, err := http.Get(source)
		if err != nil {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Errore HTTP su directory animazione " + name + ":\n" + err.Error())
				}
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fyne.Do(func() {
				if status != nil {
					status.SetText(fmt.Sprintf("Risposta %d per %s: %s", resp.StatusCode, name, string(body)))
				}
			})
			return
		}

		htmlBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Errore lettura listing " + name + ":\n" + err.Error())
				}
			})
			return
		}
		html := string(htmlBytes)

		re := regexp.MustCompile(`href="([^"]+\.(?:png|jpg|jpeg))"`)
		matches := re.FindAllStringSubmatch(html, -1)
		if len(matches) == 0 {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Nessun frame trovato per " + name)
				}
			})
			return
		}

		seen := make(map[string]bool)
		var names []string
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			n := m[1]
			if strings.HasSuffix(n, "/") {
				continue
			}
			if seen[n] {
				continue
			}
			seen[n] = true
			names = append(names, n)
		}

		if len(names) == 0 {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Nessun frame valido per " + name)
				}
			})
			return
		}

		sort.Strings(names)

		const maxFrames = 60
		if len(names) > maxFrames {
			names = names[len(names)-maxFrames:]
		}

		var frames []fyne.Resource
		for _, n := range names {
			u := n
			if !strings.HasPrefix(u, "http") {
				u = source + n
			}

			r, err := http.Get(u)
			if err != nil {
				continue
			}
			data, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil || len(data) == 0 {
				continue
			}
			frames = append(frames, fyne.NewStaticResource(n, data))
		}

		if len(frames) == 0 {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Impossibile scaricare i frame per " + name)
				}
			})
			return
		}

		fyne.Do(func() {
			SetAnimationInCache(source, frames)
			if status != nil {
				status.SetText(fmt.Sprintf("%s: %d frame pronti.", name, len(frames)))
			}
			if callback != nil {
				callback(frames)
			}
		})
	}()
}

// =======================
//  Download Bulk Files (SOHO)
// =======================

// DownloadSOHOImages scarica le immagini SOHO (C2, C3, GIF)
func DownloadSOHOImages(status *widget.Label) {
	status.SetText("Scaricamento immagini SOHO...")

	const (
		sohoC2JPEG = "https://soho.nascom.nasa.gov/data/realtime/c2/512/latest.jpg"
		sohoC3JPEG = "https://soho.nascom.nasa.gov/data/realtime/c3/512/latest.jpg"
		sohoC2GIF  = "https://sohowww.nascom.nasa.gov/data/LATEST/current_c2.gif"
		sohoC3GIF  = "https://sohowww.nascom.nasa.gov/data/LATEST/current_c3.gif"
	)

	type item struct {
		url, name string
	}

	files := []item{
		{sohoC2JPEG, "lasco_c2_latest.jpg"},
		{sohoC3JPEG, "lasco_c3_latest.jpg"},
		{sohoC2GIF, "lasco_c2_48h.gif"},
		{sohoC3GIF, "lasco_c3_48h.gif"},
	}

	go func() {
		var errs []string

		home, err := os.UserHomeDir()
		if err != nil {
			fyne.Do(func() {
				status.SetText("Errore: impossibile determinare la home utente")
			})
			return
		}
		dir := filepath.Join(home, "AstroLair", "SOHO")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fyne.Do(func() {
				status.SetText("Errore: impossibile creare directory " + dir)
			})
			return
		}

		for _, f := range files {
			resp, err := http.Get(f.url)
			if err != nil {
				errs = append(errs, f.name)
				continue
			}
			data, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil || len(data) == 0 {
				errs = append(errs, f.name)
				continue
			}

			path := filepath.Join(dir, f.name)
			if err := os.WriteFile(path, data, 0o644); err != nil {
				errs = append(errs, f.name)
				continue
			}
		}

		fyne.Do(func() {
			if len(errs) > 0 {
				status.SetText("Scaricato con errori per: " + strings.Join(errs, ", "))
			} else {
				status.SetText("Immagini salvate in: " + dir)
			}
		})
	}()
}

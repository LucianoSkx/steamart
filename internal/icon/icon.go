// Pacote icon desenha o ícone temático do SteamArt (fundo Steam + grade).
package icon

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// PNG retorna o ícone codificado como PNG no tamanho informado.
func PNG(size int) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// fundo arredondado (navy da Steam)
	roundRect(img, 0, 0, size, size, size/5, color.RGBA{27, 40, 56, 255})
	// leve moldura interna clara
	roundRect(img, size/32, size/32, size-size/32, size-size/32, size/6, color.RGBA{42, 71, 94, 255})
	roundRect(img, size/22, size/22, size-size/22, size-size/22, size/6, color.RGBA{27, 40, 56, 255})

	m := size / 5         // margem
	gap := size * 6 / 100 // espaço entre tiles
	tile := (size - 2*m - gap) / 2
	r := tile / 6

	// grade 2x2 de tiles (um em destaque laranja = "arte")
	tiles := []struct {
		x, y int
		c    color.Color
	}{
		{m, m, color.RGBA{102, 192, 244, 255}},
		{m + tile + gap, m, color.RGBA{143, 211, 255, 255}},
		{m, m + tile + gap, color.RGBA{74, 159, 214, 255}},
		{m + tile + gap, m + tile + gap, color.RGBA{240, 165, 0, 255}},
	}
	for _, t := range tiles {
		roundRect(img, t.x, t.y, t.x+tile, t.y+tile, r, t.c)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// roundRect desenha um retângulo com cantos arredondados (raio r).
func roundRect(dst *image.RGBA, x0, y0, x1, y1, r int, c color.Color) {
	w, h := x1-x0, y1-y0
	mask := image.NewAlpha(image.Rect(0, 0, w, h))
	draw.Draw(mask, mask.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	for cy := 0; cy < r; cy++ {
		for cx := 0; cx < r; cx++ {
			if (cx-r)*(cx-r)+(cy-r)*(cy-r) > r*r {
				mask.SetAlpha(cx, cy, color.Alpha{0})
				mask.SetAlpha(w-1-cx, cy, color.Alpha{0})
				mask.SetAlpha(cx, h-1-cy, color.Alpha{0})
				mask.SetAlpha(w-1-cx, h-1-cy, color.Alpha{0})
			}
		}
	}
	draw.DrawMask(dst, image.Rect(x0, y0, x1, y1), &image.Uniform{c}, image.Point{}, mask, image.Point{}, draw.Over)
}

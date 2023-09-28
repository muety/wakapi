package utils

import (
	"fmt"
	"image/color"
)

func HexToRGBA(s string) (c color.RGBA) {
	// https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
	c.A = 0xff
	switch len(s) {
	case 7:
		fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		c.R *= 17
		c.G *= 17
		c.B *= 17
	}
	return
}

func RGBAToHex(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func FadeColors(color1, color2 color.RGBA, ratio float64) color.RGBA {
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}

	r := uint8(float64(color1.R)*(1-ratio) + float64(color2.R)*ratio)
	g := uint8(float64(color1.G)*(1-ratio) + float64(color2.G)*ratio)
	b := uint8(float64(color1.B)*(1-ratio) + float64(color2.B)*ratio)
	a := uint8(float64(color1.A)*(1-ratio) + float64(color2.A)*ratio)

	return color.RGBA{R: r, G: g, B: b, A: a}
}

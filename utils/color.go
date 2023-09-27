package utils

import (
	"fmt"
	"image/color"
)

func ParseHexColor(s string) (c color.RGBA) {
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

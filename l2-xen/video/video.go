package video

import (
	"image/color"
	"luna_l2/font"
	"cmp"
)

var CursorX int = 0
var CursorY int = 0
var MemoryVideo [0x4b000]byte
var Palette [256]color.NRGBA

func Clamp[T cmp.Ordered](x T, min T, max T) T {
    if x < min {
        return min
    }
    if x > max {
        return max
    }
    return x
}

func PushChar(x, y int, ch rune, fg byte, bg byte) {
    idx := int(ch)
    glyph := font.Font[0x00]

    if idx >= 0 && idx < len(font.Font) {
        glyph = font.Font[idx]
    }

    for row := 0; row < 8; row++ {
        line := glyph[row]
        
		for col := 0; col < 8; col++ {
			mask := byte(1 << col)
			var color byte
			if line&mask != 0 {
				color = fg
			} else {
				color = bg
			}
			px := (y+row)*320 + (x+col)
			MemoryVideo[Clamp(int(px), 0, 0x4b000 - 1)] = color
		}

    }
}

func PrintChar(ch rune, fg byte, bg byte) {
	if ch == 0x0a {
		CursorY++
		CursorX = 0
		return
	} else if ch == 0x0d {
		CursorX = 0
		return
	}	

	x := CursorX * 8
	y := CursorY * 8	

	PushChar(x, y, ch, fg, bg)

	CursorX++
	if CursorX >= 320/8 {
		CursorY++
		CursorX = 0
	}
	if CursorY >= 200/8 {
		for i := 0; i <= 0x4b000 - 1; i++ {
			MemoryVideo[i] = byte(00)
		} 
		CursorY = 0
	}
}

func InitializePalette() {
	for i := 0; i < 256; i++ {
		r := (i >> 5) & 0x07
        g := (i >> 2) & 0x07 
        b := i & 0x03        
 
        R := uint8(r * 255 / 7)
        G := uint8(g * 255 / 7)
        B := uint8(b * 255 / 3)

        Palette[i] = color.NRGBA{R, G, B, 255}
    }
}

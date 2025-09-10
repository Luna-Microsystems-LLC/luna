package video

var CursorX int = 0
var CursorY int = 0

var MemoryVideo [64000]byte
func PushChar(x, y int, ch rune, fg, bg byte, font [128][8]byte) {
    idx := int(ch)
    glyph := font[0x00]

    if idx >= 0 && idx < len(font) {
        glyph = font[idx]
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
			MemoryVideo[px] = color
		}

    }
}

func PrintChar(ch rune, fg, bg byte, font [128][8]byte) {
	x := CursorX * 8
	y := CursorY * 8

	if ch == 0x0a {
		CursorY++
		CursorX = 0	
	} else if ch == 0x0d {
		CursorX = 0
	} else {
		PushChar(x, y, ch, fg, bg, font)
	}

	CursorX++
	if CursorX >= 320/8 {
		CursorY++
		CursorX = 0
	}
	if CursorY >= 200/8 {
		CursorY = 0
	}
}

package draw

import "github.com/nmaupu/gopicto/config"

type PictoCell struct {
	config.ImageWord
	X, Y float64
	W, H float64
}

func NewPictoCell(marginTB, marginLR, x, y, w, h float64, iw config.ImageWord) PictoCell {
	return PictoCell{
		ImageWord: iw,
		X:         x + marginLR,
		Y:         y + marginTB,
		W:         w - marginLR*2,
		H:         h - marginTB*2,
	}
}

package draw

import "github.com/nmaupu/gopicto/config"

type PictoCell struct {
	config.ImageWord
	X, Y float64
	W, H float64
}

func NewPictoCell(margins config.Margins, x, y, w, h float64, iw config.ImageWord) PictoCell {
	return PictoCell{
		ImageWord: iw,
		X:         x + margins.Left(),
		Y:         y + margins.Top(),
		W:         w - margins.LeftRight(),
		H:         h - margins.TopBottom(),
	}
}

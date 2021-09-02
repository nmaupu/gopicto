package main

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/nmaupu/gopicto/config"
	"github.com/nmaupu/gopicto/draw"
	"github.com/signintech/gopdf"
	"github.com/spf13/viper"
	"image"
	"os"
)

const (
	MarginsTopBottomPt = 5.67 / 2
	MarginsLeftRightPt = 5.67 / 2
	PaddingLeftRightPt = 3.0
	PaddingTopBottomPt = 3.0
)

func main() {
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToSliceHookFunc(","),
		config.MapstructureStringToRatio(),
		config.MapstructureStringToColor(),
	)

	cfg := config.PDF{}
	viper.SetConfigFile("./config.sample.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = viper.Unmarshal(&cfg, viper.DecodeHook(decodeHook))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if cfg.Lines == 0 || cfg.Cols == 0 {
		fmt.Println("Invalid configuration: cols and lines have to be > 0")
		os.Exit(1)
	}

	if cfg.Text.Ratio == 0.0 {
		cfg.Text.Ratio = 1 / 5
	}

	pdf := gopdf.GoPdf{}
	// Unit is pt as gopdf's unit support seems to be broken
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
	})

	err = pdf.AddTTFFont("rockwell", "ttf/rockwell.ttf")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = pdf.SetFont("rockwell", "", 14)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cellW := gopdf.PageSizeA4.W / float64(cfg.Cols)
	cellH := gopdf.PageSizeA4.H / float64(cfg.Lines)
	page := 0
	emptyPage := true
mainLoop:
	for {
		for l := 0; l < cfg.Lines; l++ {
			for c := 0; c < cfg.Cols; c++ {
				idx := page*cfg.Cols*cfg.Lines + cfg.Cols*l + c
				if idx >= len(cfg.Images) { // no more images
					break mainLoop
				}

				if emptyPage {
					pdf.AddPage()
					emptyPage = false
				}

				pc := draw.NewPictoCell(
					MarginsTopBottomPt,
					MarginsLeftRightPt,
					float64(c)*cellW,
					float64(l)*cellH,
					cellW,
					cellH,
					cfg.Images[idx],
				)

				printPdfCell(&pdf, cfg.Text, pc)
			}
		}

		emptyPage = true
		page++
	}

	pdf.WritePdf("/tmp/test.pdf")

}

func printPdfCell(pdf *gopdf.GoPdf, textConfig config.Text, c draw.PictoCell) {
	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeft(c.X, c.Y, c.W, c.H)

	cellTextHeightPt := c.H * textConfig.Ratio
	imgW, imgH, _ := getImageDimension(c.Image)
	var w, h float64
	if c.W >= c.H {
		// Image should fill the height of the cell except if larger than high
		h = c.H - cellTextHeightPt - 2*PaddingTopBottomPt
		w = imgW * h / imgH

		if w > c.W-2*MarginsLeftRightPt { // image width is wider than the outer cell
			w = c.W - 2*PaddingLeftRightPt
			h = imgH * w / imgW
		}
	} else {
		// Image should fill the width of the cell
		w = c.W - 2*PaddingLeftRightPt
		h = imgH * w / imgW

		if h > c.H-cellTextHeightPt-2*PaddingTopBottomPt { // image height is higher than the outer cell
			h = c.H - cellTextHeightPt - 2*PaddingTopBottomPt
			w = imgW * h / imgH
		}
	}

	var x, y float64
	x = c.X + (c.W-w)/2
	y = c.Y + PaddingTopBottomPt

	pdf.Image(c.Image, x, y, &gopdf.Rect{
		W: w,
		H: h,
	})

	textWidth, _ := pdf.MeasureTextWidth(c.Text)
	pdf.SetX(c.X + c.W/2 - textWidth/2)
	pdf.SetY(c.Y + c.H - cellTextHeightPt/2)
	pdf.SetTextColor(textConfig.FirstLetterColor.AsUints())
	pdf.Text(string(c.Text[0]))
	pdf.SetTextColor(textConfig.Color.AsUints())
	pdf.Text(c.Text[1:])
}

func getImageDimension(imagePath string) (float64, float64, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return float64(image.Width), float64(image.Height), nil
}

func printCutLines(pdf *gopdf.GoPdf, cfg config.PDF, incX, incY float64) {
	var x float64

	for i := 1; i < cfg.Cols; i++ {
		x = float64(i) * incX
		pdf.SetLineWidth(1)
		pdf.SetLineType("dotted")
		pdf.Line(x, 0, x, gopdf.PageSizeA4.H)
		x += incX
	}

	var y float64
	for y < gopdf.PageSizeA4.H {
		y += incY
		pdf.SetLineWidth(1)
		pdf.SetLineType("dotted")
		pdf.Line(0, y, gopdf.PageSizeA4.W, y)
		y += incY
	}
}

package main

import (
	"fmt"
	"github.com/nmaupu/gopicto/config"
	"github.com/signintech/gopdf"
	"github.com/spf13/viper"
	"image"
	"os"
)

const (
	MarginsTopBottomPt = 5.67
	MarginsLeftRightPt = 5.67
	CellTextHighPt     = 35
)

func main() {
	cfg := config.PDF{}

	viper.SetConfigFile("./config.sample.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	pdf := gopdf.GoPdf{}
	// Unit is pt as gopdf's unit support seems to be broken
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
	})
	pdf.AddPage()

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

	cellWpt := gopdf.PageSizeA4.W / float64(cfg.Cols)
	imgWpt := cellWpt - 4*MarginsLeftRightPt

	mx := 0.0 // pt
	my := 0.0 // pt
	imgMaxWpx := 0.0
	imgMaxHpx := 0.0
	for k, iw := range cfg.Images {
		w, h, err := getImageDimension(iw.Image)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// imgWpx is set once and for all
		// All other images are calculated from this max width
		if imgMaxWpx == 0.0 {
			baseRatio := imgWpt / float64(w)
			imgMaxWpx = float64(w) * baseRatio
			imgMaxHpx = float64(h) * baseRatio
		}

		var imageRatio float64
		if w > h {
			imageRatio = float64(w) / imgMaxWpx
		} else {
			imageRatio = float64(h) / imgMaxHpx
		}
		imgWpx := float64(w) / imageRatio
		imgHpx := float64(h) / imageRatio

		cellHpt := imgMaxHpx + 4*MarginsTopBottomPt + CellTextHighPt
		printPdfCell(&pdf, iw, mx, my, imgWpx, imgHpx, gopdf.Rect{W: cellWpt - 2*MarginsLeftRightPt, H: cellHpt})

		mx += cellWpt

		if mx >= gopdf.PageSizeA4.W { // need to go to the next line
			mx = 0
			my += cellHpt + MarginsTopBottomPt*2
		}

		if my+cellHpt >= gopdf.PageSizeA4.H { // need a new page
			printCutLines(&pdf, cfg, cellWpt, cellHpt+MarginsTopBottomPt*2)
			mx = 0
			my = 0

			if k < len(cfg.Images)-1 {
				pdf.AddPage()
			}
		}

	}

	pdf.WritePdf("/tmp/test.pdf")

}

func printPdfCell(pdf *gopdf.GoPdf, iw config.ImageWord, x, y float64, imgW float64, imgH float64, cell gopdf.Rect) {
	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeft(x+MarginsLeftRightPt, y+MarginsTopBottomPt, cell.W, cell.H)

	pdf.Image(iw.Image, x+MarginsLeftRightPt*2, y+MarginsTopBottomPt*2, &gopdf.Rect{W: imgW, H: imgH})

	pdf.SetX(x + MarginsLeftRightPt)
	pdf.SetY(y + MarginsTopBottomPt + cell.H - CellTextHighPt)
	pdf.CellWithOption(
		&gopdf.Rect{
			W: cell.W,
			H: CellTextHighPt,
		},
		iw.Text,
		gopdf.CellOption{
			Align: gopdf.Center | gopdf.Middle,
		})
}

func getImageDimension(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return image.Width, image.Height, nil
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

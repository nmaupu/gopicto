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
	CellTextHighPt     = 56.7
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
	for k, iw := range cfg.Images {
		w, h, err := getImageDimension(iw.Image)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		imgRatio := imgWpt / float64(w)
		imgWpx := float64(w) * imgRatio
		imgHpx := float64(h) * imgRatio

		printPdfCell(&pdf, iw, mx, my, imgWpx, imgHpx)

		cellHpt := imgHpx + 2*MarginsTopBottomPt + CellTextHighPt + 2*MarginsTopBottomPt
		mx += cellWpt

		if mx >= gopdf.PageSizeA4.W { // need to go to the next line
			mx = 0
			my += cellHpt + MarginsTopBottomPt*2
		}

		if my+cellHpt >= gopdf.PageSizeA4.H && k < len(cfg.Images)-1 { // need a new page
			pdf.AddPage()
			mx = 0
			my = 0
		}
	}

	pdf.WritePdf("/tmp/test.pdf")

}

func printPdfCell(pdf *gopdf.GoPdf, iw config.ImageWord, x, y float64, imgW float64, imgH float64) {
	cellRect := gopdf.Rect{
		W: imgW + 2*MarginsLeftRightPt,
		H: imgH + 2*MarginsTopBottomPt + CellTextHighPt + 2*MarginsTopBottomPt,
	}

	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeft(x+MarginsLeftRightPt, y+MarginsTopBottomPt, cellRect.W, cellRect.H)

	pdf.Image(iw.Image, x+MarginsLeftRightPt*2, y+MarginsTopBottomPt*2, &gopdf.Rect{W: imgW, H: imgH})

	pdf.SetX(x + MarginsLeftRightPt)
	pdf.SetY(y + imgH + 2*MarginsTopBottomPt)
	pdf.CellWithOption(
		&gopdf.Rect{
			W: cellRect.W,
			H: CellTextHighPt + 2*MarginsTopBottomPt,
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

package main

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/nmaupu/gopicto/config"
	"github.com/nmaupu/gopicto/draw"
	"github.com/signintech/gopdf"
	"github.com/spf13/viper"
	"image"
	"io/ioutil"
	"math"
	"os"
)

const (
	defaultFont    = "rockwell"
	fontFamilyName = "myfont"
)

var (
	PageSize *gopdf.Rect
)

func main() {
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToSliceHookFunc(","),
		config.MapstructureStringToFloat64Expr(),
		config.MapstructureStringToColor(),
		config.MapstructureStringToOrientation(),
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

	if cfg.Page.Lines == 0 || cfg.Page.Cols == 0 {
		fmt.Println("Invalid configuration: cols and lines have to be > 0")
		os.Exit(1)
	}

	cfg.Page.Margins.InitWithDefaults(config.DefaultMargins)
	cfg.Page.Paddings.InitWithDefaults(config.DefaultPaddings)

	if cfg.Text.Ratio == 0.0 {
		cfg.Text.Ratio = 1 / 5
	}

	if cfg.Text.Font == "" {
		fmt.Printf("text.font is not provided, using default %s\n", defaultFont)
		reader, err := config.LoadFont(defaultFont)
		if err != nil {
			fmt.Printf("Unable to load font %s, err=%v\n", defaultFont, err)
			os.Exit(1)
		}

		file, err := os.CreateTemp("", defaultFont)
		if err != nil {
			fmt.Printf("unable to create temporary file for font %s, err=%v\n", defaultFont, err)
			os.Exit(1)
		}
		defer file.Close()
		cfg.Text.Font = file.Name()
		defer os.Remove(cfg.Text.Font)

		data, err := ioutil.ReadAll(reader)
		if err != nil {
			fmt.Printf("unable to read data for font %s, err=%v\n", defaultFont, err)
			os.Exit(1)
		}

		ioutil.WriteFile(file.Name(), data, 0644)
	}

	pdf := gopdf.GoPdf{}

	PageSize = gopdf.PageSizeA4
	if cfg.Page.Orientation == config.Landscape {
		PageSize = &gopdf.Rect{W: gopdf.PageSizeA4.H, H: gopdf.PageSizeA4.W}
	}
	// Unit is pt as gopdf's unit support seems to be broken
	pdf.Start(gopdf.Config{
		PageSize: *PageSize,
	})

	err = pdf.AddTTFFont(fontFamilyName, cfg.Text.Font)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cellW := PageSize.W / float64(cfg.Page.Cols)
	cellH := PageSize.H / float64(cfg.Page.Lines)
	page := 0
	emptyPage := true

	// Getting font size to set
	longestText := ""
	for _, iw := range cfg.ImageWords {
		if len(iw.Text) > len(longestText) {
			longestText = iw.Text
		}
	}

	fontSize := setMaxFontSize(&pdf, longestText, cellW, cellH*cfg.Text.Ratio)

mainLoop:
	for {
		for l := 0; l < cfg.Page.Lines; l++ {
			for c := 0; c < cfg.Page.Cols; c++ {
				idx := page*cfg.Page.Cols*cfg.Page.Lines + cfg.Page.Cols*l + c
				if idx >= len(cfg.ImageWords) { // no more images
					break mainLoop
				}

				if emptyPage {
					pdf.AddPage()
					emptyPage = false
				}

				pc := draw.NewPictoCell(
					cfg.Page.Margins,
					float64(c)*cellW,
					float64(l)*cellH,
					cellW,
					cellH,
					cfg.ImageWords[idx],
				)

				printPdfCell(&pdf, cfg, pc, fontSize)
			}
		}

		emptyPage = true
		page++
	}

	pdf.WritePdf("/tmp/test.pdf")
}

func printPdfCell(pdf *gopdf.GoPdf, cfg config.PDF, c draw.PictoCell, fontSize int) {
	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeft(c.X, c.Y, c.W, c.H)

	cellTextHeightPt := c.H * cfg.Text.Ratio
	imgW, imgH, _ := getImageDimension(c.Image)
	var w, h float64
	if c.W >= c.H {
		// Image should fill the height of the cell except if larger than high
		h = c.H - cellTextHeightPt - cfg.Page.Paddings.TopBottom()
		w = imgW * h / imgH

		if w > c.W-cfg.Page.Margins.LeftRight() { // image width is wider than the outer cell
			w = c.W - cfg.Page.Paddings.LeftRight()
			h = imgH * w / imgW
		}
	} else {
		// Image should fill the width of the cell except if height is more than available space
		w = c.W - cfg.Page.Paddings.LeftRight()
		h = imgH * w / imgW

		if h > c.H-cellTextHeightPt-cfg.Page.Paddings.TopBottom() { // image height is higher than the outer cell
			h = c.H - cellTextHeightPt - cfg.Page.Paddings.TopBottom()
			w = imgW * h / imgH
		}
	}

	var x, y float64
	x = c.X + (c.W-w)/2
	y = c.Y + cfg.Page.Paddings.Top()

	pdf.Image(c.Image, x, y, &gopdf.Rect{
		W: w,
		H: h,
	})

	// Handling text
	if len(c.Text) == 0 {
		return
	}

	//pdf.RectFromUpperLeft(c.X, c.Y+c.H-cellTextHeightPt, c.W, cellTextHeightPt)

	textWidth, _ := pdf.MeasureTextWidth(c.Text)
	textHeight := gopdf.ContentObjCalTextHeight(fontSize)
	pdf.SetX(c.X + c.W/2 - textWidth/2)
	pdf.SetY(c.Y + c.H - cellTextHeightPt/2 + textHeight/2)
	pdf.SetTextColor(cfg.Text.FirstLetterColor.AsUints())
	pdf.Text(string(c.Text[0]))
	pdf.SetTextColor(cfg.Text.Color.AsUints())
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

	for i := 1; i < cfg.Page.Cols; i++ {
		x = float64(i) * incX
		pdf.SetLineWidth(1)
		pdf.SetLineType("dotted")
		pdf.Line(x, 0, x, PageSize.H)
		x += incX
	}

	var y float64
	for y < PageSize.H {
		y += incY
		pdf.SetLineWidth(1)
		pdf.SetLineType("dotted")
		pdf.Line(0, y, PageSize.W, y)
		y += incY
	}
}

func setMaxFontSize(pdf *gopdf.GoPdf, text string, maxWidth, maxHeight float64) int {
	fontSize := 110
	inc := -1
	for {
		fontSize += inc

		if fontSize < int(math.Abs(float64(inc))) {
			break // no size found
		}

		err := pdf.SetFont(fontFamilyName, "", fontSize)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		textWidth, _ := pdf.MeasureTextWidth(text)
		textHeight := gopdf.ContentObjCalTextHeight(fontSize)
		//textHeightWithMargin := textHeight + textHeight*70/100
		if textWidth < maxWidth && textHeight < maxHeight {
			// Height does not take accents and letters like p, q, etc.
			// Taking 50% size because why not 🤷‍
			newFontSize := float64(fontSize) * 0.5
			fontSize = int(newFontSize)
			err := pdf.SetFont(fontFamilyName, "", fontSize)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Printf("Font size %d is ok\n", fontSize)
			break
		}
	}

	return fontSize
}

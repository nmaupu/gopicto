package cli

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/nmaupu/gopdf"
	"github.com/nmaupu/gopicto/config"
	"github.com/nmaupu/gopicto/draw"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"image"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"
)

type pageMode string
type cellPrinter func(pdf *gopdf.GoPdf, cfg config.PDF, c draw.PictoCell, fontSize float64)

const (
	fontFamilyNameText        = "fontText"
	fontFamilyNameDefinitions = "fontDefinitions"
	pageModePictos            = "pictos"
	pageModeDefinitions       = "definitions"
	defaultLineSpacingRatio   = .3
	defaultImageWordTextRatio = 1 / 5
	defaultTwoSidedOffsetMMx  = -3
	defaultTwoSidedOffsetMMy  = 0
)

var (
	pageSize *gopdf.Rect

	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate PDF containing a set of picto/word",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			generateCmdFunc()
		},
	}

	ViperConfigKey = "gopictoConfig"

	defaultConfigFile = "./config.yaml"
	defaultOutputFile = "/tmp/gopicto.pdf"

	cfgFile  string
	outFile  string
	cutLines bool
)

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&cfgFile, ConfigFlag, "c", defaultConfigFile, "Config file to use")
	generateCmd.Flags().StringVarP(&outFile, OutputFlag, "o", defaultOutputFile, "Specify the name of the file generated")
	generateCmd.Flags().BoolVarP(&cutLines, CutLinesFlag, "k", false, "Draw cut lines around cells")
}

func initConfig() {
	if cfgFile == "" {
		cfgFile = defaultConfigFile
	}

	if outFile == "" {
		outFile = defaultOutputFile
	}
	viper.Set(OutputFlag, outFile)

	viper.AutomaticEnv()

	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToSliceHookFunc(","),
		config.MapstructureStringToFloat64Expr(),
		config.MapstructureStringToColor(),
		config.MapstructureStringToOrientation(),
		config.MapstructureStringToTextAlign(),
	)

	cfg := config.PDF{}
	readCfgLogger := log.With().Str("config", cfgFile).Logger()
	viper.SetConfigFile(cfgFile)
	err := viper.ReadInConfig()
	if err != nil {
		readCfgLogger.Fatal().
			Err(err).
			Msg("Unable to read configuration file")
	}
	err = viper.Unmarshal(&cfg, viper.DecodeHook(decodeHook))
	if err != nil {
		readCfgLogger.Fatal().
			Err(err).
			Msg("Unable to unmarshal configuration file")
	}

	// Validate and init inputs
	if cfg.Page.Lines == 0 || cfg.Page.Cols == 0 {
		readCfgLogger.Fatal().Msg("Invalid configuration: cols and lines have to be > 0")
	}

	cfg.Page.PageMargins.InitWithDefaults(config.DefaultPageMargins)
	cfg.Page.Margins.InitWithDefaults(config.DefaultMargins)
	cfg.Page.Paddings.InitWithDefaults(config.DefaultPaddings)

	if cfg.Page.TwoSidedOffsetMM.X == 0 {
		cfg.Page.TwoSidedOffsetMM.X = defaultTwoSidedOffsetMMx
	}

	if cfg.Page.TwoSidedOffsetMM.Y == 0 {
		cfg.Page.TwoSidedOffsetMM.Y = defaultTwoSidedOffsetMMy
	}

	if cfg.Text.Ratio == 0.0 {
		cfg.Text.Ratio = defaultImageWordTextRatio
	}

	for k, iw := range cfg.ImageWords {
		if iw.Def.LineSpacingRatio == 0 {
			// Can't use iw here because it's a copy of the original object
			cfg.ImageWords[k].Def.LineSpacingRatio = defaultLineSpacingRatio
			if iw.Def.Align == "" {
				cfg.ImageWords[k].Def.Align = config.TextAlignCenter
			}
		}
	}

	if cfg.Text.Font == "" {
		fontLogger := log.With().Str("font", config.DefaultFont).Logger()
		log.Info().Msgf("Config text.font is not provided, using default font %s", config.DefaultFont)
		reader, err := config.LoadFont(config.DefaultFont)
		if err != nil {
			fontLogger.Fatal().Err(err).Msg("unable to load font")
		}

		fontLogger.Debug().Msg("Creating temporary file to load font")
		file, err := os.CreateTemp("", config.DefaultFont)
		if err != nil {
			fontLogger.Fatal().Err(err).Msg("unable to create temporary file for font")
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Error().Err(err).Msg("unable to close file")
			}
		}(file)
		cfg.Text.Font = file.Name()

		rootCmd.PostRun = func(cmd *cobra.Command, args []string) {
			log.Debug().
				Str("file", cfg.Text.Font).
				Msg("Cleaning temp file")
			err := os.Remove(cfg.Text.Font) // Cleaning
			if err != nil {
				log.Error().Err(err).Msg("unable to remove temp file")
			}
		}

		data, err := ioutil.ReadAll(reader)
		if err != nil {
			fontLogger.Fatal().Err(err).Msg("unable to read font data")
		}

		err = ioutil.WriteFile(file.Name(), data, 0644)
		if err != nil {
			fontLogger.Error().Err(err).Msg("unable to write file")
		}
	}

	viper.Set(ViperConfigKey, cfg)
}

func generateCmdFunc() {
	pdf := gopdf.GoPdf{}
	cfg := viper.Get(ViperConfigKey).(config.PDF)

	pageSize = gopdf.PageSizeA4
	if cfg.Page.Orientation == config.Landscape {
		pageSize = &gopdf.Rect{W: gopdf.PageSizeA4.H, H: gopdf.PageSizeA4.W}
	}
	// Unit is pt as gopdf's unit support seems to be broken
	pdf.Start(gopdf.Config{
		PageSize: *pageSize,
		//Unit:     gopdf.UnitMM,
	})

	// Even page contains definitions ?
	haveDefinitions := false
	for _, iw := range cfg.ImageWords {
		if iw.Def.Text != "" {
			haveDefinitions = true
			break
		}
	}

	err := pdf.AddTTFFont(fontFamilyNameText, cfg.Text.Font)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("font", cfg.Text.Font).
			Msg("unable to use font")
	}

	if haveDefinitions {
		defFont := cfg.Text.Definitions.Font
		if defFont == "" {
			defFont = cfg.Text.Font
		}
		err := pdf.AddTTFFont(fontFamilyNameDefinitions, defFont)
		if err != nil {
			log.Fatal().
				Err(err).
				Str("font", cfg.Text.Definitions.Font).
				Msg("unable to use font")
		}
	}

	cellW := (pageSize.W - cfg.Page.PageMargins.Left() - cfg.Page.PageMargins.Right()) / float64(cfg.Page.Cols)
	cellH := (pageSize.H - cfg.Page.PageMargins.Top() - cfg.Page.PageMargins.Bottom()) / float64(cfg.Page.Lines)

	// Getting font size to set
	longestText := ""
	for _, iw := range cfg.ImageWords {
		if len(longestText) < len(iw.Text) {
			longestText = iw.Text
		}
	}

	pictoTextFontSize := cfg.Text.FontSize
	if pictoTextFontSize == 0 {
		pictoTextFontSize = setMaxFontSize(&pdf, longestText, cellW, cellH*cfg.Text.Ratio)
	}

	nbPictoPages := cfg.GetNbPictoPages()
	for page := 0; page < nbPictoPages; page++ {
		printPdfPage(&pdf, cfg, page, cellW, cellH, pageModePictos, pictoTextFontSize)
		if haveDefinitions {
			printPdfPage(&pdf, cfg, page, cellW, cellH, pageModeDefinitions, pictoTextFontSize)
		}

	}

	output := viper.GetString(OutputFlag)
	err = pdf.WritePdf(output)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("file", output).
			Msg("unable to write pdf")
	}
	log.Info().
		Str("file", output).
		Msg("PDF written successfully")
}

// printPdfPage prints a page
func printPdfPage(pdf *gopdf.GoPdf, cfg config.PDF, page int, cellW float64, cellH float64, mode pageMode, fontSize float64) {
	pdf.AddPage()

	// Printer are misaligned when printing two-sided, adding an offset on odd pages to compensate
	offsetX := float64(0)
	offsetY := float64(0)
	if mode == pageModePictos {
		offsetX = gopdf.UnitsToPoints(gopdf.UnitMM, cfg.Page.TwoSidedOffsetMM.X)
		offsetY = gopdf.UnitsToPoints(gopdf.UnitMM, cfg.Page.TwoSidedOffsetMM.Y)
	}

	if cutLines {
		printCutLines(pdf, cfg, offsetX, offsetY)
	}

	for l := 0; l < cfg.Page.Lines; l++ {
		for c := 0; c < cfg.Page.Cols; c++ {
			idx := page*cfg.Page.Cols*cfg.Page.Lines + cfg.Page.Cols*l + c
			if idx >= len(cfg.ImageWords) { // no more images
				return
			}

			x := cfg.Page.PageMargins.Left() + float64(c)*cellW + offsetX
			y := cfg.Page.PageMargins.Top() + float64(l)*cellH + offsetY
			if mode == pageModeDefinitions {
				x = cfg.Page.PageMargins.Left() + (float64(cfg.Page.Cols)-float64(c)-1)*cellW
			}

			pc := draw.NewPictoCell(
				cfg.Page.Margins,
				x,
				y,
				cellW,
				cellH,
				cfg.ImageWords[idx],
			)

			printPdfCell(pdf, cfg, pc, fontSize, mode)
		}
	}
}

func printPdfCell(pdf *gopdf.GoPdf, cfg config.PDF, c draw.PictoCell, fontSize float64, mode pageMode) {
	pdf.SetLineWidth(1)
	pdf.SetLineType("")
	if mode == pageModePictos || (mode == pageModeDefinitions && (cfg.Text.Definitions.Borders || c.Def.Borders)) {
		pdf.RectFromUpperLeft(c.X, c.Y, c.W, c.H)
	}

	var cellPrinterFunc cellPrinter
	switch mode {
	case pageModePictos:
		cellPrinterFunc = printCellPicto
	case pageModeDefinitions:
		cellPrinterFunc = printCellDefinition
	}
	cellPrinterFunc(pdf, cfg, c, fontSize)

}

// printCellPicto prints a cell with a picto and a text on top or bottom
func printCellPicto(pdf *gopdf.GoPdf, cfg config.PDF, c draw.PictoCell, fontSize float64) {
	pdf.SetFont(fontFamilyNameText, "", fontSize)

	cellTextHeightPt := c.H * cfg.Text.Ratio
	imgW, imgH, _ := getImageDimension(c.Image)
	var w, h float64
	if c.W >= c.H {
		// Image should fill the height of the cell except if larger than height
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

	// Depending on the font, this does not take into account "high/low" letters (e.g. f,g,y,t,l etc.)
	textHeight := gopdf.ContentObjCalTextHeightPrecise(fontSize)
	textOffsetY := c.H - cellTextHeightPt/2 + textHeight/2 - cfg.Page.Paddings.Bottom()
	imageOffsetY := cfg.Page.Paddings.Top()
	if cfg.Text.Top { // Drawing text on the top of the cell
		textOffsetY = textHeight + cfg.Page.Paddings.Top()
		imageOffsetY = cellTextHeightPt + cfg.Page.Paddings.Top()
	}

	var x, y float64
	x = c.X + (c.W-w)/2
	y = c.Y + imageOffsetY

	err := pdf.Image(c.Image, x, y, &gopdf.Rect{
		W: w,
		H: h,
	})
	if err != nil {
		log.Error().Err(err).Msg("problem creating pdf image")
	}

	//pdf.RectFromUpperLeft(c.X, c.Y+c.H-cellTextHeightPt, c.W, cellTextHeightPt)

	ptwcX := c.X + c.W/2
	ptwcY := c.Y + textOffsetY
	printTextWithColors(pdf,
		ptwcX,
		ptwcY,
		fontSize,
		[]string{c.Text},
		0,
		c.ImageWord.TextColors, cfg.Text.Color,
		config.TextAlignCenter,
	)
}

// printCellDefinition prints a cell with a text/definition wrapped and centered
func printCellDefinition(pdf *gopdf.GoPdf, cfg config.PDF, c draw.PictoCell, fontSize float64) {
	if strings.Trim(c.Def.Text, " ") == "" {
		return
	}

	newFontSize := cfg.Text.Definitions.Size
	if c.Def.Size > 0 {
		newFontSize = c.Def.Size
	}
	if newFontSize == 0 {
		newFontSize = fontSize
	}

	if c.Def.Font != "" {
		pdf.AddTTFFont(path.Base(c.Def.Font), c.Def.Font)
		pdf.SetFont(path.Base(c.Def.Font), "", newFontSize)
	} else {
		pdf.SetFont(fontFamilyNameDefinitions, "", newFontSize)
	}

	//textHeight := gopdf.ContentObjCalTextHeightPrecise(newFontSize)
	defaultColor := cfg.Text.Color
	if !c.Def.Color.IsBlack() {
		defaultColor = c.Def.Color
	}

	lines, err := pdf.SplitTextWithWordWrap(c.Def.Text, c.W-cfg.Page.Paddings.LeftRight())
	if err != nil {
		if err != nil {
			log.Error().Err(err).
				Str("line", c.Def.Text).
				Msg("unable to word wrap text")
		}
	}

	if c.Def.LineSpacingRatio == 0 && len(lines) > 1 {
		log.Warn().
			Str("text", fmt.Sprintf("%.30s...", c.Def.Text)).
			Msg("lineSpacingRatio is zero")
	}

	ptwcX := c.X + c.W/2
	ptwcY := c.Y + c.H/2
	if c.Def.Definition.Align == config.TextAlignLeft {
		ptwcX = c.X
	}
	printTextWithColors(pdf,
		ptwcX,
		ptwcY,
		newFontSize,
		lines,
		c.Def.LineSpacingRatio,
		c.Def.TextColors,
		defaultColor,
		c.Def.Definition.Align,
	)
}

// printTextWithColors prints lines of text that have been wrapped beforehand
// x should be in the center of the cell so text is centered, the true x will be calculated taken into account the real width of each line
// y is the y coordinate of the **center of the text block**. Each lines will be spaced depending on line height and the given font size
// So the whole text will be written so that y is in its center.
// If there is one line, y is used as is
func printTextWithColors(pdf *gopdf.GoPdf, x, y float64, fontSize float64, textLines []string, lineSpacingRatio float64, colors config.TextColors, defaultColor config.Color, textAlign config.TextAlign) {
	if len(textLines) == 0 {
		return
	}

	pdf.SetFontSize(fontSize)

	extraSpaceBetweenLines := fontSize * lineSpacingRatio
	textHeight := gopdf.ContentObjCalTextHeightPrecise(fontSize) + extraSpaceBetweenLines*2
	if len(textLines) > 1 {
		// printing lines from the bottom left, so we need to actually subtract 1 line which will be printed above cursor
		y -= (float64(len(textLines)-1) * textHeight) / 2
	}

	charPos := 0
	for j, line := range textLines {
		// Wrapping text removes spaces from original text
		if j > 0 {
			charPos++
		}
		textWidth, err := pdf.MeasureTextWidth(line)
		if err != nil {
			log.Error().Err(err).
				Str("line", line).
				Msg("unable to calculate text width")
		}
		if textAlign == config.TextAlignLeft {
			pdf.SetX(x)
		} else {
			pdf.SetX(x - textWidth/2)
		}
		pdf.SetY(y + float64(j)*textHeight)
		for _, char := range line {
			color, ok := colors[charPos]
			if !ok {
				color = defaultColor
			}

			pdf.SetTextColor(color.AsUints())
			err = pdf.Text(string(char))
			if err != nil {
				log.Error().Err(err).
					Str("char", string(char)).
					Msg("Error adding char to PDF")
			}

			charPos++
		}
	}
}

func getImageDimension(imagePath string) (float64, float64, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().Err(err).Msg("unable to close file")
		}
	}(file)

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return float64(img.Width), float64(img.Height), nil
}

// printCutLines prints cut lines with an offset on the left (to be able to align two-sided prints horizontally)
func printCutLines(pdf *gopdf.GoPdf, cfg config.PDF, offsetX, offsetY float64) {
	var x float64

	pdf.SetLineWidth(1)
	pdf.SetLineType("dotted")

	width := (pageSize.W - cfg.Page.PageMargins.LeftRight()) / float64(cfg.Page.Cols)
	height := (pageSize.H - cfg.Page.PageMargins.TopBottom()) / float64(cfg.Page.Lines)

	for i := 1; i < cfg.Page.Cols; i++ {
		x = cfg.Page.PageMargins.Left() + float64(i)*width + offsetX
		pdf.Line(x, 0, x, pageSize.H)
		x += width
	}

	for i := 1; i < cfg.Page.Lines; i++ {
		y := cfg.Page.PageMargins.Top() + float64(i)*height + offsetY
		pdf.Line(0, y, pageSize.W, y)
		y += height
	}
}

func setMaxFontSize(pdf *gopdf.GoPdf, text string, maxWidth, maxHeight float64) float64 {
	fontSize := 110
	inc := -1
	for {
		fontSize += inc

		if fontSize < int(math.Abs(float64(inc))) {
			break // no size found
		}

		err := pdf.SetFont(fontFamilyNameText, "", fontSize)
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to enable font")
		}
		pdf.SetFontSize(float64(fontSize))

		textWidth, _ := pdf.MeasureTextWidth(text)
		textHeight := gopdf.ContentObjCalTextHeight(fontSize)
		//textHeightWithMargin := textHeight + textHeight*70/100
		if textWidth < maxWidth && textHeight < maxHeight {
			// Height does not take accents and letters like p, q, etc.
			// Taking 50% size because why not 🤷‍
			newFontSize := float64(fontSize) * 0.5
			fontSize = int(newFontSize)
			err := pdf.SetFont(fontFamilyNameText, "", fontSize)
			if err != nil {
				log.Fatal().Msg("unable to enable font")
			}

			log.Debug().
				Int("size", fontSize).
				Msg("Setting font size")
			break
		}
	}

	return float64(fontSize)
}

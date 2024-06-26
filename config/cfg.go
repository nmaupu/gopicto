package config

import "math"

const (
	Portrait        = Orientation("portrait")
	Landscape       = Orientation("landscape")
	TextAlignCenter = TextAlign("center")
	TextAlignLeft   = TextAlign("left")
)

var (
	DefaultPageMargins = Margins{float64ptr(15), float64ptr(15), float64ptr(15), float64ptr(15)}
	DefaultMargins     = Margins{float64ptr(2.835), float64ptr(2.835), float64ptr(2.835), float64ptr(2.835)}
	DefaultPaddings    = Margins{float64ptr(3), float64ptr(3), float64ptr(3), float64ptr(3)}
	DefaultTextAlign   = TextAlignCenter
)

type Orientation string

type TextAlign string

type TextColors map[int]Color

type PDF struct {
	Page       Page        `mapstructure:"page"`
	Text       Text        `mapstructure:"text"`
	ImageWords []ImageWord `mapstructure:"images"`
}

func (p PDF) GetNbPictoPages() int {
	return int(math.Ceil(float64(len(p.ImageWords)) / float64(p.Page.Cols*p.Page.Lines)))
}

type Page struct {
	TwoSidedOffsetMM struct {
		X float64 `mapstructure:"x"`
		Y float64 `mapstructure:"y"`
	} `mapstructure:"twoSidedOffsetMM"`
	Cols        int         `mapstructure:"cols"`
	Lines       int         `mapstructure:"lines"`
	Orientation Orientation `mapstructure:"orientation"`
	Margins     Margins     `mapstructure:"margins"`
	Paddings    Margins     `mapstructure:"paddings"`
	PageMargins Margins     `mapstructure:"page_margins"`
}

type Margins struct {
	T *float64 `mapstructure:"top"`
	B *float64 `mapstructure:"bottom"`
	L *float64 `mapstructure:"left"`
	R *float64 `mapstructure:"right"`
}

func (m *Margins) InitWithDefaults(defaults Margins) {
	if m.T == nil {
		m.T = defaults.T
	}
	if m.B == nil {
		m.B = defaults.B
	}
	if m.L == nil {
		m.L = defaults.L
	}
	if m.R == nil {
		m.R = defaults.R
	}
}

func (m Margins) Top() float64 {
	if m.T == nil {
		return 0
	}
	return *m.T
}
func (m Margins) Bottom() float64 {
	if m.B == nil {
		return 0
	}
	return *m.B
}
func (m Margins) Left() float64 {
	if m.L == nil {
		return 0
	}
	return *m.L
}
func (m Margins) Right() float64 {
	if m.R == nil {
		return 0
	}
	return *m.R
}

func (m Margins) LeftRight() float64 {
	return m.Left() + m.Right()
}
func (m Margins) TopBottom() float64 {
	return m.Top() + m.Bottom()
}

type ImageWord struct {
	Image      string     `mapstructure:"image"`
	Text       string     `mapstructure:"text"`
	TextColors TextColors `mapstructure:"textColors"`
	Def        struct {
		Definition `mapstructure:",squash"`
		Text       string     `mapstructure:"text"`
		TextColors TextColors `mapstructure:"textColors"`
	} `mapstructure:"def"`
}

type Text struct {
	Font        string     `mapstructure:"font"`
	Ratio       float64    `mapstructure:"ratio"`
	FontSize    float64    `mapstructure:"size"`
	Color       Color      `mapstructure:"color"`
	Top         bool       `mapstructure:"top"`
	Definitions Definition `mapstructure:"definitions"`
}

type Definition struct {
	Borders          bool      `mapstructure:"borders"`
	Font             string    `mapstructure:"font"`
	Size             float64   `mapstructure:"size"`
	Color            Color     `mapstructure:"color"`
	LineSpacingRatio float64   `mapstructure:"lineSpacingRatio"`
	Align            TextAlign `mapstructure:"align"`
}

type Color struct {
	R, G, B uint8
}

func (c Color) IsBlack() bool {
	return c.R == 0 && c.G == 0 && c.B == 0
}

func (c Color) Equals(col Color) bool {
	return c.R == col.R &&
		c.G == col.G &&
		c.B == col.B
}

func (c Color) Red() uint8 {
	return c.R
}

func (c Color) Green() uint8 {
	return c.G
}

func (c Color) Blue() uint8 {
	return c.B
}

func (c Color) AsUints() (uint8, uint8, uint8) {
	return c.R, c.G, c.B
}

func float64ptr(f float64) *float64 {
	return &f
}

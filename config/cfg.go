package config

const (
	Portrait  = Orientation("portrait")
	Landscape = Orientation("landscape")
)

var (
	DefaultPageMargins = Margins{float64ptr(15), float64ptr(15), float64ptr(15), float64ptr(15)}
	DefaultMargins     = Margins{float64ptr(2.835), float64ptr(2.835), float64ptr(2.835), float64ptr(2.835)}
	DefaultPaddings    = Margins{float64ptr(3), float64ptr(3), float64ptr(3), float64ptr(3)}
)

type Orientation string

type PDF struct {
	Page       Page        `mapstructure:"page"`
	Text       Text        `mapstructure:"text"`
	ImageWords []ImageWord `mapstructure:"images"`
}

type Page struct {
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
	Image string `mapstructure:"image"`
	Text  string `mapstructure:"text"`
}

type Text struct {
	Font             string  `mapstructure:"font"`
	Ratio            float64 `mapstructure:"ratio"`
	Color            Color   `mapstructure:"color"`
	FirstLetterColor Color   `mapstructure:"firstLetterColor"`
	Top              bool    `mapstructure:"top"`
}

type Color struct {
	R, G, B uint8
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

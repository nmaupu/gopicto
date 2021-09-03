package config

const (
	Portrait  = Orientation("portrait")
	Landscape = Orientation("landscape")
)

type Orientation string

type PDF struct {
	Cols        int         `mapstructure:"cols"`
	Lines       int         `mapstructure:"lines"`
	Orientation Orientation `mapstructure:"orientation"`
	Text        Text        `mapstructure:"text"`
	ImageWords  []ImageWord `mapstructure:"images"`
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

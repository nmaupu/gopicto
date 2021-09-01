package config

type PDF struct {
	Lines int `mapstructure:"lines"`
	Cols int `mapstructure:"cols"`
	Images []ImageWord `mapstructure:"images"`
}

type ImageWord struct {
	Image string `mapstructure:"image"`
	Text string `mapstructure:"text"`
}
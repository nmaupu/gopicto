package config

type PDF struct {
	Cols   int         `mapstructure:"cols"`
	Images []ImageWord `mapstructure:"images"`
}

type ImageWord struct {
	Image string `mapstructure:"image"`
	Text  string `mapstructure:"text"`
}

package cli

import (
	"github.com/mitchellh/mapstructure"
	"github.com/nmaupu/gopicto/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
)

const (
	AppName = "gopicto"

	ConfigFlag = "config"
	OutputFlag = "output"
)

var (
	ViperConfigKey = "gopictoConfig"

	defaultConfigFile = "./config.yaml"
	defaultOutputFile = "/tmp/gopicto.pdf"

	cfgFile string
	outFile string
)

var RootCmd = &cobra.Command{
	Use:   AppName,
	Short: "Picto and associated word PDF generator",
	Run:   mainCmd,
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVarP(&cfgFile, ConfigFlag, "c", defaultConfigFile, "Config file to use")
	RootCmd.PersistentFlags().StringVarP(&outFile, OutputFlag, "o", defaultOutputFile, "Specify the name of the file generated")
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

	cfg.Page.Margins.InitWithDefaults(config.DefaultMargins)
	cfg.Page.Paddings.InitWithDefaults(config.DefaultPaddings)

	if cfg.Text.Ratio == 0.0 {
		cfg.Text.Ratio = 1 / 5
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

		RootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
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

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("An error occurred")
	}
}

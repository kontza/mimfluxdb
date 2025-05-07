/*
Copyright © 2024 Juha Ruotsalainen <juha.ruotsalainen@iki.fi>
*/
package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

type DatabaseConnection struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
}

type AppConfig struct {
	cfgFile             string
	Tokens              []string
	LogLevel            string             `mapstructure:"log_level"`
	LogWithoutTimestamp bool               `mapstructure:"log_without_timestamp"`
	StructuredLogging   bool               `mapstructure:"structured_logging"`
	DatabaseConnection  DatabaseConnection `mapstructure:"database_connection"`
}

var appConfig AppConfig

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Short: "Provides a write interface looking like InfluxDB, but stores data into a PostgreSQL instance",
	Run:   rootRunner,
}

func Bootstrap(name string, version string) {
	rootCmd.Version = version
	if strings.TrimSpace(name) == "" {
		name = "mimfluxdb"
	}
	rootCmd.Use = name
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := setUpLogs(appConfig.StructuredLogging); err != nil {
			return err
		}
		return nil
	}

	rootCmd.PersistentFlags().StringVar(&appConfig.cfgFile, "config", "", fmt.Sprintf("config file (default is $XDG_CONFIG_HOME/%s/config.toml)", rootCmd.Use))
	rootCmd.PersistentFlags().BoolVarP(&appConfig.StructuredLogging, "structured-logging", "s", false, "Use structured logging")
	rootCmd.PersistentFlags().BoolVarP(&appConfig.LogWithoutTimestamp, "log-without-timestamp", "w", false, "Do not print timestamp (applies to non-structured logging)")
	rootCmd.PersistentFlags().StringVarP(&appConfig.LogLevel, "log-level", "l", "info", "Logging level to use")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if appConfig.cfgFile != "" {
		if configPath, err := filepath.Abs(appConfig.cfgFile); err != nil {
			log.Error().Str("cfgFile", appConfig.cfgFile).Err(err).Msg("Failed to get absolute path for")
			os.Exit(1)
		} else {
			viper.SetConfigFile(configPath)
		}
	} else {
		cfgDir := os.Getenv("XDG_CONFIG_HOME")
		if len(cfgDir) == 0 {
			log.Warn().Msg("$XDG_CONFIG_HOME is not defined, falling back to '$HOME/.config'.")
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			cfgDir = path.Join(home, ".config", rootCmd.Use)
		} else {
			cfgDir = path.Join(cfgDir, rootCmd.Use)
		}
		viper.AddConfigPath(cfgDir)
		viper.SetConfigType("toml")
		viper.SetConfigName("config.toml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Unmarshal(&appConfig); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal config")
			os.Exit(1)
		}
	} else {
		log.Warn().Err(err).Msg("Failed to read in config")
	}
}

func setUpLogs(structuredLogging bool) error {
	if !structuredLogging {
		if appConfig.LogWithoutTimestamp {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, PartsExclude: []string{"time"}})
		} else {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMilli})
		}
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
	return nil
}

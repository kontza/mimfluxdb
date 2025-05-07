package cmd

import (
	"net/http"

	"github.com/Xiol/zerochi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var db *gorm.DB

func rootRunner(cmd *cobra.Command, args []string) {
	switch appConfig.LogLevel {
	case "debug":
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	case "info":
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	case "warn":
		log.Logger = log.Logger.Level(zerolog.WarnLevel)
	case "error":
		log.Logger = log.Logger.Level(zerolog.ErrorLevel)
	case "fatal":
		log.Logger = log.Logger.Level(zerolog.FatalLevel)
	case "panic":
		log.Logger = log.Logger.Level(zerolog.PanicLevel)
	case "disabled":
		log.Logger = log.Logger.Level(zerolog.Disabled)
	}
	db = openDb()
	r := chi.NewRouter()
	r.Use(middleware.RequestID,
		middleware.RealIP,
		zerochi.Logger(&log.Logger),
		middleware.Recoverer)
	r.Post("/api/v2/write", writeHandler)
	log.Info().Msg("Server started on :8086")
	http.ListenAndServe(":8086", r)
}

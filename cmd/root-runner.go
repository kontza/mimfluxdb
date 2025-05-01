package cmd

import (
	"net/http"

	"github.com/Xiol/zerochi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func rootRunner(cmd *cobra.Command, args []string) {
	r := chi.NewRouter()
	r.Use(middleware.RequestID,
		middleware.RealIP,
		zerochi.Logger(&log.Logger),
		middleware.Recoverer)
	r.Post("/api/v2/write", writeHandler)
	log.Info().Msg("Server started on :8086")
	http.ListenAndServe(":8086", r)
}

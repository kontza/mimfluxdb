package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type ParseResponse struct {
	HTTPStatusCode int    `json:"-"`      // http response status code
	StatusText     string `json:"status"` // user-level status message
}

func (pr *ParseResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read body:")
		render.Render(w, r, errInvalidRequest(err))
	}
	validToken := false
	for _, token := range r.Header["Authorization"] {
		for _, expected := range appConfig.Tokens {
			if fmt.Sprintf("Token %s", expected) == token {
				validToken = true
				break
			}
		}
		if validToken {
			break
		}
	}
	if !validToken {
		render.Render(w, r, errUnauthorized())
	}
	log.Info().Interface("header", r.Header).Str("body", string(body)).Msg("Received")
	render.Render(w, r, parseBody(body))
}

func parseBody(body []byte) *ParseResponse {
	payload := string(body)
	firstSpace := strings.Index(payload, " ")
	lastSpace := strings.LastIndex(payload, " ")
	locationPart := strings.TrimSpace(payload[:firstSpace])
	tagPart := strings.TrimSpace(payload[firstSpace:lastSpace])
	timestampPart := strings.TrimSpace(payload[lastSpace:])
	log.Info().
		Str("locationPart", locationPart).
		Str("tagPart", tagPart).
		Str("timestampPart", timestampPart).
		Msg("Parsed")
	return &ParseResponse{http.StatusOK, "OK"}
}

package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type ParseResponse struct {
	HTTPStatusCode int    `json:"-"`      // http response status code
	StatusText     string `json:"status"` // user-level status message
}

type MeasurementData struct {
	Location    string
	Temperature float64
	Rssi        int
	Count       int
	Device      string
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
	firstComma := strings.Index(payload, ",")
	lastSpace := strings.LastIndex(payload, " ")
	data := payload[firstComma+1 : lastSpace]
	bySpaces := strings.Split(data, " ")
	measurementData := MeasurementData{}
	for _, part := range bySpaces {
		byCommas := strings.Split(part, ",")
		var err error
		for _, subPart := range byCommas {
			equals := strings.Index(subPart, "=")
			switch subPart[:equals] {
			case "location":
				measurementData.Location = subPart[equals+1:]
			case "count":
				measurementData.Count, err = strconv.Atoi(subPart[equals+1:])
				if err != nil {
					log.Error().Err(err).Msg("Failed to parse count")
				}
			case "device":
				measurementData.Device = subPart[equals+1:]
			case "rssi":
				measurementData.Rssi, err = strconv.Atoi(subPart[equals+1:])
				if err != nil {
					log.Error().Err(err).Msg("Failed to parse rssi")
				}
			case "temperature":
				measurementData.Temperature, err = strconv.ParseFloat(subPart[equals+1:], 64)
				if err != nil {
					log.Error().Err(err).Msg("Failed to parse rssi")
				}
			}
		}
	}
	log.Info().Interface("measurementData", measurementData).Msg("Parsed")
	return &ParseResponse{http.StatusOK, "OK"}
}

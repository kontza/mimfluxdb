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

const TEMPERATURE_FIELD = "temperature"
const LOCATION_FIELD = "location"

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
	data := parseBody(body)
	render.Render(w, r, storeData(data))
}

func parseBody(body []byte) map[string]any {
	payload := string(body)
	firstComma := strings.Index(payload, ",")
	lastSpace := strings.LastIndex(payload, " ")
	data := payload[firstComma+1 : lastSpace]
	bySpaces := strings.Split(data, " ")
	measurementData := make(map[string]any)
	measurementData["DATABASE"] = payload[:firstComma]
	integerKeys := []string{"rssi", "count"}
	floatKeys := []string{TEMPERATURE_FIELD}
	for _, part := range bySpaces {
		byCommas := strings.Split(part, ",")
		for _, subPart := range byCommas {
			equals := strings.Index(subPart, "=")
			key := subPart[:equals]
			value := subPart[equals+1:]
			valueProcessed := false
			for _, integerKey := range integerKeys {
				if key == integerKey {
					intValue, err := strconv.Atoi(value)
					if err != nil {
						log.Error().Err(err).Str("key", key).Str("value", value).Msg("Failed to convert to int")
						break
					}
					measurementData[key] = intValue
					valueProcessed = true
				}
			}
			if !valueProcessed {
				for _, floatKey := range floatKeys {
					if key == floatKey {
						floatValue, err := strconv.ParseFloat(value, 64)
						if err != nil {
							log.Error().Err(err).Str("key", key).Str("value", value).Msg("Failed to convert to float")
						}
						measurementData[key] = floatValue
						valueProcessed = true
					}
				}
			}
			if !valueProcessed {
				measurementData[key] = value
			}
		}
	}
	log.Info().Interface("measurementData", measurementData).Msg("Parsed")
	return measurementData
}

func storeData(data map[string]any) *ParseResponse {
	temperatureValue, temperatureExists := data[TEMPERATURE_FIELD].(float64)
	locationValue, locationExists := data[LOCATION_FIELD].(string)
	var fields []string
	if !temperatureExists {
		fields = append(fields, TEMPERATURE_FIELD)
	}
	if !locationExists {
		fields = append(fields, LOCATION_FIELD)
	}
	if len(fields) > 0 {
		statusText := fmt.Sprintf("Missing required fields: %s", strings.Join(fields, ", "))
		return &ParseResponse{http.StatusBadRequest, statusText}
	}
	log.Info().Str(LOCATION_FIELD, locationValue).Float64(TEMPERATURE_FIELD, temperatureValue).Msg("Storing data")
	return &ParseResponse{http.StatusOK, "OK"}
}

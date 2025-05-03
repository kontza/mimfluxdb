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

const COUNT_FIELD = "count"
const DEVICE_FIELD = "device"
const LOCATION_FIELD = "location"
const RSSI_FIELD = "rssi"
const TEMPERATURE_FIELD = "temperature"
const TIMESTAMP_FIELD = "__TIMESTAMP"

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
	timestampString := strings.TrimSpace(payload[lastSpace+1:])
	if timestampValue, err := strconv.ParseInt(timestampString, 10, 64); err != nil {
		log.Error().Err(err).Str("timestamp", timestampString).Msg("Failed to convert")
	} else {
		measurementData[TIMESTAMP_FIELD] = timestampValue
	}
	integerKeys := []string{RSSI_FIELD, COUNT_FIELD}
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
	log.Info().Interface("map", data).Msg("Storing")
	timestampValue, timestampExists := data[TIMESTAMP_FIELD].(int)
	temperatureValue, temperatureExists := data[TEMPERATURE_FIELD].(float64)
	locationValue, locationExists := data[LOCATION_FIELD].(string)
	countValue, countExists := data[COUNT_FIELD].(int)
	rssiValue, rssiExists := data[RSSI_FIELD].(int)
	deviceValue, deviceExists := data[DEVICE_FIELD].(string)
	var fields []string
	if !timestampExists {
		fields = append(fields, TIMESTAMP_FIELD)
	} else {
		log.Info().Int(TIMESTAMP_FIELD, timestampValue).Msg(toTitleCase(TIMESTAMP_FIELD))
	}
	if !deviceExists {
		fields = append(fields, DEVICE_FIELD)
	} else {
		log.Info().Str(DEVICE_FIELD, deviceValue).Msg(toTitleCase(DEVICE_FIELD))
	}
	if !temperatureExists {
		fields = append(fields, TEMPERATURE_FIELD)
	} else {
		log.Info().Float64(TEMPERATURE_FIELD, temperatureValue).Msg(toTitleCase(TEMPERATURE_FIELD))
	}
	if len(fields) > 0 {
		statusText := fmt.Sprintf("Missing required fields: %s", strings.Join(fields, ", "))
		log.Error().Msg(statusText)
		return &ParseResponse{http.StatusBadRequest, statusText}
	}
	if locationExists {
		log.Info().Str(LOCATION_FIELD, locationValue).Msg(toTitleCase(LOCATION_FIELD))
	}
	if countExists {
		log.Info().Int(COUNT_FIELD, countValue).Msg(toTitleCase(COUNT_FIELD))
	}
	if rssiExists {
		log.Info().Int(RSSI_FIELD, rssiValue).Msg(toTitleCase(RSSI_FIELD))
	}
	return &ParseResponse{http.StatusOK, "OK"}
}

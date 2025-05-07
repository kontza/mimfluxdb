package cmd

import (
	"context"
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
	render.Render(w, r, storeData(r.Context(), data))
}

func getToken(data string, token string) string {
	token = fmt.Sprintf("%s=", token)
	log.Debug().Str("token", token).Str("data", data).Msg("Getting token")
	startIndex := strings.Index(data, token)
	if startIndex == -1 {
		log.Error().Str("token", token).Msg("Did not find token")
		return ""
	}
	log.Debug().Str("start", data[startIndex:]).Msg("Start")
	remainingData := data[startIndex+len(token):]
	nextIndex := strings.Index(remainingData, "=")
	log.Debug().Str("remaining data", remainingData).Msg("Search next token from")
	if nextIndex == -1 {
		return strings.Trim(remainingData, " \"")
	} else {
		nextEquals := strings.Index(remainingData, "=")
		log.Debug().Int("nextEquals", nextEquals).Msg("Found")
		log.Debug().Str("slice", remainingData[:nextEquals]).Msg("Next")
		lastSpace := strings.LastIndex(remainingData[:nextEquals], " ")
		lastComma := strings.LastIndex(remainingData[:nextEquals], ",")
		cutPoint := max(lastSpace, lastComma)
		log.Debug().Int("cutPoint", cutPoint).Int("lastSpace", lastSpace).Int("lastComma", lastComma).Msg("Cut point")
		return strings.Trim(remainingData[:cutPoint], " \"")
	}
}

func parseInt(value string, key string) (int, error) {
	if intValue, err := strconv.Atoi(strings.TrimSuffix(value, "i")); err == nil {
		return intValue, nil
	}
	return 0, fmt.Errorf("'%s': Failed to convert '%s' to int", key, value)
}

func parseTemperature(value string, key string) (float64, error) {
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue, nil
	}
	return 0, fmt.Errorf("'%s': Failed to convert '%s' to float", key, value)
}

func parseBody(body []byte) map[string]any {
	bodyString := string(body)
	firstComma := strings.Index(bodyString, ",")
	lastSpace := strings.LastIndex(bodyString, " ")
	payload := bodyString[firstComma+1 : lastSpace]
	measurementData := make(map[string]any)
	timestampString := strings.TrimSpace(bodyString[lastSpace+1:])
	if timestampValue, err := strconv.ParseInt(timestampString, 10, 64); err != nil {
		log.Error().Err(err).Str("timestamp", timestampString).Msg("Failed to convert")
	} else {
		measurementData[TIMESTAMP_FIELD] = timestampValue
	}
	field := COUNT_FIELD
	receivedValue := getToken(payload, field)
	if strings.TrimSpace(receivedValue) != "" {
		if parsedInt, err := parseInt(receivedValue, field); err != nil {
			log.Error().Err(err).Msg("Error:")
		} else {
			measurementData[field] = parsedInt
		}
	}
	field = RSSI_FIELD
	receivedValue = getToken(payload, field)
	if strings.TrimSpace(receivedValue) != "" {
		if parsedInt, err := parseInt(receivedValue, field); err != nil {
			log.Error().Err(err).Msg("Error:")
		} else {
			measurementData[field] = parsedInt
		}
	}
	field = TEMPERATURE_FIELD
	receivedValue = getToken(payload, field)
	if strings.TrimSpace(receivedValue) != "" {
		if parsedTemperature, err := parseTemperature(receivedValue, field); err != nil {
			log.Error().Err(err).Msg("Error:")
		} else {
			measurementData[field] = parsedTemperature
		}
	}
	field = DEVICE_FIELD
	receivedValue = getToken(payload, field)
	if strings.TrimSpace(receivedValue) != "" {
		measurementData[field] = receivedValue
	}
	field = LOCATION_FIELD
	receivedValue = getToken(payload, field)
	if strings.TrimSpace(receivedValue) != "" {
		measurementData[field] = receivedValue
	}
	log.Info().Interface("measurementData", measurementData).Msg("Parsed")
	return measurementData
}

func storeData(ctx context.Context, data map[string]any) *ParseResponse {
	log.Info().Interface("map", data).Msg("Storing")
	timestampValue, timestampExists := data[TIMESTAMP_FIELD].(int64)
	temperatureValue, temperatureExists := data[TEMPERATURE_FIELD].(float64)
	locationValue, _ := data[LOCATION_FIELD].(string)
	countValue, countExists := data[COUNT_FIELD].(int)
	rssiValue, rssiExists := data[RSSI_FIELD].(int)
	deviceValue, deviceExists := data[DEVICE_FIELD].(string)
	var fields []string
	if !timestampExists {
		fields = append(fields, TIMESTAMP_FIELD)
	}
	if !deviceExists {
		fields = append(fields, DEVICE_FIELD)
	}
	if !temperatureExists {
		fields = append(fields, TEMPERATURE_FIELD)
	}
	if len(fields) > 0 {
		statusText := fmt.Sprintf("Missing required fields: %s", strings.Join(fields, ", "))
		log.Error().Msg(statusText)
		return &ParseResponse{http.StatusBadRequest, statusText}
	}
	device := getDevice(ctx, deviceValue, locationValue)
	if countExists {
		storeCount(ctx, timestampValue, device, countValue)
	}
	if rssiExists {
		storeRssi(ctx, timestampValue, device, rssiValue)
	}
	storeTemperature(ctx, timestampValue, device, temperatureValue)
	return &ParseResponse{http.StatusOK, "OK"}
}

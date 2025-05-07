package cmd

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
)

const MEASUREMENT_NAME = "temperatures"
const DEVICE_VALUE = "28deadbeef"
const LOCATION_VALUE = "Death Valley"
const COUNT_VALUE = 69
const RSSI_VALUE = -42
const TEMPERATURE_VALUE = math.Pi
const TIMESTAMP_VALUE = 1746503561000000000

var testData []byte
var tokenData []byte

func setup(t *testing.T) {
	testData = []byte(MEASUREMENT_NAME + ",")
	testData = fmt.Appendf(testData, "%s=%s", LOCATION_FIELD, LOCATION_VALUE)
	testData = fmt.Appendf(testData, " %s=%s", DEVICE_FIELD, DEVICE_VALUE)
	testData = fmt.Appendf(testData, ",%s=%di", COUNT_FIELD, COUNT_VALUE)
	testData = fmt.Appendf(testData, ",%s=%di", RSSI_FIELD, RSSI_VALUE)
	testData = fmt.Appendf(testData, ",%s=%g", TEMPERATURE_FIELD, TEMPERATURE_VALUE)
	testData = fmt.Appendf(testData, " %d", TIMESTAMP_VALUE)
	tokenData = []byte("")
	tokenData = fmt.Appendf(tokenData, "%s=%s", LOCATION_FIELD, LOCATION_VALUE)
	tokenData = fmt.Appendf(tokenData, " %s=%c%s%c", DEVICE_FIELD, '"', LOCATION_VALUE, '"')
	tokenData = fmt.Appendf(tokenData, ",%s=%di", COUNT_FIELD, COUNT_VALUE)
	tokenData = fmt.Appendf(tokenData, ",%s=%di", RSSI_FIELD, RSSI_VALUE)
	tokenData = fmt.Appendf(tokenData, ",%s=%g", TEMPERATURE_FIELD, TEMPERATURE_VALUE)
	setUpLogs(false)
	t.Helper()
}

func TestGetToken(t *testing.T) {
	setup(t)
	tests := []struct {
		name     string
		data     string
		token    string
		expected string
	}{
		{
			name:     "Location with spaces",
			data:     string(tokenData),
			token:    LOCATION_FIELD,
			expected: LOCATION_VALUE,
		},
		{
			name:     "Count",
			data:     string(tokenData),
			token:    COUNT_FIELD,
			expected: fmt.Sprintf("%di", COUNT_VALUE),
		},
		{
			name:     "RSSI from between commas",
			data:     string(tokenData),
			token:    RSSI_FIELD,
			expected: fmt.Sprintf("%di", RSSI_VALUE),
		},
		{
			name:     "Quoted device with spaces",
			data:     string(tokenData),
			token:    DEVICE_FIELD,
			expected: LOCATION_VALUE,
		}, {
			name:     "Temperature",
			data:     string(tokenData),
			token:    TEMPERATURE_FIELD,
			expected: fmt.Sprintf("%f", TEMPERATURE_VALUE),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getToken(tt.data, tt.token)
			log.Debug().Str("result", result).Msg("Result")
			if result != tt.expected {
				t.Errorf("getToken() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseBody(t *testing.T) {
	setup(t)
	tests := []struct {
		name     string
		body     []byte
		expected map[string]any
	}{
		{
			name: "Valid input",
			body: testData,
			expected: map[string]any{
				COUNT_FIELD:       COUNT_VALUE,
				DEVICE_FIELD:      DEVICE_VALUE,
				LOCATION_FIELD:    LOCATION_VALUE,
				RSSI_FIELD:        RSSI_VALUE,
				TEMPERATURE_FIELD: TEMPERATURE_VALUE,
				TIMESTAMP_FIELD:   TIMESTAMP_VALUE,
			},
		}, {
			name:     "Invalid input",
			body:     []byte("invalid data"),
			expected: make(map[string]any),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBody(tt.body)
			if strings.HasPrefix(tt.name, "Invalid") {
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("parseBody() = %v, want %v", result, tt.expected)
				}
			} else {
				for key, value := range tt.expected {
					expected := fmt.Sprintf("%v", value)
					resultValue := fmt.Sprintf("%v", result[key])
					if expected != resultValue {
						t.Errorf("parseBody(): key = %v, value = %v, want = %v", key, result[key], value)
					}
				}
			}
		})
	}
}

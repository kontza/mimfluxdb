package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func writeHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read body:")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Internal Server Error",
			"detail": err.Error(),
		})
		return
	}
	validToken := false
	for _, token := range c.Request.Header["Authorization"] {
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	log.Info().Interface("header", c.Request.Header).Str("body", string(body)).Msg("Received")
	status, payload := parseBody(string(body))
	c.JSON(status, payload)
}

func parseBody(payload string) (int, gin.H) {
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
	return http.StatusOK, gin.H{"message": "OK"}
}

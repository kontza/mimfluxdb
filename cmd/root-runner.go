package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func rootRunner(cmd *cobra.Command, args []string) {
	r := gin.New()
	r.Use(Logger("gin"), gin.Recovery())
	r.SetTrustedProxies(nil)
	r.POST("/api/v2/write", func(c *gin.Context) {
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
		c.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})
	r.Run(":8086")
}

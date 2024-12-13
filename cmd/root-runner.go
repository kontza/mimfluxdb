package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func rootRunner(cmd *cobra.Command, args []string) {
	r := gin.New()
	r.Use(zeroLogger("gin"), gin.Recovery())
	r.SetTrustedProxies(nil)
	r.POST("/api/v2/write", writeHandler)
	r.Run(":8086")
}

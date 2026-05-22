package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CollectStatus(c *gin.Context) {
	if collectRunner == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "collector runner not initialized"})
		return
	}

	c.JSON(http.StatusOK, collectRunner.GetStatus())
}

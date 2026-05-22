package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CollectRunTemp(c *gin.Context) {
	if collectRunner == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "collector runner not initialized"})
		return
	}

	started, err := collectRunner.StartTemperatureOnly()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	if !started {
		c.JSON(http.StatusConflict, gin.H{"ok": false, "error": "collection already running"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "started": true})
}

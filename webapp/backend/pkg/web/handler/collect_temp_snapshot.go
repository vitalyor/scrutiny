package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/analogj/scrutiny/webapp/backend/pkg/database"
	"github.com/analogj/scrutiny/webapp/backend/pkg/models/collector"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const smartctlBinary = "smartctl"

func CollectTempSnapshot(c *gin.Context) {
	logger := c.MustGet("LOGGER").(*logrus.Entry)
	deviceRepo := c.MustGet("DEVICE_REPOSITORY").(database.DeviceRepo)

	devices, err := deviceRepo.GetDevices(c)
	if err != nil {
		logger.Errorln("An error occurred while retrieving devices for temp snapshot", err)
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to list devices"})
		return
	}

	ctx, cancel := context.WithTimeout(c, 2*time.Minute)
	defer cancel()

	now := time.Now().UTC()
	temps := map[string]gin.H{}

	for _, device := range devices {
		if strings.TrimSpace(device.DeviceName) == "" || device.ScrutinyUUID.IsNil() {
			continue
		}

		devicePath := device.DeviceName
		if !strings.HasPrefix(devicePath, "/") {
			devicePath = "/dev/" + devicePath
		}

		args := []string{"--attributes", "--json"}
		if len(device.DeviceType) > 0 && device.DeviceType != "scsi" && device.DeviceType != "ata" {
			args = append(args, "--device", device.DeviceType)
		}
		args = append(args, devicePath)

		cmd := exec.CommandContext(ctx, smartctlBinary, args...)
		output, cmdErr := cmd.Output()
		if cmdErr != nil {
			logger.Warnf("temp snapshot smartctl failed for %s: %v", devicePath, cmdErr)
			continue
		}

		var info collector.SmartInfo
		if unmarshalErr := json.Unmarshal(output, &info); unmarshalErr != nil {
			logger.Warnf("temp snapshot parse failed for %s: %v", devicePath, unmarshalErr)
			continue
		}

		temp := info.Temperature.Current
		if temp == 0 && info.NvmeSmartHealthInformationLog.Temperature > 0 {
			temp = info.NvmeSmartHealthInformationLog.Temperature
			if temp > 200 {
				temp = temp - 273
			}
		}
		if temp == 0 {
			continue
		}

		temps[device.ScrutinyUUID.String()] = gin.H{
			"temp": temp,
			"date": now,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
		"data": gin.H{
			"temps": temps,
			"collectedAt": fmt.Sprintf("%s", now.Format(time.RFC3339)),
		},
	})
}

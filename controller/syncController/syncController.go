package syncController

import (
	"Bea-Cukai/helper"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

type SyncController struct{}

func NewSyncController() *SyncController {
	return &SyncController{}
}

// RunSync executes the database sync script
// GET /api/sync/run
func (sc *SyncController) RunSync(c *gin.Context) {
	// Get script path from environment variable
	scriptPath := helper.GetEnv("SYNC_SCRIPT_PATH")

	if scriptPath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "SYNC_SCRIPT_PATH environment variable is not set",
		})
		return
	}

	// Execute the script with bash (for Windows Git Bash compatibility)
	cmd := exec.Command("bash", scriptPath)

	output, err := cmd.CombinedOutput()

	// Check if script returns exit code 2 (already running)
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 2 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "running",
			"message": "Sinkronisasi sedang berjalan.",
		})
		return
	}

	// Check for other errors
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
			"output":  string(output),
		})
		return
	}

	// Success
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Sinkronisasi database berhasil.",
		"output":  string(output),
	})
}

// GetSyncStatus checks if sync is currently running
// GET /api/sync/status
func (sc *SyncController) GetSyncStatus(c *gin.Context) {
	// Check if lock file exists (both production and test paths)
	lockFiles := []string{"/tmp/sync_fkk_db.lock", "/tmp/sync_fkk_db_test.lock"}

	for _, lockFile := range lockFiles {
		cmd := exec.Command("bash", "-c", "test -f "+lockFile)
		err := cmd.Run()

		if err == nil {
			// Lock file exists, sync is running
			c.JSON(http.StatusOK, gin.H{
				"status":  "running",
				"message": "Sinkronisasi sedang berjalan.",
			})
			return
		}
	}

	// Lock file doesn't exist, sync is not running
	c.JSON(http.StatusOK, gin.H{
		"status":  "idle",
		"message": "Tidak ada sinkronisasi yang sedang berjalan.",
	})
}

// GetSyncLog retrieves the latest sync log
// GET /api/sync/log
func (sc *SyncController) GetSyncLog(c *gin.Context) {
	// Try multiple log locations (production and test)
	logPaths := []string{
		"./sync_test.log",          // Local test log
		"/var/log/sync_fkk_db.log", // Production log
		"/tmp/sync_fkk_db.log",     // Alternative production log
	}

	var logFile string
	var logContent []byte
	var err error

	// Try to find and read the first available log file
	for _, path := range logPaths {
		cmd := exec.Command("bash", "-c", "test -f "+path)
		if cmd.Run() == nil {
			// File exists, try to read it
			cmd = exec.Command("bash", "-c", "tail -n 100 "+path)
			logContent, err = cmd.Output()
			if err == nil {
				logFile = path
				break
			}
		}
	}

	if logFile == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Log file tidak ditemukan.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"logFile": logFile,
		"content": string(logContent),
	})
}

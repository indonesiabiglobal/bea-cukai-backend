package pabeanController

import (
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/service/pabeanService"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PabeanController struct {
	PabeanService *pabeanService.PabeanService
}

func NewPabeanController(svc *pabeanService.PabeanService) *PabeanController {
	return &PabeanController{PabeanService: svc}
}

// ==========================
// API endpoints
// ==========================

// GET /pabean
func (c *PabeanController) GetAll(ctx *gin.Context) {
	res, err := c.PabeanService.GetAll()
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail get pabean documents", err, gin.H{})
		return
	}

	apiresponse.OK(ctx, res, "ok", gin.H{
		"count": len(res),
	})
}

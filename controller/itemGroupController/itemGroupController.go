package itemGroupController

import (
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/service/itemGroupService"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ItemGroupController struct {
	ItemGroupService *itemGroupService.ItemGroupService
}

func NewItemGroupController(svc *itemGroupService.ItemGroupService) *ItemGroupController {
	return &ItemGroupController{ItemGroupService: svc}
}

// ==========================
// API endpoints
// ==========================

// GET /item-groups
func (c *ItemGroupController) GetAll(ctx *gin.Context) {
	res, err := c.ItemGroupService.GetAll()
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail get item groups", err, gin.H{})
		return
	}

	apiresponse.OK(ctx, res, "ok", gin.H{
		"count": len(res),
	})
}

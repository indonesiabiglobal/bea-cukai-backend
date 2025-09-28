package wipPositionReportController

import (
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/service/wipPositionReportService"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WipPositionReportController struct {
	WipPositionReportService *wipPositionReportService.WipPositionReportService
}

func NewWipPositionReportController(svc *wipPositionReportService.WipPositionReportService) *WipPositionReportController {
	return &WipPositionReportController{WipPositionReportService: svc}
}

// ==========================
// Report endpoints
// ==========================

// GET /report/wip-position?from=YYYY-MM-DD&to=YYYY-MM-DD&item_code=...&item_name=...&page=1&rows=10
// Note: Using 'rows' parameter to match the PHP API convention
func (c *WipPositionReportController) GetReport(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"),
			"to":   ctx.Query("to"),
		})
		return
	}

	// Get optional filter parameters
	itemCode := ctx.Query("item_code")
	itemName := ctx.Query("item_name")

	// Get pagination parameters (using 'limit' to match PHP API)
	page := apiRequest.ParseInt(ctx, "page", 1)
	limit := apiRequest.ParseInt(ctx, "limit", 10) // Using 'limit' instead of 'rows' to match PHP

	res, totalCount, err := c.WipPositionReportService.GetReport(from, to, itemCode, itemName, page, limit)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail get WIP position report", err, gin.H{
			"from":      from.Format("2006-01-02"),
			"to":        to.Format("2006-01-02"),
			"item_code": itemCode,
			"item_name": itemName,
			"page":      page,
			"rows":      limit,
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit)) // ceil division
	hasNext := page < totalPages
	hasPrev := page > 1

	// Response format matching PHP API structure
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from":      from.Format("2006-01-02"),
		"to":        to.Format("2006-01-02"),
		"item_code": itemCode,
		"item_name": itemName,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"totalCount": totalCount,
			"totalPages": totalPages,
			"count":      len(res),
			"hasNext":    hasNext,
			"hasPrev":    hasPrev,
		},
	})
}

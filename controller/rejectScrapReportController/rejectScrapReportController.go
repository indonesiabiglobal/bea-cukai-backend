package rejectScrapReportController

import (
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/service/rejectScrapReportService"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RejectScrapReportController struct {
	RejectScrapReportService *rejectScrapReportService.RejectScrapReportService
}

func NewRejectScrapReportController(svc *rejectScrapReportService.RejectScrapReportService) *RejectScrapReportController {
	return &RejectScrapReportController{RejectScrapReportService: svc}
}

// ==========================
// Report endpoints
// ==========================

// GET /report/reject-scrap?from=YYYY-MM-DD&to=YYYY-MM-DD&item_code=...&item_name=...&page=1&limit=10
func (c *RejectScrapReportController) GetReport(ctx *gin.Context) {
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

	// Get pagination parameters
	page := apiRequest.ParseInt(ctx, "page", 1)
	limit := apiRequest.ParseInt(ctx, "limit", 10)

	res, totalCount, err := c.RejectScrapReportService.GetReport(from, to, itemCode, itemName, page, limit)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail to get reject and scrap report", err, gin.H{
			"from":      from.Format("2006-01-02"),
			"to":        to.Format("2006-01-02"),
			"item_code": itemCode,
			"item_name": itemName,
			"page":      page,
			"limit":     limit,
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

package expenditureProductController

import (
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/service/expenditureProductService"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExpenditureProductController struct {
	ExpenditureProductService *expenditureProductService.ExpenditureProductService
}

func NewExpenditureProductController(svc *expenditureProductService.ExpenditureProductService) *ExpenditureProductController {
	return &ExpenditureProductController{ExpenditureProductService: svc}
}

// ==========================
// Report endpoints
// ==========================

// GET /report/expenditure-products?from=YYYY-MM-DD&to=YYYY-MM-DD&pabeanType=...&productGroup=...&noPabean=...&productCode=...&productName=...&page=1&limit=10
func (c *ExpenditureProductController) GetReport(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"),
			"to":   ctx.Query("to"),
		})
		return
	}

	// Get optional filter parameters
	pabeanType := ctx.Query("pabeanType")
	productGroup := ctx.Query("productGroup")
	noPabean := ctx.Query("noPabean")
	productCode := ctx.Query("productCode")
	productName := ctx.Query("productName")

	// Get pagination parameters
	page := apiRequest.ParseInt(ctx, "page", 1)    // default to page 1 if not provided or invalid
	limit := apiRequest.ParseInt(ctx, "limit", 10) // default to 10 items per page if not provided or invalid

	res, totalCount, err := c.ExpenditureProductService.GetReport(from, to, pabeanType, productGroup, noPabean, productCode, productName, page, limit)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail get expenditure products", err, gin.H{
			"from":         from.Format("2006-01-02"),
			"to":           to.Format("2006-01-02"),
			"pabeanType":   pabeanType,
			"productGroup": productGroup,
			"noPabean":     noPabean,
			"productCode":  productCode,
			"productName":  productName,
			"page":         page,
			"limit":        limit,
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit)) // ceil division
	hasNext := page < totalPages
	hasPrev := page > 1

	apiresponse.OK(ctx, res, "ok", gin.H{
		"from":         from.Format("2006-01-02"),
		"to":           to.Format("2006-01-02"),
		"pabeanType":   pabeanType,
		"productGroup": productGroup,
		"noPabean":     noPabean,
		"productCode":  productCode,
		"productName":  productName,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"totalCount": totalCount,
			"totalPages": totalPages,
			"count":      len(res),
			"hasNext":    hasNext,
			"hasPrev":    hasPrev,
		},
		"timezone": from.Location().String(),
	})
}

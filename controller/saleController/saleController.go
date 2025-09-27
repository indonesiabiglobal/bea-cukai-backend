// controller/saleController/sale_controller.go
package saleController

import (
	"Dashboard-TRDP/helper/apiRequest"
	"Dashboard-TRDP/helper/apiresponse"
	"Dashboard-TRDP/model"
	"Dashboard-TRDP/service/saleService"

	"github.com/gin-gonic/gin"
)

type SaleController struct {
	SaleService *saleService.SaleService
}

func NewSaleController(svc *saleService.SaleService) *SaleController {
	return &SaleController{SaleService: svc}
}

// ===== endpoints =====

// GET /dashboard/sales/summary?from=YYYY-MM-DD&to=YYYY-MM-DD
func (c *SaleController) GetKPISummary(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}

	category := apiRequest.ParseString(ctx, "category", "")

	requestData := model.SaleRequestParam{
		From:     from,
		To:       to,
		Category: category,
	}
	res, err := c.SaleService.GetKPISummary(requestData)
	if err != nil {
		apiresponse.InternalServerError(ctx, "SALES_KPI_FAILED", "fail get sales summary", err, apiresponse.MetaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", apiresponse.MetaRange(from, to))
}

// GET /dashboard/sales/trend?from=YYYY-MM-DD&to=YYYY-MM-DD
func (c *SaleController) GetTrend(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}

	category := apiRequest.ParseString(ctx, "category", "")

	requestData := model.SaleRequestParam{
		From:     from,
		To:       to,
		Category: category,
	}
	res, err := c.SaleService.GetTrend(requestData)
	if err != nil {
		apiresponse.InternalServerError(ctx, "SALES_TREND_FAILED", "fail get sales trend", err, apiresponse.MetaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", apiresponse.MetaRange(from, to))
}

// GET /dashboard/sales/top-products?from=YYYY-MM-DD&to=YYYY-MM-DD&limit=10
func (c *SaleController) GetTopProducts(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}

	category := apiRequest.ParseString(ctx, "category", "")

	requestData := model.SaleRequestParam{
		From:     from,
		To:       to,
		Category: category,
	}
	res, err := c.SaleService.GetTopProducts(requestData)
	if err != nil {
		apiresponse.InternalServerError(ctx, "SALES_TOP_PRODUCTS_FAILED", "fail get top products", err, apiresponse.MetaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from":     from.Format("2006-01-02"),
		"to":       to.Format("2006-01-02"),
		"timezone": "Asia/Jakarta",
	})
}

// GET /dashboard/sales/by-category?from=YYYY-MM-DD&to=YYYY-MM-DD&limit=10
func (c *SaleController) GetByCategory(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	limit := apiRequest.ParseLimit(ctx, 10)

	category := apiRequest.ParseString(ctx, "category", "")

	requestData := model.SaleRequestParam{
		From:     from,
		To:       to,
		Category: category,
	}
	res, err := c.SaleService.GetByCategory(requestData, limit)
	if err != nil {
		apiresponse.InternalServerError(ctx, "SALES_BY_CATEGORY_FAILED", "fail get sales by category", err, apiresponse.MetaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"limit":    limit,
		"from":     from.Format("2006-01-02"),
		"to":       to.Format("2006-01-02"),
		"timezone": "Asia/Jakarta",
	})
}

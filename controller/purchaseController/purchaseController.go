package purchaseController

import (
	"Dashboard-TRDP/helper/apiRequest"
	"Dashboard-TRDP/helper/apiresponse"
	"Dashboard-TRDP/model"
	"Dashboard-TRDP/service/purchaseService"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PurchaseController struct {
	Svc *purchaseService.PurchaseService
}

func NewPurchaseController(svc *purchaseService.PurchaseService) *PurchaseController {
	return &PurchaseController{Svc: svc}
}

var ErrBadRange = &badRangeError{}

type badRangeError struct{}

func (*badRangeError) Error() string { return "invalid date range (use YYYY-MM-DD query: from, to)" }

/* ========= Handlers ========= */
// GET /dashboard/purchase/summary?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *PurchaseController) GetKPISummary(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	fmt.Println(from, to, err)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}

	category := apiRequest.ParseString(ctx, "category", "")
	vendor := apiRequest.ParseString(ctx, "vendor", "")

	requestData := model.PurchaseRequestParam{
		From:     from,
		To:       to,
		Category: category,
		Vendor:   vendor,
	}

	res, err := h.Svc.GetKPISummary(requestData)
	if err != nil {
		apiresponse.InternalServerError(ctx, "PURCHASE_FAILED", "fail get purchase summary", err, nil)
		return
	}
	apiresponse.OK(ctx, res, "purchase summary", nil)
}

// GET /dashboard/purchase/trend?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *PurchaseController) GetTrend(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}

	category := apiRequest.ParseString(ctx, "category", "")
	vendor := apiRequest.ParseString(ctx, "vendor", "")

	requestData := model.PurchaseRequestParam{
		From:     from,
		To:       to,
		Category: category,
		Vendor:   vendor,
	}
	res, err := h.Svc.GetTrend(requestData)
	if err != nil {
		apiresponse.InternalServerError(ctx, "PURCHASE_FAILED", "fail get purchase trend", err, nil)
		return
	}
	apiresponse.OK(ctx, res, "purchase trend", nil)
}

// GET /dashboard/purchase/top-vendors?from=YYYY-MM-DD&to=YYYY-MM-DD&limit=10
func (h *PurchaseController) GetTopVendors(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}

	category := apiRequest.ParseString(ctx, "category", "")
	vendor := apiRequest.ParseString(ctx, "vendor", "")

	requestData := model.PurchaseRequestParam{
		From:     from,
		To:       to,
		Category: category,
		Vendor:   vendor,
	}
	res, err := h.Svc.GetTopVendors(requestData)
	if err != nil {
		apiresponse.InternalServerError(ctx, "PURCHASE_FAILED", "fail get top vendors", err, nil)
		return
	}
	apiresponse.OK(ctx, res, "top vendors", nil)
}

// GET /dashboard/purchase/top-products?from=YYYY-MM-DD&to=YYYY-MM-DD&limit=10
func (h *PurchaseController) GetTopProducts(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}

	category := apiRequest.ParseString(ctx, "category", "")
	vendor := apiRequest.ParseString(ctx, "vendor", "")

	requestData := model.PurchaseRequestParam{
		From:     from,
		To:       to,
		Category: category,
		Vendor:   vendor,
	}
	res, err := h.Svc.GetTopProducts(requestData)
	if err != nil {
		apiresponse.InternalServerError(ctx, "PURCHASE_TOP_PRODUCTS_FAILED", "fail get top products", err, nil)
		return
	}
	apiresponse.OK(ctx, res, "top products", nil)
}

// GET /dashboard/purchase/by-category?from=YYYY-MM-DD&to=YYYY-MM-DD&limit=10
func (h *PurchaseController) GetByCategory(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	vendor := apiRequest.ParseString(ctx, "vendor", "")

	requestData := model.PurchaseRequestParam{
		From:     from,
		To:       to,
		Vendor:   vendor,
	}
	limit := apiRequest.ParseLimit(ctx, 10)
	res, err := h.Svc.GetByCategory(requestData, limit)
	if err != nil {
		apiresponse.InternalServerError(ctx, "PURCHASE_BY_CATEGORY_FAILED", "fail get purchase by category", err, apiresponse.MetaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "purchase by category", nil)
}

func (h *PurchaseController) GetVendors(c *gin.Context) {
	search := apiRequest.ParseString(c, "search", "")
	page := apiRequest.ParseInt(c, "page", 1)
	limit := apiRequest.ParseInt(c, "limit", 20)

	res, err := h.Svc.GetVendors(c.Request.Context(), search, page, limit)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "MP_CATEGORIES_FETCH_FAILED", "fail get categories", err, nil)
		return
	}

	apiresponse.OK(c, res.Items, "ok", apiresponse.PageMeta{
		Page:  page,
		Limit: limit,
		Total: res.Total,
	})
}
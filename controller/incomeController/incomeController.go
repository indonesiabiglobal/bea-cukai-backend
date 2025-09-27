package incomeController

import (
	"net/http"
	"strconv"

	"Dashboard-TRDP/helper/apiRequest"
	"Dashboard-TRDP/helper/apiresponse"
	"Dashboard-TRDP/service/incomeService"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type IncomeController struct {
	IncomeService *incomeService.IncomeService
}

func NewIncomeController(svc *incomeService.IncomeService) *IncomeController {
	return &IncomeController{IncomeService: svc}
}

// ==========================
// Template-compatible endpoints
// ==========================

// GET /incomes
func (p *IncomeController) GetAllIncomes(ctx *gin.Context) {
	// token user (optional)
	var userID uint
	if v, exists := ctx.Get("userData"); exists {
		if claims, ok := v.(jwt.MapClaims); ok {
			if idf, ok := claims["id"].(float64); ok {
				userID = uint(idf)
			}
		}
	}

	rows, err := p.IncomeService.GetAllIncomes(userID)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "INCOMES_FETCH_FAILED", "fail get incomes", err, nil)
		return
	}
	apiresponse.OK(ctx, rows, "ok", nil)
}

// GET /incomes/:id
func (c *IncomeController) GetIncomeByID(ctx *gin.Context) {
	paramIncomeID := ctx.Param("id")
	incomeID, err := strconv.Atoi(paramIncomeID)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_ID", "invalid income id", err, gin.H{"id": paramIncomeID})
		return
	}
	row, err := c.IncomeService.GetIncomeByID(uint(incomeID))
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "INCOME_FETCH_FAILED", "fail get income", err, gin.H{"id": incomeID})
		return
	}
	apiresponse.OK(ctx, row, "ok", gin.H{"id": incomeID})
}

// ==========================
// Dashboard endpoints
// ==========================

// GET /dashboard/income/summary?from=YYYY-MM-DD&to=YYYY-MM-DD

func (c *IncomeController) GetKPISummary(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
    if err != nil {
        apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
            "from": ctx.Query("from"),
            "to":   ctx.Query("to"),
        })
        return
    }

    res, err := c.IncomeService.GetKPISummary(from, to)
    if err != nil {
        apiresponse.Error(ctx, http.StatusInternalServerError, "KPI_FETCH_FAILED", "fail get summary", err, gin.H{
            "from":     from.Format("2006-01-02"),
            "to":       to.Format("2006-01-02"),
            "timezone": from.Location().String(),
        })
        return
    }

    apiresponse.OK(ctx, res, "ok", gin.H{
        "from":     from.Format("2006-01-02"),
        "to":       to.Format("2006-01-02"),
        "timezone": from.Location().String(),
    })
}

// GET /dashboard/income/trend?from=...&to=...
func (c *IncomeController) GetRevenueTrend(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"), "to": ctx.Query("to"),
		})
		return
	}
	res, err := c.IncomeService.GetRevenueTrend(from, to)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "TREND_FETCH_FAILED", "fail get trend", err, gin.H{
			"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"),
			"timezone": from.Location().String(),
		})
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"),
		"timezone": from.Location().String(),
	})
}

// GET /dashboard/income/top-units?from=...&to=...&limit=10
func (c *IncomeController) GetTopUnits(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"), "to": ctx.Query("to"),
		})
		return
	}
	limit := apiRequest.ParseLimit(ctx, 10)
	res, err := c.IncomeService.GetTopUnits(from, to, limit)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "TOP_UNITS_FETCH_FAILED", "fail get top units", err, gin.H{
			"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"), "limit": limit,
		})
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"), "limit": limit,
	})
}

// GET /dashboard/income/top-providers?from=...&to=...&limit=10
func (c *IncomeController) GetTopProviders(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"), "to": ctx.Query("to"),
		})
		return
	}
	limit := apiRequest.ParseLimit(ctx, 10)
	res, err := c.IncomeService.GetTopProviders(from, to, limit)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "TOP_PROVIDERS_FETCH_FAILED", "fail get top providers", err, gin.H{
			"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"), "limit": limit,
		})
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"), "limit": limit,
	})
}

// GET /dashboard/income/top-guarantors?from=...&to=...&limit=10
func (c *IncomeController) GetTopGuarantors(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"), "to": ctx.Query("to"),
		})
		return
	}
	limit := apiRequest.ParseLimit(ctx, 10)
	res, err := c.IncomeService.GetTopGuarantors(from, to, limit)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "TOP_GUARANTORS_FETCH_FAILED", "fail get top guarantors", err, gin.H{
			"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"), "limit": limit,
		})
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"), "limit": limit,
	})
}

// GET /dashboard/income/top-guarantor-groups?from=...&to=...&limit=10
func (c *IncomeController) GetTopGuarantorGroups(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"), "to": ctx.Query("to"),
		})
		return
	}
	limit := apiRequest.ParseLimit(ctx, 10)
	res, err := c.IncomeService.GetTopGuarantorGroups(from, to, limit)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "TOP_GUARANTOR_GROUPS_FETCH_FAILED", "fail get top guarantor groups", err, gin.H{
			"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"), "limit": limit,
		})
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"), "limit": limit,
	})
}

// GET /dashboard/income/by-layanan?from=...&to=...&limit=10
func (c *IncomeController) GetRevenueByService(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"), "to": ctx.Query("to"),
		})
		return
	}
	res, err := c.IncomeService.GetRevenueByService(from, to)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "REVENUE_LAYANAN_FETCH_FAILED", "fail get revenue by layanan", err, gin.H{
			"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"),
		})
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"),
	})
}

// GET /dashboard/income/mix-ipop?from=...&to=...
func (c *IncomeController) GetRevenueByIPOP(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"), "to": ctx.Query("to"),
		})
		return
	}
	res, err := c.IncomeService.GetRevenueByIPOP(from, to)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "MIX_IPOP_FETCH_FAILED", "fail get mix ipop", err, gin.H{
			"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"),
		})
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"),
	})
}

// GET /dashboard/income/by-dow?from=...&to=...
func (c *IncomeController) GetRevenueByDOW(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"), "to": ctx.Query("to"),
		})
		return
	}
	res, err := c.IncomeService.GetRevenueByDOW(from, to)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "REVENUE_DOW_FETCH_FAILED", "fail get revenue by dow", err, gin.H{
			"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"),
		})
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from": from.Format("2006-01-02"), "to": to.Format("2006-01-02"),
	})
}
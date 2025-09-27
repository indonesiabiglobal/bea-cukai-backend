package visitController

import (
	"Dashboard-TRDP/helper/apiresponse"
	"Dashboard-TRDP/service/visitService"

	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type VisitController struct {
	VisitService *visitService.VisitService
}

func NewVisitController(svc *visitService.VisitService) *VisitController {
	return &VisitController{VisitService: svc}
}

/* ===== Helpers (samakan dengan helper globalmu bila sudah ada) ===== */

func getRange(ctx *gin.Context) (time.Time, time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	const layout = "2006-01-02"

	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")
	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, http.ErrNoLocation
	}

	from, err := time.ParseInLocation(layout, fromStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.ParseInLocation(layout, toStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil
}

func parseLimit(ctx *gin.Context, def int) int {
	s := ctx.DefaultQuery("limit", strconv.Itoa(def))
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func metaRange(from, to time.Time) any {
	return gin.H{
		"from":     from.Format("2006-01-02"),
		"to":       to.Format("2006-01-02"),
		"timezone": "Asia/Jakarta",
	}
}

/* ===== Endpoints ===== */

// GET /dashboard/visits/summary?from=YYYY-MM-DD&to=YYYY-MM-DD
func (c *VisitController) GetKPISummary(ctx *gin.Context) {
	from, to, err := getRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	res, err := c.VisitService.GetKPISummary(from, to)
	if err != nil {
		apiresponse.InternalServerError(ctx, "VISITS_KPI_FAILED", "fail get visits summary", err, metaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", metaRange(from, to))
}

// GET /dashboard/visits/trend?from=YYYY-MM-DD&to=YYYY-MM-DD
func (c *VisitController) GetTrend(ctx *gin.Context) {
	from, to, err := getRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	res, err := c.VisitService.GetTrend(from, to)
	if err != nil {
		apiresponse.InternalServerError(ctx, "VISITS_TREND_FAILED", "fail get visits trend", err, metaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", metaRange(from, to))
}

// GET /dashboard/visits/top-services?from=YYYY-MM-DD&to=YYYY-MM-DD&limit=10
func (c *VisitController) GetTopServices(ctx *gin.Context) {
	from, to, err := getRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	res, err := c.VisitService.GetTopServices(from, to)
	if err != nil {
		apiresponse.InternalServerError(ctx, "VISITS_TOP_SERVICES_FAILED", "fail get top services", err, metaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from":     from.Format("2006-01-02"),
		"to":       to.Format("2006-01-02"),
		"timezone": "Asia/Jakarta",
	})
}

// GET /dashboard/visits/top-guarantors?from=YYYY-MM-DD&to=YYYY-MM-DD&limit=10
func (c *VisitController) GetTopGuarantors(ctx *gin.Context) {
	from, to, err := getRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	res, err := c.VisitService.GetTopGuarantors(from, to)
	if err != nil {
		apiresponse.InternalServerError(ctx, "VISITS_TOP_GUARANTORS_FAILED", "fail get top guarantors", err, metaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from":     from.Format("2006-01-02"),
		"to":       to.Format("2006-01-02"),
		"timezone": "Asia/Jakarta",
	})
}

// GET /dashboard/visits/by-dow?from=YYYY-MM-DD&to=YYYY-MM-DD
func (c *VisitController) GetByDOW(ctx *gin.Context) {
	from, to, err := getRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	res, err := c.VisitService.GetByDOW(from, to)
	if err != nil {
		apiresponse.InternalServerError(ctx, "VISITS_BY_DOW_FAILED", "fail get visits by day-of-week", err, metaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", metaRange(from, to))
}

// GET /dashboard/visits/by-region?from=YYYY-MM-DD&to=YYYY-MM-DD&limit=10
func (c *VisitController) GetByRegionKota(ctx *gin.Context) {
	from, to, err := getRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	limit := parseLimit(ctx, 10)
	res, err := c.VisitService.GetByRegionKota(from, to, limit)
	if err != nil {
		apiresponse.InternalServerError(ctx, "VISITS_BY_REGION_FAILED", "fail get visits by region", err, metaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", gin.H{
		"limit":    limit,
		"from":     from.Format("2006-01-02"),
		"to":       to.Format("2006-01-02"),
		"timezone": "Asia/Jakarta",
	})
}

// GET /dashboard/visits/mix-ipop?from=YYYY-MM-DD&to=YYYY-MM-DD
func (c *VisitController) GetMixIPOP(ctx *gin.Context) {
	from, to, err := getRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	res, err := c.VisitService.GetMixIPOP(from, to)
	if err != nil {
		apiresponse.InternalServerError(ctx, "VISITS_MIX_IPOP_FAILED", "fail get visits mix ipop", err, metaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", metaRange(from, to))
}

// GET /dashboard/visits/los-buckets?from=YYYY-MM-DD&to=YYYY-MM-DD
func (c *VisitController) GetLOSBuckets(ctx *gin.Context) {
	from, to, err := getRange(ctx)
	if err != nil {
		apiresponse.BadRequest(ctx, "INVALID_DATE_RANGE", "invalid date range", err, nil)
		return
	}
	res, err := c.VisitService.GetLOSBuckets(from, to)
	if err != nil {
		apiresponse.InternalServerError(ctx, "VISITS_LOS_BUCKETS_FAILED", "fail get LOS buckets", err, metaRange(from, to))
		return
	}
	apiresponse.OK(ctx, res, "ok", metaRange(from, to))
}

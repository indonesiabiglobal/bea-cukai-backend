package productController

import (
	"net/http"
	"strconv"
	"strings"

	"Dashboard-TRDP/helper/apiresponse"
	"Dashboard-TRDP/repo/productRepository"
	"Dashboard-TRDP/service/productService"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	svc *productService.ProductService
}

func NewProductController(svc *productService.ProductService) *ProductController {
	return &ProductController{svc: svc}
}

/* ===== Helpers ===== */

func qStr(c *gin.Context, key, def string) string {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		return def
	}
	return v
}

func qInt(c *gin.Context, key string, def int) int {
	if v := strings.TrimSpace(c.Query(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func qBoolPtr(c *gin.Context, key string) *bool {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		return nil
	}
	if b, err := strconv.ParseBool(v); err == nil {
		return &b
	}
	return nil
}

/* ===== Routes =====
GET /master-products/categories?search=&page=1&limit=20
GET /master-products/products?search=&category_code=&status=&kode_akun=&kode_principle=&generik=&keras=&narkotik=&psikotropika=&page=1&limit=20
*/

func (h *ProductController) GetCategories(c *gin.Context) {
	search := qStr(c, "search", "")
	page := qInt(c, "page", 1)
	limit := qInt(c, "limit", 20)

	res, err := h.svc.GetCategories(c.Request.Context(), search, page, limit)
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

func (h *ProductController) GetProducts(c *gin.Context) {
	f := productRepository.ProductFilter{
		Search:        qStr(c, "search", ""),
		CategoryCode:  qStr(c, "category_code", ""),
		Status:        qStr(c, "status", ""),
		KodeAkun:      qStr(c, "kode_akun", ""),
		KodePrinciple: qStr(c, "kode_principle", ""),
		ObatGenerik:   qBoolPtr(c, "generik"),
		ObatKeras:     qBoolPtr(c, "keras"),
		Narkotik:      qBoolPtr(c, "narkotik"),
		Psikotropika:  qBoolPtr(c, "psikotropika"),
	}
	page := qInt(c, "page", 1)
	limit := qInt(c, "limit", 20)

	res, err := h.svc.GetProducts(c.Request.Context(), f, page, limit)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "MP_PRODUCTS_FETCH_FAILED", "fail get products", err, nil)
		return
	}

	apiresponse.OK(c, res.Items, "ok", apiresponse.PageMeta{
		Page:  page,
		Limit: limit,
		Total: res.Total,
	})
}

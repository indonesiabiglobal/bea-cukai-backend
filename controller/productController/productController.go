package productController

import (
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/model"
	"Bea-Cukai/service/productService"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productService *productService.ProductService
}

func NewProductController(productService *productService.ProductService) *ProductController {
	return &ProductController{
		productService: productService,
	}
}

func (p *ProductController) GetAll(ctx *gin.Context) {
	var req model.ProductRequest

	// Parse query parameters
	req.ItemCode = ctx.Query("item_code")
	req.ItemName = ctx.Query("item_name")
	req.ItemGroup = ctx.Query("item_group")
	req.ItemTypeCode = ctx.Query("item_type_code")

	// Parse pagination parameters
	req.Page = apiRequest.ParseInt(ctx, "page", 1)
	req.Limit = apiRequest.ParseLimit(ctx, 10)

	// Get data from service
	products, total, meta, err := p.productService.GetAll(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get products",
			"error":   err.Error(),
		})
		return
	}

	// Return response
	ctx.JSON(http.StatusOK, gin.H{
		"message":  "Products retrieved successfully",
		"data":     products,
		"meta":     meta,
		"total":    total,
	})
}

func (p *ProductController) GetByCode(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Item code is required",
		})
		return
	}

	product, err := p.productService.GetByCode(code)
	if err != nil {
		if err.Error() == "product with code "+code+" not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get product",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Product retrieved successfully",
		"data":    product,
	})
}
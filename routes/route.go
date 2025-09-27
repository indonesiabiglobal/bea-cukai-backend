package routes

import (
	"Dashboard-TRDP/controller/incomeController"
	"Dashboard-TRDP/controller/patientController"
	"Dashboard-TRDP/controller/productController"
	"Dashboard-TRDP/controller/purchaseController"
	"Dashboard-TRDP/controller/saleController"
	"Dashboard-TRDP/controller/userController"
	"Dashboard-TRDP/controller/visitController"
	"Dashboard-TRDP/middleware"
	"Dashboard-TRDP/repo/incomeRepository"
	"Dashboard-TRDP/repo/patientRepository"
	"Dashboard-TRDP/repo/productRepository"
	"Dashboard-TRDP/repo/purchaseRepository"
	"Dashboard-TRDP/repo/saleRepository"
	"Dashboard-TRDP/repo/userRepository"
	"Dashboard-TRDP/repo/visitRepository"
	"Dashboard-TRDP/service/incomeService"
	"Dashboard-TRDP/service/patientService"
	"Dashboard-TRDP/service/productService"
	"Dashboard-TRDP/service/purchaseService"
	"Dashboard-TRDP/service/saleService"
	"Dashboard-TRDP/service/userService"
	"Dashboard-TRDP/service/visitService"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func corsConfig() cors.Config {
	allow := os.Getenv("CORS_ALLOW_ORIGINS")
	origins := []string{"http://202.157.187.183:3131", "http://localhost:3000", "http://127.0.0.1:3000"}
	if allow != "" {
		parts := strings.Split(allow, ",")
		origins = make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				origins = append(origins, p)
			}
		}
	}

	return cors.Config{
		AllowOrigins: origins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"Accept-Language", "X-Requested-With", "Cache-Control", "Pragma",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // kalau pakai cookies; kalau tidak, bisa false
		MaxAge:           12 * time.Hour,
	}
}

func NewRoute(db *gorm.DB) *gin.Engine {
	// Repositories
	userRepository := userRepository.NewUserRepository(db)
	incomeRepository := incomeRepository.NewIncomeRepository(db)
	purchaseRepository := purchaseRepository.NewPurchaseRepository(db)
	saleRepository := saleRepository.NewSaleRepository(db)

	// Services
	userService := userService.NewUserService(userRepository)
	incomeService := incomeService.NewIncomeService(incomeRepository)
	purchaseService := purchaseService.NewPurchaseService(purchaseRepository)
	saleService := saleService.NewSaleService(saleRepository)

	// Controllers
	userController := userController.NewUserController(userService)
	incomeController := incomeController.NewIncomeController(incomeService)
	purchaseController := purchaseController.NewPurchaseController(purchaseService)
	saleController := saleController.NewSaleController(saleService)

	visitRepo := visitRepository.NewVisitRepository(db)
	visitSvc := visitService.NewVisitService(visitRepo)
	visitCtrl := visitController.NewVisitController(visitSvc)

	app := gin.Default()

	// CORS (dev)

	// 1) CORS paling pertama
	app.Use(cors.New(corsConfig()))

	// 2) Global preflight OK (aman walau cors middleware sudah handle)
	app.OPTIONS("/*any", func(c *gin.Context) { c.Status(204) })

	/* API Routes */

	// User Routes
	users := app.Group("/users")
	{
		users.POST("/register", userController.CreateUser)
		users.POST("/login", userController.LoginUser)

		// Protected user endpoints
		users.Use(middleware.Authentication())
		{
			users.PUT("/", userController.UpdateUser)
			users.DELETE("/", userController.DeleteUser)
		}
	}

	// Income Routes (raw list/detail)
	income := app.Group("/income")
	{
		income.GET("/", incomeController.GetAllIncomes)
		income.GET("/:id", incomeController.GetIncomeByID)
	}

	// Dashboard: Income analytics
	dashboardIncome := app.Group("/dashboard/income")
	{
		dashboardIncome.GET("/summary", incomeController.GetKPISummary)
		dashboardIncome.GET("/trend", incomeController.GetRevenueTrend)
		dashboardIncome.GET("/top-units", incomeController.GetTopUnits)
		dashboardIncome.GET("/top-providers", incomeController.GetTopProviders)
		dashboardIncome.GET("/top-guarantors", incomeController.GetTopGuarantors)
		dashboardIncome.GET("/top-guarantor-groups", incomeController.GetTopGuarantorGroups)
		dashboardIncome.GET("/revenue-by-service", incomeController.GetRevenueByService)
		dashboardIncome.GET("/mix-ipop", incomeController.GetRevenueByIPOP)
		dashboardIncome.GET("/by-dow", incomeController.GetRevenueByDOW)
	}

	dashboardPurchase := app.Group("/dashboard/purchase")
	{
		dashboardPurchase.GET("/summary", purchaseController.GetKPISummary)
		dashboardPurchase.GET("/trend", purchaseController.GetTrend)
		dashboardPurchase.GET("/top-vendors", purchaseController.GetTopVendors)
		dashboardPurchase.GET("/top-products", purchaseController.GetTopProducts)
		dashboardPurchase.GET("/by-category", purchaseController.GetByCategory)
	}

	purchase := app.Group("/purchases")
	{
		purchase.GET("/vendors", purchaseController.GetVendors)
	}

	dashboardSales := app.Group("/dashboard/sales")
	{
		dashboardSales.GET("/summary", saleController.GetKPISummary)
		dashboardSales.GET("/trend", saleController.GetTrend)
		dashboardSales.GET("/top-products", saleController.GetTopProducts)
		dashboardSales.GET("/by-category", saleController.GetByCategory)
	}

	visits := app.Group("/dashboard/visits")
	{
		visits.GET("/summary", visitCtrl.GetKPISummary)
		visits.GET("/trend", visitCtrl.GetTrend)
		visits.GET("/top-services", visitCtrl.GetTopServices)
		visits.GET("/top-guarantors", visitCtrl.GetTopGuarantors)
		visits.GET("/by-dow", visitCtrl.GetByDOW)
		visits.GET("/by-region", visitCtrl.GetByRegionKota)
		visits.GET("/mix-ipop", visitCtrl.GetMixIPOP)
		visits.GET("/los-buckets", visitCtrl.GetLOSBuckets)
	}

	// Product Routes
	productRepository := productRepository.NewProductRepository(db)
	productService := productService.NewProductService(productRepository)
	productController := productController.NewProductController(productService)

	mp := app.Group("/products")
	{
		mp.GET("/", productController.GetProducts)
		mp.GET("/categories", productController.GetCategories)
	}

	// Patient
	patientRepo := patientRepository.NewPatientRepository(db)
	patientSvc := patientService.NewPatientService(patientRepo)
	patientCtrl := patientController.NewPatientController(patientSvc)

	patient := app.Group("/patients")
	{
		patient.GET("/inpatient/monitoring", patientCtrl.MonitoringInpatient)
		patient.GET("/inpatient/discharged", patientCtrl.DischargedPatient)
	}

	return app
}

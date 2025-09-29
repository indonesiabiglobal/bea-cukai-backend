package routes

import (
	"Bea-Cukai/controller/auxiliaryMaterialReportController"
	"Bea-Cukai/controller/entryProductController"
	"Bea-Cukai/controller/expenditureProductController"
	"Bea-Cukai/controller/finishedProductReportController"
	"Bea-Cukai/controller/itemGroupController"
	"Bea-Cukai/controller/machineToolReportController"
	"Bea-Cukai/controller/pabeanController"
	"Bea-Cukai/controller/productController"
	"Bea-Cukai/controller/rawMaterialReportController"
	"Bea-Cukai/controller/rejectScrapReportController"
	"Bea-Cukai/controller/userController"
	"Bea-Cukai/controller/wipPositionReportController"
	"Bea-Cukai/middleware"
	"Bea-Cukai/repo/auxiliaryMaterialReportRepository"
	"Bea-Cukai/repo/entryProductRepository"
	"Bea-Cukai/repo/expenditureProductRepository"
	"Bea-Cukai/repo/finishedProductReportRepository"
	"Bea-Cukai/repo/itemGroupRepository"
	"Bea-Cukai/repo/machineToolReportRepository"
	"Bea-Cukai/repo/pabeanRepository"
	"Bea-Cukai/repo/productRepository"
	"Bea-Cukai/repo/rawMaterialReportRepository"
	"Bea-Cukai/repo/rejectScrapReportRepository"
	"Bea-Cukai/repo/userRepository"
	"Bea-Cukai/repo/wipPositionReportRepository"
	"Bea-Cukai/service/auxiliaryMaterialReportService"
	"Bea-Cukai/service/entryProductService"
	"Bea-Cukai/service/expenditureProductService"
	"Bea-Cukai/service/finishedProductReportService"
	"Bea-Cukai/service/itemGroupService"
	"Bea-Cukai/service/machineToolReportService"
	"Bea-Cukai/service/pabeanService"
	"Bea-Cukai/service/productService"
	"Bea-Cukai/service/rawMaterialReportService"
	"Bea-Cukai/service/rejectScrapReportService"
	"Bea-Cukai/service/userService"
	"Bea-Cukai/service/wipPositionReportService"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func corsConfig() cors.Config {
	allow := os.Getenv("CORS_ALLOW_ORIGINS")
	origins := []string{"http://192.168.100.100:3131", "http://localhost:3000", "http://127.0.0.1:3000","http://localhost:3131", "http://127.0.0.1:3131"}
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
		// AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

func NewRoute(db *gorm.DB) *gin.Engine {
	// Repositories
	userRepository := userRepository.NewUserRepository(db)
	entryProductRepository := entryProductRepository.NewEntryProductRepository(db)
	expenditureProductRepository := expenditureProductRepository.NewExpenditureProductRepository(db)
	pabeanRepository := pabeanRepository.NewPabeanRepository(db)
	itemGroupRepository := itemGroupRepository.NewItemGroupRepository(db)
	productRepository := productRepository.NewProductRepository(db)
	wipPositionReportRepository := wipPositionReportRepository.NewWipPositionReportRepository(db)
	rawMaterialReportRepository := rawMaterialReportRepository.NewRawMaterialReportRepository(db)
	finishedProductReportRepository := finishedProductReportRepository.NewFinishedProductReportRepository(db)
	machineToolReportRepository := machineToolReportRepository.NewMachineToolReportRepository(db)
	rejectScrapReportRepository := rejectScrapReportRepository.NewRejectScrapReportRepository(db)
	auxiliaryMaterialReportRepository := auxiliaryMaterialReportRepository.NewAuxiliaryMaterialReportRepository(db)

	// Services
	userService := userService.NewUserService(userRepository)
	entryProductService := entryProductService.NewEntryProductService(entryProductRepository)
	expenditureProductService := expenditureProductService.NewExpenditureProductService(expenditureProductRepository)
	pabeanService := pabeanService.NewPabeanService(pabeanRepository)
	itemGroupService := itemGroupService.NewItemGroupService(itemGroupRepository)
	productService := productService.NewProductService(productRepository)
	wipPositionReportService := wipPositionReportService.NewWipPositionReportService(wipPositionReportRepository)
	rawMaterialReportService := rawMaterialReportService.NewRawMaterialReportService(rawMaterialReportRepository)
	finishedProductReportService := finishedProductReportService.NewFinishedProductReportService(finishedProductReportRepository)
	machineToolReportService := machineToolReportService.NewMachineToolReportService(machineToolReportRepository)
	rejectScrapReportService := rejectScrapReportService.NewRejectScrapReportService(rejectScrapReportRepository)
	auxiliaryMaterialReportService := auxiliaryMaterialReportService.NewAuxiliaryMaterialReportService(auxiliaryMaterialReportRepository)

	// Controllers
	userController := userController.NewUserController(userService)
	entryProductController := entryProductController.NewEntryProductController(entryProductService)
	expenditureProductController := expenditureProductController.NewExpenditureProductController(expenditureProductService)
	pabeanController := pabeanController.NewPabeanController(pabeanService)
	itemGroupController := itemGroupController.NewItemGroupController(itemGroupService)
	productController := productController.NewProductController(productService)
	wipPositionReportController := wipPositionReportController.NewWipPositionReportController(wipPositionReportService)
	rawMaterialReportController := rawMaterialReportController.NewRawMaterialReportController(rawMaterialReportService)
	finishedProductReportController := finishedProductReportController.NewFinishedProductReportController(finishedProductReportService)
	machineToolReportController := machineToolReportController.NewMachineToolReportController(machineToolReportService)
	rejectScrapReportController := rejectScrapReportController.NewRejectScrapReportController(rejectScrapReportService)
	auxiliaryMaterialReportController := auxiliaryMaterialReportController.NewAuxiliaryMaterialReportController(auxiliaryMaterialReportService)

	app := gin.Default()

	// CORS (dev)

	// 1) CORS paling pertama
	app.Use(cors.New(corsConfig()))

	// 2) Global preflight OK (aman walau cors middleware sudah handle)
	app.OPTIONS("/*any", func(c *gin.Context) { c.Status(204) })

	/* API Routes */
	// auth Routes
	auth := app.Group("/auth")
	{
		auth.POST("/login", userController.LoginUser)
		auth.POST("/register", userController.CreateUser)

		// logout
		auth.Use(middleware.Authentication())
		{
			auth.POST("/logout", userController.LogoutUser)
		}
	}

	// User Routes
	users := app.Group("/users")
	{
		// Protected user endpoints
		users.Use(middleware.Authentication())
		{
			users.GET("", userController.GetAll)
			users.GET("/profile", userController.GetProfile)
			users.PUT("/", userController.UpdateUser)
			users.DELETE("/", userController.DeleteUser)
		}
	}

	// Report: EntryProduct analytics
	reportEntryProduct := app.Group("/report/entry-products")
	{
		reportEntryProduct.GET("", entryProductController.GetReport)
	}

	// Report: ExpenditureProduct analytics
	reportExpenditureProduct := app.Group("/report/expenditure-products")
	{
		reportExpenditureProduct.GET("", expenditureProductController.GetReport)
	}

	// Report: WIP Position
	reportWipPosition := app.Group("/report/wip-position")
	{
		reportWipPosition.GET("", wipPositionReportController.GetReport)
	}

	// Report: Raw Material
	reportRawMaterial := app.Group("/report/raw-material")
	{
		reportRawMaterial.GET("", rawMaterialReportController.GetReport)
	}

	// Report: Finished Product
	reportFinishedProduct := app.Group("/report/finished-product")
	{
		reportFinishedProduct.GET("", finishedProductReportController.GetReport)
	}

	// Report: Machine and Tool
	reportMachineTool := app.Group("/report/machine-tool")
	{
		reportMachineTool.GET("", machineToolReportController.GetReport)
	}

	// Report: Reject and Scrap
	reportRejectScrap := app.Group("/report/reject-scrap-product")
	{
		reportRejectScrap.GET("", rejectScrapReportController.GetReport)
	}

	// Report: Auxiliary Material
	reportAuxiliaryMaterial := app.Group("/auxiliary-material")
	{
		reportAuxiliaryMaterial.GET("", auxiliaryMaterialReportController.GetReport)
	}

	// Pabean: Master pabean document
	pabean := app.Group("/pabean")
	{
		pabean.GET("", pabeanController.GetAll)
	}

	// Item Groups: System item groups
	itemGroups := app.Group("/item-groups")
	{
		itemGroups.GET("", itemGroupController.GetAll)
	}

	// Products: Master products (ms_item)
	products := app.Group("/products")
	{
		products.GET("", productController.GetAll)
		products.GET("/:code", productController.GetByCode)
	}

	return app
}

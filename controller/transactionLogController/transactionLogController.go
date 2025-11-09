package transactionLogController

import (
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/model"
	"Bea-Cukai/service/transactionLogService"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type TransactionLogController struct {
	service *transactionLogService.TransactionLogService
}

func NewTransactionLogController(service *transactionLogService.TransactionLogService) *TransactionLogController {
	return &TransactionLogController{
		service: service,
	}
}

// GetAll - get all transaction logs with filtering
// @Summary Get all transaction logs
// @Tags TransactionLog
// @Produce json
// @Param start_date query string false "Start Date (YYYY-MM-DD)"
// @Param end_date query string false "End Date (YYYY-MM-DD)"
// @Param user_name query string false "Filter by username"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} apiresponse.Response
// @Router /transaction-logs [get]
func (c *TransactionLogController) GetAll(ctx *gin.Context) {
	var req model.TransactionLogRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		apiresponse.BadRequest(ctx, "INVALID_PARAMS", "Invalid request parameters", err, nil)
		return
	}

	result, err := c.service.GetAll(req)
	if err != nil {
		apiresponse.InternalServerError(ctx, "FETCH_FAILED", "Failed to get transaction logs", err, nil)
		return
	}

	apiresponse.OK(ctx, result, "Transaction logs retrieved successfully", nil)
}

// ExportExcel - export transaction logs to Excel
// @Summary Export transaction logs to Excel
// @Tags TransactionLog
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param start_date query string false "Start Date (YYYY-MM-DD)"
// @Param end_date query string false "End Date (YYYY-MM-DD)"
// @Param user_name query string false "Filter by username"
// @Success 200 {file} file
// @Router /transaction-logs/export [get]
func (c *TransactionLogController) ExportExcel(ctx *gin.Context) {
	var req model.TransactionLogRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		apiresponse.BadRequest(ctx, "INVALID_PARAMS", "Invalid request parameters", err, nil)
		return
	}

	// Get all data without pagination for export
	req.Page = 0
	req.Limit = 0
	result, err := c.service.GetAll(req)
	if err != nil {
		apiresponse.InternalServerError(ctx, "EXPORT_FAILED", "Failed to export transaction logs", err, nil)
		return
	}

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Transaction Log"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		apiresponse.InternalServerError(ctx, "EXCEL_ERROR", "Failed to create Excel sheet", err, nil)
		return
	}

	// Set headers
	headers := []string{"Tanggal", "User", "Module", "Kode Item/Transaksi", "Log"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
	}

	// Style for header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
		},
	})
	f.SetCellStyle(sheetName, "A1", "E1", headerStyle)

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 20) // Tanggal
	f.SetColWidth(sheetName, "B", "B", 15) // User
	f.SetColWidth(sheetName, "C", "C", 25) // Module
	f.SetColWidth(sheetName, "D", "D", 25) // Kode
	f.SetColWidth(sheetName, "E", "E", 50) // Log

	// Fill data
	for i, log := range result.Data {
		row := i + 2
		transDate := ""
		if log.TransDate.IsZero() {
			transDate = "-"
		} else {
			transDate = log.TransDate.Format("02/01/2006 15:04:05")
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), transDate)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), log.UserName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), log.Module)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), log.ActionCode)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), log.ActivityLog)
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Delete default sheet if exists
	sheetIndex, _ := f.GetSheetIndex("Sheet1")
	if sheetIndex != -1 {
		f.DeleteSheet("Sheet1")
	}

	// Generate filename
	filename := fmt.Sprintf("transaction-log-%s.xlsx", time.Now().Format("20060102-150405"))

	// Set headers for download
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", filename))
	ctx.Header("Content-Transfer-Encoding", "binary")

	// Write to response
	if err := f.Write(ctx.Writer); err != nil {
		apiresponse.InternalServerError(ctx, "WRITE_ERROR", "Failed to write Excel file", err, nil)
		return
	}
}

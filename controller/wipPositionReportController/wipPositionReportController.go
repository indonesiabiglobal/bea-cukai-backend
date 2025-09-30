package wipPositionReportController

import (
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/model"
	"Bea-Cukai/repo/wipPositionReportRepository"
	"Bea-Cukai/service/wipPositionReportService"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type WipPositionReportController struct {
	WipPositionReportService *wipPositionReportService.WipPositionReportService
}

func NewWipPositionReportController(svc *wipPositionReportService.WipPositionReportService) *WipPositionReportController {
	return &WipPositionReportController{WipPositionReportService: svc}
}

// ==========================
// Report endpoints
// ==========================

// GET /report/wip-position?from=YYYY-MM-DD&to=YYYY-MM-DD&item_code=...&item_name=...&page=1&rows=10
// Note: Using 'rows' parameter to match the PHP API convention
func (c *WipPositionReportController) GetReport(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"),
			"to":   ctx.Query("to"),
		})
		return
	}

	// Get optional filter parameters
	itemCode := ctx.Query("item_code")
	itemName := ctx.Query("item_name")

	// Get pagination parameters (using 'limit' to match PHP API)
	page := apiRequest.ParseInt(ctx, "page", 0)
	limit := apiRequest.ParseInt(ctx, "limit", 0) // Using 'limit' instead of 'rows' to match PHP

	filter := wipPositionReportRepository.GetReportFilter{
		TglAwal:  from,
		TglAkhir: to,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     page,
		Limit:    limit,
	}
	res, totalCount, err := c.WipPositionReportService.GetReport(filter)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail get WIP position report", err, gin.H{
			"from":      from.Format("2006-01-02"),
			"to":        to.Format("2006-01-02"),
			"item_code": itemCode,
			"item_name": itemName,
			"page":      page,
			"rows":      limit,
		})
		return
	}

	// Calculate pagination metadata
	var totalPages int
	var hasNext, hasPrev bool
	
	if limit > 0 {
		totalPages = int((totalCount + int64(limit) - 1) / int64(limit)) // ceil division
		hasNext = page < totalPages
		hasPrev = page > 1
	} else {
		// No pagination case
		totalPages = 1
		hasNext = false
		hasPrev = false
	}

	// Response format matching PHP API structure
	apiresponse.OK(ctx, res, "ok", gin.H{
		"from":      from.Format("2006-01-02"),
		"to":        to.Format("2006-01-02"),
		"item_code": itemCode,
		"item_name": itemName,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"totalCount": totalCount,
			"totalPages": totalPages,
			"count":      len(res),
			"hasNext":    hasNext,
			"hasPrev":    hasPrev,
		},
	})
}

// GET /report/wip-position/export?from=YYYY-MM-DD&to=YYYY-MM-DD&item_code=...&item_name=...
func (c *WipPositionReportController) ExportExcel(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"),
			"to":   ctx.Query("to"),
		})
		return
	}

	// Get optional filter parameters
	itemCode := ctx.Query("item_code")
	itemName := ctx.Query("item_name")

	// For export, we don't use pagination - get all data
	filter := wipPositionReportRepository.GetReportFilter{
		TglAwal:  from,
		TglAkhir: to,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     0, // No pagination
		Limit:    0, // No limit
	}

	res, _, err := c.WipPositionReportService.GetReport(filter)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail get WIP position report for export", err, gin.H{
			"from":      from.Format("2006-01-02"),
			"to":        to.Format("2006-01-02"),
			"item_code": itemCode,
			"item_name": itemName,
		})
		return
	}

	// Generate Excel file
	excelFile, err := c.generateExcelFile(res, from)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "EXCEL_GENERATION_FAILED", "failed to generate Excel file", err, gin.H{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		})
		return
	}
	defer excelFile.Close()

	// Set headers for Excel file download
	filename := fmt.Sprintf("laporan_posisi_wip_%s.xlsx", 
		from.Format("2006-01-02"))
	
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.Header("Cache-Control", "no-cache")
	
	// Write Excel file to response
	buffer, err := excelFile.WriteToBuffer()
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "EXCEL_WRITE_FAILED", "failed to write Excel file", err, gin.H{})
		return
	}
	
	ctx.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())
}

// generateExcelFile creates a real XLSX file using excelize for WIP Position Report
func (c *WipPositionReportController) generateExcelFile(data []model.WipPositionReportResponse, reportDate time.Time) (*excelize.File, error) {
	// Create a new Excel file
	f := excelize.NewFile()
	sheetName := "Laporan Posisi WIP"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)

	// Company and report header (no border table)
	f.SetCellValue(sheetName, "A1", "PT FUKUSUKE KOGYO INDONESIA")
	f.SetCellValue(sheetName, "A2", "LAPORAN PERTANGGUNGJAWABAN POSISI WIP")
	f.SetCellValue(sheetName, "A3", "") // Empty row
	f.SetCellValue(sheetName, "A4", "Nama Kawasan Berikat")
	f.SetCellValue(sheetName, "C4", ": PT FUKUSUKE KOGYO INDONESIA")
	f.SetCellValue(sheetName, "A5", "NPWP")
	f.SetCellValue(sheetName, "C5", ": 01.071.250.3-052.000")
	f.SetCellValue(sheetName, "A6", "Alamat")
	f.SetCellValue(sheetName, "C6", ": Blok M-3-2, Kawasan MM2100, Cikarang Barat, Bekasi, 17520")
	f.SetCellValue(sheetName, "A7", "Periode Laporan")
	f.SetCellValue(sheetName, "C7", fmt.Sprintf(": %s", reportDate.Format("02-01-2006")))

	// Set header info style (left aligned, bold)
	headerInfoStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Font:      &excelize.Font{Bold: true, Size: 12},
	})
	f.SetCellStyle(sheetName, "A1", "A7", headerInfoStyle)

	// Merge cells for company info values
	f.MergeCell(sheetName, "C4", "E4")
	f.MergeCell(sheetName, "C5", "E5")
	f.MergeCell(sheetName, "C6", "E6")
	f.MergeCell(sheetName, "C7", "E7")

	// Set table headers starting from row 9
	headers := [][]string{
		{"No.", "KODE BARANG", "NAMA BARANG", "SAT", "SALDO AKHIR"},
		{"", "", "", "", reportDate.Format("2006-01-02")},
	}

	// Set first header row (row 9)
	for col, header := range headers[0] {
		cell, _ := excelize.CoordinatesToCellName(col+1, 9)
		f.SetCellValue(sheetName, cell, header)
	}

	// Set second header row (row 10)
	for col, header := range headers[1] {
		if header != "" {
			cell, _ := excelize.CoordinatesToCellName(col+1, 10)
			f.SetCellValue(sheetName, cell, header)
		}
	}

	// Merge header cells
	f.MergeCell(sheetName, "A9", "A10")  // No.
	f.MergeCell(sheetName, "B9", "B10")  // KODE BARANG
	f.MergeCell(sheetName, "C9", "C10")  // NAMA BARANG
	f.MergeCell(sheetName, "D9", "D10")  // SAT
	// E9 = "SALDO AKHIR", E10 = date

	// Set header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Font:      &excelize.Font{Bold: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F0F0F0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A9", "E10", headerStyle)

	// Add data rows starting from row 11
	for i, wipItem := range data {
		row := i + 11
		
		// Set values for each column  
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)                       // No
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), wipItem.ItemCode)          // Item Code (kode_barang)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), wipItem.ItemName)          // Item Name (nama_barang)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), wipItem.UnitCode)          // Unit Code (sat)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), wipItem.Jumlah)            // Jumlah (saldo akhir)
	}

	// Set data style with borders
	if len(data) > 0 {
		dataStyle, _ := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
		})
		lastRow := len(data) + 10
		f.SetCellStyle(sheetName, "A11", fmt.Sprintf("E%d", lastRow), dataStyle)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 5)   // No
	f.SetColWidth(sheetName, "B", "B", 15)  // Item Code
	f.SetColWidth(sheetName, "C", "C", 40)  // Item Name
	f.SetColWidth(sheetName, "D", "D", 8)   // Unit Code
	f.SetColWidth(sheetName, "E", "E", 15)  // Saldo Akhir

	// Delete default sheet if it exists
	f.DeleteSheet("Sheet1")

	return f, nil
}

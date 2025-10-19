package finishedProductReportController

import (
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/model"
	"Bea-Cukai/repo/finishedProductReportRepository"
	"Bea-Cukai/service/finishedProductReportService"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type FinishedProductReportController struct {
	FinishedProductReportService *finishedProductReportService.FinishedProductReportService
}

func NewFinishedProductReportController(svc *finishedProductReportService.FinishedProductReportService) *FinishedProductReportController {
	return &FinishedProductReportController{FinishedProductReportService: svc}
}

// ==========================
// Report endpoints
// ==========================

// GET /report/finished-product?from=YYYY-MM-DD&to=YYYY-MM-DD&item_code=...&item_name=...&page=1&limit=10
func (c *FinishedProductReportController) GetReport(ctx *gin.Context) {
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

	// Get pagination parameters
	page := apiRequest.ParseInt(ctx, "page", 0)
	limit := apiRequest.ParseInt(ctx, "limit", 0)

	filter := finishedProductReportRepository.GetReportFilter{
		From:     from,
		To:       to,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     page,
		Limit:    limit,
	}
	res, totalCount, err := c.FinishedProductReportService.GetReport(filter)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail to get finished product report", err, gin.H{
			"from":      from.Format("2006-01-02"),
			"to":        to.Format("2006-01-02"),
			"item_code": itemCode,
			"item_name": itemName,
			"page":      page,
			"limit":     limit,
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

// ExportExcel generates Excel file for finished product report
func (c *FinishedProductReportController) ExportExcel(ctx *gin.Context) {
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

	filter := finishedProductReportRepository.GetReportFilter{
		From:     from,
		To:       to,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     0, // No pagination for export
		Limit:    0, // Get all data
	}

	res, _, err := c.FinishedProductReportService.GetReport(filter)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail to get finished product report", err, gin.H{
			"from":      from.Format("2006-01-02"),
			"to":        to.Format("2006-01-02"),
			"item_code": itemCode,
			"item_name": itemName,
		})
		return
	}

	// Generate Excel file
	filename, err := c.generateExcelFile(res, from, to)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "EXCEL_GENERATION_FAILED", "fail to generate excel file", err, nil)
		return
	}

	// Set headers for file download
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Expires", "0")
	ctx.Header("Cache-Control", "must-revalidate")
	ctx.Header("Pragma", "public")

	// Stream the file
	ctx.File(filename)
}

func (c *FinishedProductReportController) generateExcelFile(data []model.FinishedProductReportResponse, from, to time.Time) (string, error) {
	// Create a new workbook
	f := excelize.NewFile()

	// Create a new worksheet
	sheetName := "Laporan Barang Jadi"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return "", err
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Company header section
	f.SetCellValue(sheetName, "A1", "PT FUKUSUKE KOGYO INDONESIA")
	f.SetCellValue(sheetName, "A2", "LAPORAN PERTANGGUNGJAWABAN MUTASI BARANG JADI")
	f.SetCellValue(sheetName, "A3", "") // Empty row
	f.SetCellValue(sheetName, "A4", "Nama Kawasan Berikat")
	f.SetCellValue(sheetName, "C4", ": PT FUKUSUKE KOGYO INDONESIA")
	f.SetCellValue(sheetName, "A5", "NPWP")
	f.SetCellValue(sheetName, "C5", ": 01.071.250.3-052.000")
	f.SetCellValue(sheetName, "A6", "Alamat")
	f.SetCellValue(sheetName, "C6", ": Blok M-3-2, Kawasan MM2100, Cikarang Barat, Bekasi, 17520")
	f.SetCellValue(sheetName, "A7", "Periode Laporan")
	f.SetCellValue(sheetName, "C7", fmt.Sprintf(": %s s.d %s", from.Format("02-01-2006"), to.Format("02-01-2006")))
	
	// Merge cells for company info values (extend to column L for 12 columns)
	f.MergeCell(sheetName, "A1", "L1") // Company name
	f.MergeCell(sheetName, "A2", "L2") // Report title
	f.MergeCell(sheetName, "C4", "L4") // Kawasan Berikat
	f.MergeCell(sheetName, "C5", "L5") // NPWP
	f.MergeCell(sheetName, "C6", "L6") // Alamat
	f.MergeCell(sheetName, "C7", "L7") // Periode

	// Set header info style (center for titles, left for info)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Font:      &excelize.Font{Bold: true, Size: 12},
	})
	f.SetCellStyle(sheetName, "A1", "A2", titleStyle)

	headerInfoStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Font:      &excelize.Font{Bold: true, Size: 10},
	})
	f.SetCellStyle(sheetName, "A4", "A7", headerInfoStyle)

	// Table headers - first row (row 9)
	headers1 := []string{"No.", "KODE BARANG", "NAMA BARANG", "SAT", "SALDO AWAL", "PEMASUKAN", "PENGELUARAN", "PENYESUAIAN", "SALDO AKHIR", "STOK OPNAME", "SELISIH", "KETERANGAN"}
	for i, header := range headers1 {
		cell := fmt.Sprintf("%c9", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Merge cells for headers that span both rows
	f.MergeCell(sheetName, "A9", "A10") // No.
	f.MergeCell(sheetName, "B9", "B10") // KODE BARANG
	f.MergeCell(sheetName, "C9", "C10") // NAMA BARANG
	f.MergeCell(sheetName, "D9", "D10") // SAT
	f.MergeCell(sheetName, "E9", "E10") // SALDO AWAL
	f.MergeCell(sheetName, "F9", "F10") // PEMASUKAN
	f.MergeCell(sheetName, "G9", "G10") // PENGELUARAN
	f.MergeCell(sheetName, "H9", "H10") // PENYESUAIAN
	f.MergeCell(sheetName, "K9", "K10") // SELISIH
	f.MergeCell(sheetName, "L9", "L10") // KETERANGAN

	// Second row headers (row 10) - only for columns I and J
	f.SetCellValue(sheetName, "I10", to.Format("02-01-2006"))
	f.SetCellValue(sheetName, "J10", to.Format("02-01-2006"))

	// Fill data starting from row 11
	for i, item := range data {
		row := i + 11
		
		// Convert decimal.Decimal to float64 for proper number formatting in Excel
		awalFloat, _ := item.Awal.Float64()
		mskFloat, _ := item.Msk.Float64()
		keluarFloat, _ := item.Keluar.Float64()
		penyFloat, _ := item.Peny.Float64()
		akhrFloat, _ := item.Akhr.Float64()
		selisihFloat, _ := item.Selisih.Float64()

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.ItemCode)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.ItemName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), item.UnitCode)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), awalFloat)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), mskFloat)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), keluarFloat)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), penyFloat)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), akhrFloat)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), akhrFloat)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), selisihFloat)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), "")
	}

	// Apply styling
	// Header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Data style
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Number style for numeric columns (without decimal places)
	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 3, // "#,##0" - format ribuan tanpa desimal
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Apply header styles
	f.SetCellStyle(sheetName, "A9", "L10", headerStyle)

	// Apply data styles
	if len(data) > 0 {
		lastRow := len(data) + 10
		f.SetCellStyle(sheetName, "A11", fmt.Sprintf("A%d", lastRow), dataStyle)  // No
		f.SetCellStyle(sheetName, "B11", fmt.Sprintf("D%d", lastRow), dataStyle)  // Text columns
		f.SetCellStyle(sheetName, "E11", fmt.Sprintf("K%d", lastRow), numberStyle) // Number columns
		f.SetCellStyle(sheetName, "L11", fmt.Sprintf("L%d", lastRow), dataStyle)  // Keterangan
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 5)   // No
	f.SetColWidth(sheetName, "B", "B", 15)  // Kode Barang
	f.SetColWidth(sheetName, "C", "C", 25)  // Nama Barang
	f.SetColWidth(sheetName, "D", "D", 8)   // Satuan
	f.SetColWidth(sheetName, "E", "K", 12)  // Numeric columns
	f.SetColWidth(sheetName, "L", "L", 15)  // Keterangan

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("Finished_Product_Report_%s.xlsx", timestamp)

	// Save the file
	dir := "file/export"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}
	filePath := filepath.Join(dir, filename)
	if err = f.SaveAs(filePath); err != nil {
		return "", err
	}
	filename = filePath

	return filename, nil
}

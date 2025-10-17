package rawMaterialReportController

import (
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/model"
	"Bea-Cukai/repo/rawMaterialReportRepository"
	"Bea-Cukai/service/rawMaterialReportService"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type RawMaterialReportController struct {
	RawMaterialReportService *rawMaterialReportService.RawMaterialReportService
}

func NewRawMaterialReportController(svc *rawMaterialReportService.RawMaterialReportService) *RawMaterialReportController {
	return &RawMaterialReportController{RawMaterialReportService: svc}
}

// ==========================
// Report endpoints
// ==========================

// GET /report/raw-material?from=YYYY-MM-DD&to=YYYY-MM-DD&item_code=...&item_name=...&page=1&limit=10
func (c *RawMaterialReportController) GetReport(ctx *gin.Context) {
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

	filter := rawMaterialReportRepository.GetReportFilter{
		From:     from,
		To:       to,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     page,
		Limit:    limit,
	}
	res, totalCount, err := c.RawMaterialReportService.GetReport(filter)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail to get raw material report", err, gin.H{
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

// GET /report/raw-material/export?from=YYYY-MM-DD&to=YYYY-MM-DD&item_code=...&item_name=...
func (c *RawMaterialReportController) ExportExcel(ctx *gin.Context) {
	fromStr := ctx.Query("from")
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE", "invalid from date", err, gin.H{
			"from": fromStr,
		})
		return
	}
	toStr := ctx.Query("to")
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE", "invalid to date", err, gin.H{
			"to": toStr,
		})
		return
	}

	// Get optional filter parameters
	itemCode := ctx.Query("item_code")
	itemName := ctx.Query("item_name")

	// For export, we don't use pagination - get all data
	filter := rawMaterialReportRepository.GetReportFilter{
		From:     from,
		To:       to,
		ItemCode: itemCode,
		ItemName: itemName,
		Page:     0, // No pagination
		Limit:    0, // No limit
	}

	res, _, err := c.RawMaterialReportService.GetReport(filter)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail to get raw material report for export", err, gin.H{
			"from":      from,
			"to":        to,
			"item_code": itemCode,
			"item_name": itemName,
		})
		return
	}

	// Generate Excel file
	excelFile, err := c.generateExcelFile(res, from, to)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "EXCEL_GENERATION_FAILED", "failed to generate Excel file", err, gin.H{
			"from": from,
			"to":   to,
		})
		return
	}
	defer excelFile.Close()

	// Set headers for Excel file download
	filename := fmt.Sprintf("laporan_mutasi_bahan_baku_%s_%s.xlsx", from.Format("2006-01-02"), to.Format("2006-01-02"))
	
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

// generateExcelFile creates a real XLSX file using excelize for Raw Material Report
func (c *RawMaterialReportController) generateExcelFile(data []model.RawMaterialReportResponse, from, to time.Time) (*excelize.File, error) {
	// Create a new Excel file
	f := excelize.NewFile()
	sheetName := "Laporan Mutasi Bahan Baku"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)

	// Company and report header (no border table)
	f.SetCellValue(sheetName, "A1", "PT FUKUSUKE KOGYO INDONESIA")
	f.SetCellValue(sheetName, "A2", "LAPORAN PERTANGGUNGJAWABAN MUTASI BAHAN BAKU DAN BAHAN PENOLONG")
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

	// Set table headers starting from row 9
	headers := [][]string{
		{"No.", "KODE BARANG", "NAMA BARANG", "SAT", "SALDO AWAL", "PEMASUKAN", "PENGELUARAN", "PENYESUAIAN", "SALDO AKHIR", "STOK OPNAME", "SELISIH", "KETERANGAN"},
		{"", "", "", "", "", "", "", "", to.Format("2006-01-02"), to.Format("2006-01-02"), "", ""},
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

	// Merge header cells for most columns that don't have sub-headers
	f.MergeCell(sheetName, "A9", "A10")   // No.
	f.MergeCell(sheetName, "B9", "B10")   // KODE BARANG
	f.MergeCell(sheetName, "C9", "C10")   // NAMA BARANG
	f.MergeCell(sheetName, "D9", "D10")   // SAT
	f.MergeCell(sheetName, "E9", "E10")   // SALDO AWAL
	f.MergeCell(sheetName, "F9", "F10")   // PEMASUKAN
	f.MergeCell(sheetName, "G9", "G10")   // PENGELUARAN
	f.MergeCell(sheetName, "H9", "H10")   // PENYESUAIAN
	// I9 = "SALDO AKHIR", I10 = date
	// J9 = "STOK OPNAME", J10 = date
	f.MergeCell(sheetName, "K9", "K10")   // SELISIH
	f.MergeCell(sheetName, "L9", "L10")   // KETERANGAN

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
	f.SetCellStyle(sheetName, "A9", "L10", headerStyle)

	// Add data rows starting from row 11
	for i, rawMaterial := range data {
		row := i + 11
		
		// Set values for each column
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)                           // No
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), rawMaterial.ItemCode)          // Item Code
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), rawMaterial.ItemName)          // Item Name
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), rawMaterial.UnitCode)          // Unit Code
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), rawMaterial.Awal)              // Saldo Awal
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), rawMaterial.Masuk)             // Pemasukan
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), rawMaterial.Keluar)            // Pengeluaran
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), rawMaterial.Peny)              // Penyesuaian
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), rawMaterial.Akhir)             // Saldo Akhir
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), rawMaterial.Opname)            // Stok Opname
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), rawMaterial.Selisih)           // Selisih
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), "")                            // Keterangan (empty for now)
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
		f.SetCellStyle(sheetName, "A11", fmt.Sprintf("L%d", lastRow), dataStyle)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 5)   // No
	f.SetColWidth(sheetName, "B", "B", 15)  // Item Code
	f.SetColWidth(sheetName, "C", "C", 30)  // Item Name
	f.SetColWidth(sheetName, "D", "D", 8)   // Unit Code
	f.SetColWidth(sheetName, "E", "E", 12)  // Saldo Awal
	f.SetColWidth(sheetName, "F", "F", 12)  // Pemasukan
	f.SetColWidth(sheetName, "G", "G", 12)  // Pengeluaran
	f.SetColWidth(sheetName, "H", "H", 12)  // Penyesuaian
	f.SetColWidth(sheetName, "I", "I", 12)  // Saldo Akhir
	f.SetColWidth(sheetName, "J", "J", 12)  // Stok Opname
	f.SetColWidth(sheetName, "K", "K", 10)  // Selisih
	f.SetColWidth(sheetName, "L", "L", 15)  // Keterangan

	// Delete default sheet if it exists
	f.DeleteSheet("Sheet1")

	return f, nil
}

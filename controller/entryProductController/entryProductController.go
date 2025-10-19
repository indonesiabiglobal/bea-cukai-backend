package entryProductController

import (
	"Bea-Cukai/helper/apiRequest"
	"Bea-Cukai/helper/apiresponse"
	"Bea-Cukai/model"
	"Bea-Cukai/repo/entryProductRepository"
	"Bea-Cukai/service/entryProductService"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type EntryProductController struct {
	EntryProductService *entryProductService.EntryProductService
}

func NewEntryProductController(svc *entryProductService.EntryProductService) *EntryProductController {
	return &EntryProductController{EntryProductService: svc}
}

// ==========================
// Report endpoints
// ==========================

// GET /report/entryProduct/all?from=YYYY-MM-DD&to=YYYY-MM-DD&pabeanType=...&productGroup=...&noPabean=...&productCode=...&productName=...&page=1&limit=10
func (c *EntryProductController) GetReport(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"),
			"to":   ctx.Query("to"),
		})
		return
	}

	// Get optional filter parameters
	pabeanType := ctx.Query("pabeanType")
	productGroup := ctx.Query("productGroup")
	noPabean := ctx.Query("noPabean")
	productCode := ctx.Query("productCode")
	productName := ctx.Query("productName")

	// Get pagination parameters
	// Use 0 as default to indicate no pagination when not provided
	page := apiRequest.ParseInt(ctx, "page", 0)
	limit := apiRequest.ParseInt(ctx, "limit", 0)
	
	filter := entryProductRepository.GetReportFilter{
		From:         from,
		To:           to,
		PabeanType:   pabeanType,
		ProductGroup: productGroup,
		NoPabean:     noPabean,
		ProductCode:  productCode,
		ProductName:  productName,
		Page:         page,
		Limit:        limit,
		IsExport:     false,
	}

	res, totalCount, err := c.EntryProductService.GetReport(filter)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail get entry products", err, gin.H{
			"from":         from.Format("2006-01-02"),
			"to":           to.Format("2006-01-02"),
			"pabeanType":   pabeanType,
			"productGroup": productGroup,
			"noPabean":     noPabean,
			"productCode":  productCode,
			"productName":  productName,
			"page":         page,
			"limit":        limit,
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

	apiresponse.OK(ctx, res, "ok", gin.H{
		"from":         from.Format("2006-01-02"),
		"to":           to.Format("2006-01-02"),
		"pabeanType":   pabeanType,
		"productGroup": productGroup,
		"noPabean":     noPabean,
		"productCode":  productCode,
		"productName":  productName,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"totalCount": totalCount,
			"totalPages": totalPages,
			"count":      len(res),
			"hasNext":    hasNext,
			"hasPrev":    hasPrev,
		},
		"timezone": from.Location().String(),
	})
}

// GET /report/entryProduct/export?from=YYYY-MM-DD&to=YYYY-MM-DD&pabeanType=...&productGroup=...&noPabean=...&productCode=...&productName=...
func (c *EntryProductController) ExportExcel(ctx *gin.Context) {
	from, to, err := apiRequest.GetRange(ctx)
	if err != nil {
		apiresponse.Error(ctx, http.StatusBadRequest, "BAD_DATE_RANGE", "invalid date range", err, gin.H{
			"from": ctx.Query("from"),
			"to":   ctx.Query("to"),
		})
		return
	}

	// Get optional filter parameters
	pabeanType := ctx.Query("pabeanType")
	productGroup := ctx.Query("productGroup")
	noPabean := ctx.Query("noPabean")
	productCode := ctx.Query("productCode")
	productName := ctx.Query("productName")

	// For export, we don't use pagination - get all data
	filter := entryProductRepository.GetReportFilter{
		From:         from,
		To:           to,
		PabeanType:   pabeanType,
		ProductGroup: productGroup,
		NoPabean:     noPabean,
		ProductCode:  productCode,
		ProductName:  productName,
		Page:         0, // No pagination
		Limit:        0, // No limit
		IsExport:     true,
	}

	res, _, err := c.EntryProductService.GetReport(filter)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "DATA_FETCH_FAILED", "fail get entry products for export", err, gin.H{
			"from":         from.Format("2006-01-02"),
			"to":           to.Format("2006-01-02"),
			"pabeanType":   pabeanType,
			"productGroup": productGroup,
			"noPabean":     noPabean,
			"productCode":  productCode,
			"productName":  productName,
		})
		return
	}

	// Generate Excel file
	excelFile, err := c.generateExcelFile(res, from, to)
	if err != nil {
		apiresponse.Error(ctx, http.StatusInternalServerError, "EXCEL_GENERATION_FAILED", "failed to generate Excel file", err, gin.H{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		})
		return
	}
	defer excelFile.Close()

	// Set headers for Excel file download
	filename := fmt.Sprintf("laporan_pemasukan_barang_%s_%s.xlsx", 
		from.Format("2006-01-02"), 
		to.Format("2006-01-02"))
	
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

// generateExcelFile creates a real XLSX file using excelize
func (c *EntryProductController) generateExcelFile(data []model.EntryProduct, from, to time.Time) (*excelize.File, error) {
	// Create a new Excel file
	f := excelize.NewFile()
	sheetName := "Laporan Pemasukan Barang"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)

	// Set title and header information
	title1 := "LAPORAN PENERIMAAN BARANG PER DOKUMEN PABEAN"
	title2 := "PT FUKUSUKE KOGYO INDONESIA"
	title3 := fmt.Sprintf("PERIODE : %s S.D %s", from.Format("2006-01-02"), to.Format("2006-01-02"))
	
	f.SetCellValue(sheetName, "A1", title1)
	f.SetCellValue(sheetName, "A2", title2)
	f.SetCellValue(sheetName, "A3", title3)

	// Merge cells for titles
	f.MergeCell(sheetName, "A1", "M1")
	f.MergeCell(sheetName, "A2", "M2")
	f.MergeCell(sheetName, "A3", "M3")

	// Set title styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Font:      &excelize.Font{Bold: true, Size: 14},
	})
	f.SetCellStyle(sheetName, "A1", "M3", titleStyle)

	// Set table headers starting from row 5
	headers := [][]string{
		{"No.", "DOKUMEN PABEAN", "", "", "BUKTI PENERIMAAN BARANG", "", "PENGIRIM BARANG", "KODE BARANG", "NAMA BARANG", "JUMLAH", "SATUAN", "VALAS", "NILAI"},
		{"", "JENIS", "NOMOR", "TANGGAL", "NOMOR", "TANGGAL", "", "", "", "", "", "", ""},
	}

	// Set first header row (row 5)
	for col, header := range headers[0] {
		cell, _ := excelize.CoordinatesToCellName(col+1, 5)
		f.SetCellValue(sheetName, cell, header)
	}

	// Set second header row (row 6)
	for col, header := range headers[1] {
		if header != "" {
			cell, _ := excelize.CoordinatesToCellName(col+1, 6)
			f.SetCellValue(sheetName, cell, header)
		}
	}

	// Merge header cells
	f.MergeCell(sheetName, "A5", "A6") // No.
	f.MergeCell(sheetName, "B5", "D5") // DOKUMEN PABEAN
	f.MergeCell(sheetName, "E5", "F5") // BUKTI PENERIMAAN BARANG
	f.MergeCell(sheetName, "G5", "G6") // PENGIRIM BARANG
	f.MergeCell(sheetName, "H5", "H6") // KODE BARANG
	f.MergeCell(sheetName, "I5", "I6") // NAMA BARANG
	f.MergeCell(sheetName, "J5", "J6") // JUMLAH
	f.MergeCell(sheetName, "K5", "K6") // SATUAN
	f.MergeCell(sheetName, "L5", "L6") // VALAS
	f.MergeCell(sheetName, "M5", "M6") // NILAI

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
	f.SetCellStyle(sheetName, "A5", "M6", headerStyle)

	// Add data rows starting from row 7
	for i, entryProduct := range data {
		row := i + 7
		
		// Set values for each column
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)                                    // No
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), entryProduct.JenisPabean)               // Jenis Pabean
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), entryProduct.NoPabean)                  // No Pabean
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), entryProduct.TglPabean.Format("2006-01-02")) // Tgl Pabean
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), entryProduct.VendDlvNo)                 // Vend Dlv No
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), entryProduct.TransDate.Format("2006-01-02")) // Trans Date
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), entryProduct.VendorName)                // Vendor Name
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), entryProduct.ItemCode)                  // Item Code
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), entryProduct.ItemName)                  // Item Name
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), entryProduct.RcvQty.String())           // Rcv Qty
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), entryProduct.PchUnit)                   // Pch Unit
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), entryProduct.CurrCode)                  // Curr Code
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), entryProduct.NetAmount.String())        // Net Amount
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
		lastRow := len(data) + 6
		f.SetCellStyle(sheetName, "A7", fmt.Sprintf("M%d", lastRow), dataStyle)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 5)   // No
	f.SetColWidth(sheetName, "B", "B", 12)  // Jenis Pabean
	f.SetColWidth(sheetName, "C", "C", 20)  // No Pabean
	f.SetColWidth(sheetName, "D", "D", 12)  // Tgl Pabean
	f.SetColWidth(sheetName, "E", "E", 15)  // Vend Dlv No
	f.SetColWidth(sheetName, "F", "F", 12)  // Trans Date
	f.SetColWidth(sheetName, "G", "G", 25)  // Vendor Name
	f.SetColWidth(sheetName, "H", "H", 15)  // Item Code
	f.SetColWidth(sheetName, "I", "I", 30)  // Item Name
	f.SetColWidth(sheetName, "J", "J", 12)  // Rcv Qty
	f.SetColWidth(sheetName, "K", "K", 10)  // Pch Unit
	f.SetColWidth(sheetName, "L", "L", 8)   // Curr Code
	f.SetColWidth(sheetName, "M", "M", 15)  // Net Amount

	// Delete default sheet if it exists
	f.DeleteSheet("Sheet1")

	return f, nil
}

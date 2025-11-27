// File: internal/sheets/formatter.go
package sheets

import (
	"fmt"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
)

// FormatSalesData formats sale records for Google Sheets export
func FormatSalesData(records []*data.SaleExportRecord, exportedBy string) [][]interface{} {
	// Create header row
	header := []interface{}{
		"Transaction ID",
		"Date",
		"Time",
		"User Email",
		"User Name",
		"Product Name",
		"Unit Price",
		"Quantity",
		"Total Amount",
	}

	// Initialize data slice with header
	formattedData := [][]interface{}{header}

	// Add data rows
	for _, record := range records {
		row := []interface{}{
			record.TransactionID,
			record.Date,
			record.Time,
			record.UserEmail,
			record.UserName,
			record.ProductName,
			fmt.Sprintf("%.2f", record.ProductPrice),
			record.Quantity,
			fmt.Sprintf("%.2f", record.TotalAmount),
		}
		formattedData = append(formattedData, row)
	}

	// Add summary section
	if len(records) > 0 {
		formattedData = append(formattedData, []interface{}{})
		formattedData = append(formattedData, []interface{}{"Summary", "", "", "", "", "", "", "", ""})
		
		// Calculate totals
		totalAmount := 0.0
		totalQuantity := int64(0)
		for _, record := range records {
			totalAmount += record.TotalAmount
			totalQuantity += record.Quantity
		}

		formattedData = append(formattedData, []interface{}{"Total Transactions:", len(records), "", "", "", "", "", "", ""})
		formattedData = append(formattedData, []interface{}{"Total Items Sold:", totalQuantity, "", "", "", "", "", "", ""})
		formattedData = append(formattedData, []interface{}{"Total Revenue:", fmt.Sprintf("%.2f", totalAmount), "", "", "", "", "", "", ""})
	}

	// Add export metadata
	formattedData = append(formattedData, []interface{}{})
	formattedData = append(formattedData, []interface{}{"Export Information", "", "", "", "", "", "", "", ""})
	formattedData = append(formattedData, []interface{}{"Exported By:", exportedBy, "", "", "", "", "", "", ""})
	formattedData = append(formattedData, []interface{}{"Export Date:", time.Now().Format("2006-01-02 15:04:05"), "", "", "", "", "", "", ""})

	return formattedData
}

// FormatSalesSummaryData formats sales summary data for Google Sheets export
func FormatSalesSummaryData(records []*data.SaleExportRecord, exportedBy string) [][]interface{} {
	// Create header row
	header := []interface{}{
		"Product Name",
		"Total Quantity Sold",
		"Total Revenue",
		"Number of Transactions",
		"Average Sale Amount",
	}

	// Initialize data slice with header
	formattedData := [][]interface{}{header}

	// Aggregate data by product
	productStats := make(map[string]*ProductSummary)

	for _, record := range records {
		if _, exists := productStats[record.ProductName]; !exists {
			productStats[record.ProductName] = &ProductSummary{
				ProductName:   record.ProductName,
				TotalQuantity: 0,
				TotalRevenue:  0,
				TransactionCount: 0,
			}
		}

		stats := productStats[record.ProductName]
		stats.TotalQuantity += record.Quantity
		stats.TotalRevenue += record.TotalAmount
		stats.TransactionCount++
	}

	// Add data rows
	grandTotal := 0.0
	grandQuantity := int64(0)
	grandTransactions := 0

	for _, stats := range productStats {
		avgAmount := 0.0
		if stats.TransactionCount > 0 {
			avgAmount = stats.TotalRevenue / float64(stats.TransactionCount)
		}

		row := []interface{}{
			stats.ProductName,
			stats.TotalQuantity,
			fmt.Sprintf("%.2f", stats.TotalRevenue),
			stats.TransactionCount,
			fmt.Sprintf("%.2f", avgAmount),
		}
		formattedData = append(formattedData, row)

		grandTotal += stats.TotalRevenue
		grandQuantity += stats.TotalQuantity
		grandTransactions += stats.TransactionCount
	}

	// Add grand totals
	if len(productStats) > 0 {
		formattedData = append(formattedData, []interface{}{})
		formattedData = append(formattedData, []interface{}{
			"Grand Total",
			grandQuantity,
			fmt.Sprintf("%.2f", grandTotal),
			grandTransactions,
			fmt.Sprintf("%.2f", grandTotal/float64(grandTransactions)),
		})
	}

	// Add export metadata
	formattedData = append(formattedData, []interface{}{})
	formattedData = append(formattedData, []interface{}{"Export Information", "", "", "", ""})
	formattedData = append(formattedData, []interface{}{"Exported By:", exportedBy, "", "", ""})
	formattedData = append(formattedData, []interface{}{"Export Date:", time.Now().Format("2006-01-02 15:04:05"), "", "", ""})

	return formattedData
}

// ProductSummary holds aggregated statistics for a product
type ProductSummary struct {
	ProductName      string
	TotalQuantity    int64
	TotalRevenue     float64
	TransactionCount int
}

// FormatDateRange formats date range for display
func FormatDateRange(startDate, endDate *time.Time) string {
	if startDate == nil && endDate == nil {
		return "All Time"
	}

	if startDate != nil && endDate != nil {
		return fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}

	if startDate != nil {
		return fmt.Sprintf("From %s", startDate.Format("2006-01-02"))
	}

	if endDate != nil {
		return fmt.Sprintf("Until %s", endDate.Format("2006-01-02"))
	}

	return "All Time"
}

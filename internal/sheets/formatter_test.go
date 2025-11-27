// File: internal/sheets/formatter_test.go
package sheets

import (
	"testing"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
)

func TestFormatSalesData(t *testing.T) {
	// Create test records
	soldAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	records := []*data.SaleExportRecord{
		{
			TransactionID: 1,
			Date:          "2024-01-15",
			Time:          "10:30:00",
			UserEmail:     "user1@example.com",
			UserName:      "John Doe",
			ProductName:   "Product A",
			ProductPrice:  10.50,
			Quantity:      2,
			TotalAmount:   21.00,
			SoldAt:        soldAt,
		},
		{
			TransactionID: 2,
			Date:          "2024-01-15",
			Time:          "11:00:00",
			UserEmail:     "user2@example.com",
			UserName:      "Jane Smith",
			ProductName:   "Product B",
			ProductPrice:  15.00,
			Quantity:      1,
			TotalAmount:   15.00,
			SoldAt:        soldAt,
		},
	}

	exportedBy := "Admin User (admin@example.com)"
	result := FormatSalesData(records, exportedBy)

	// Verify header row
	if len(result) < 1 {
		t.Fatal("Expected at least a header row")
	}

	headerRow := result[0]
	expectedHeaders := []string{
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

	if len(headerRow) != len(expectedHeaders) {
		t.Fatalf("Expected %d columns in header, got %d", len(expectedHeaders), len(headerRow))
	}

	for i, expected := range expectedHeaders {
		if headerRow[i] != expected {
			t.Errorf("Header column %d: expected %q, got %q", i, expected, headerRow[i])
		}
	}

	// Verify data rows exist
	if len(result) < 3 { // header + 2 data rows
		t.Fatalf("Expected at least 3 rows (header + 2 data), got %d", len(result))
	}

	// Verify first data row
	firstDataRow := result[1]
	if firstDataRow[0] != int64(1) {
		t.Errorf("Expected transaction ID 1, got %v", firstDataRow[0])
	}
	if firstDataRow[3] != "user1@example.com" {
		t.Errorf("Expected user email user1@example.com, got %v", firstDataRow[3])
	}
}

func TestFormatSalesSummaryData(t *testing.T) {
	// Create test records
	soldAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	records := []*data.SaleExportRecord{
		{
			TransactionID: 1,
			ProductName:   "Product A",
			ProductPrice:  10.00,
			Quantity:      2,
			TotalAmount:   20.00,
			SoldAt:        soldAt,
		},
		{
			TransactionID: 2,
			ProductName:   "Product A",
			ProductPrice:  10.00,
			Quantity:      3,
			TotalAmount:   30.00,
			SoldAt:        soldAt,
		},
		{
			TransactionID: 3,
			ProductName:   "Product B",
			ProductPrice:  15.00,
			Quantity:      1,
			TotalAmount:   15.00,
			SoldAt:        soldAt,
		},
	}

	exportedBy := "Admin User"
	result := FormatSalesSummaryData(records, exportedBy)

	// Verify header row
	if len(result) < 1 {
		t.Fatal("Expected at least a header row")
	}

	headerRow := result[0]
	expectedHeaders := []string{
		"Product Name",
		"Total Quantity Sold",
		"Total Revenue",
		"Number of Transactions",
		"Average Sale Amount",
	}

	if len(headerRow) != len(expectedHeaders) {
		t.Fatalf("Expected %d columns in header, got %d", len(expectedHeaders), len(headerRow))
	}

	// Verify at least 2 product rows (Product A and Product B)
	if len(result) < 3 { // header + 2 products + other rows
		t.Fatalf("Expected at least 3 rows, got %d", len(result))
	}
}

func TestGenerateSheetName(t *testing.T) {
	tests := []struct {
		name       string
		exportType string
		startDate  *time.Time
		endDate    *time.Time
		contains   string
	}{
		{
			name:       "Daily export with date",
			exportType: "daily",
			startDate:  ptrTime(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)),
			endDate:    nil,
			contains:   "Daily_Sales_2024-01-15",
		},
		{
			name:       "Monthly export with date",
			exportType: "monthly",
			startDate:  ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			endDate:    nil,
			contains:   "Monthly_Sales_2024-01",
		},
		{
			name:       "Custom export with date range",
			exportType: "custom",
			startDate:  ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			endDate:    ptrTime(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)),
			contains:   "Sales_2024-01-01_to_2024-01-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSheetName(tt.exportType, tt.startDate, tt.endDate)
			if result != tt.contains {
				t.Errorf("GenerateSheetName() = %q, want %q", result, tt.contains)
			}
		})
	}
}

func TestFormatDateRange(t *testing.T) {
	tests := []struct {
		name      string
		startDate *time.Time
		endDate   *time.Time
		expected  string
	}{
		{
			name:      "Both dates nil",
			startDate: nil,
			endDate:   nil,
			expected:  "All Time",
		},
		{
			name:      "Both dates provided",
			startDate: ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			endDate:   ptrTime(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)),
			expected:  "2024-01-01 to 2024-01-31",
		},
		{
			name:      "Only start date",
			startDate: ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			endDate:   nil,
			expected:  "From 2024-01-01",
		},
		{
			name:      "Only end date",
			startDate: nil,
			endDate:   ptrTime(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)),
			expected:  "Until 2024-01-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDateRange(tt.startDate, tt.endDate)
			if result != tt.expected {
				t.Errorf("FormatDateRange() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Helper function to create a pointer to a time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}

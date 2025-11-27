// File: internal/sheets/service.go
package sheets

import (
	"fmt"
	"time"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
)

// Service provides high-level operations for Google Sheets exports
type Service struct {
	client *Client
}

// NewService creates a new sheets service
func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

// ExportSales exports sales records to a Google Sheet
func (s *Service) ExportSales(sheetName string, records []*data.SaleExportRecord, exportedBy string) (int, error) {
	// Create or get sheet
	_, err := s.client.CreateSheet(sheetName)
	if err != nil {
		return 0, fmt.Errorf("failed to create sheet: %v", err)
	}

	// Clear existing data
	if err := s.client.ClearSheet(sheetName); err != nil {
		return 0, fmt.Errorf("failed to clear sheet: %v", err)
	}

	// Prepare data
	formattedData := FormatSalesData(records, exportedBy)

	// Write data
	if err := s.client.WriteData(sheetName, "A1", formattedData); err != nil {
		return 0, fmt.Errorf("failed to write data: %v", err)
	}

	// Format header
	if err := s.client.FormatHeader(sheetName, len(formattedData[0])); err != nil {
		return 0, fmt.Errorf("failed to format header: %v", err)
	}

	return len(records), nil
}

// GetSpreadsheetInfo returns information about the configured spreadsheet
func (s *Service) GetSpreadsheetInfo() (map[string]interface{}, error) {
	spreadsheet, err := s.client.GetSpreadsheet()
	if err != nil {
		return nil, err
	}

	sheets := make([]string, 0, len(spreadsheet.Sheets))
	for _, sheet := range spreadsheet.Sheets {
		sheets = append(sheets, sheet.Properties.Title)
	}

	info := map[string]interface{}{
		"spreadsheet_id":    spreadsheet.SpreadsheetId,
		"spreadsheet_title": spreadsheet.Properties.Title,
		"sheets":            sheets,
		"sheet_count":       len(sheets),
	}

	return info, nil
}

// TestConnection tests the connection to Google Sheets
func (s *Service) TestConnection() error {
	_, err := s.client.GetSpreadsheet()
	return err
}

// ExportSalesSummary exports a summary of sales to a Google Sheet
func (s *Service) ExportSalesSummary(sheetName string, records []*data.SaleExportRecord, exportedBy string) (int, error) {
	// Create or get sheet
	_, err := s.client.CreateSheet(sheetName)
	if err != nil {
		return 0, fmt.Errorf("failed to create sheet: %v", err)
	}

	// Clear existing data
	if err := s.client.ClearSheet(sheetName); err != nil {
		return 0, fmt.Errorf("failed to clear sheet: %v", err)
	}

	// Prepare summary data
	formattedData := FormatSalesSummaryData(records, exportedBy)

	// Write data
	if err := s.client.WriteData(sheetName, "A1", formattedData); err != nil {
		return 0, fmt.Errorf("failed to write data: %v", err)
	}

	// Format header
	if err := s.client.FormatHeader(sheetName, len(formattedData[0])); err != nil {
		return 0, fmt.Errorf("failed to format header: %v", err)
	}

	return len(formattedData) - 1, nil // -1 for header row
}

// GenerateSheetName generates a sheet name based on export type and date range
func GenerateSheetName(exportType string, startDate, endDate *time.Time) string {
	now := time.Now()

	switch exportType {
	case "daily":
		if startDate != nil {
			return fmt.Sprintf("Daily_Sales_%s", startDate.Format("2006-01-02"))
		}
		return fmt.Sprintf("Daily_Sales_%s", now.Format("2006-01-02"))
	case "monthly":
		if startDate != nil {
			return fmt.Sprintf("Monthly_Sales_%s", startDate.Format("2006-01"))
		}
		return fmt.Sprintf("Monthly_Sales_%s", now.Format("2006-01"))
	case "custom":
		if startDate != nil && endDate != nil {
			return fmt.Sprintf("Sales_%s_to_%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		}
		return fmt.Sprintf("Sales_Export_%s", now.Format("2006-01-02"))
	default:
		return fmt.Sprintf("Sales_Export_%s", now.Format("2006-01-02_15-04-05"))
	}
}

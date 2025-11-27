// File: internal/sheets/client.go
package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Client wraps the Google Sheets API client
type Client struct {
	service       *sheets.Service
	spreadsheetID string
}

// Config holds configuration for the Google Sheets client
type Config struct {
	ServiceAccountKeyPath string
	SpreadsheetID         string
}

// NewClient creates a new Google Sheets client with service account authentication
func NewClient(cfg Config) (*Client, error) {
	ctx := context.Background()

	// Read service account key file
	credentials, err := os.ReadFile(cfg.ServiceAccountKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read service account key file: %v", err)
	}

	// Create OAuth2 config from service account
	config, err := google.JWTConfigFromJSON(credentials, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse service account key: %v", err)
	}

	// Create HTTP client
	httpClient := config.Client(ctx)

	// Create Sheets service
	service, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create sheets service: %v", err)
	}

	return &Client{
		service:       service,
		spreadsheetID: cfg.SpreadsheetID,
	}, nil
}

// NewClientFromJSON creates a new Google Sheets client from JSON credentials
func NewClientFromJSON(credentialsJSON string, spreadsheetID string) (*Client, error) {
	ctx := context.Background()

	// Parse JSON credentials
	config, err := google.JWTConfigFromJSON([]byte(credentialsJSON), sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse service account key: %v", err)
	}

	// Create HTTP client
	httpClient := config.Client(ctx)

	// Create Sheets service
	service, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create sheets service: %v", err)
	}

	return &Client{
		service:       service,
		spreadsheetID: spreadsheetID,
	}, nil
}

// GetSpreadsheet retrieves spreadsheet metadata
func (c *Client) GetSpreadsheet() (*sheets.Spreadsheet, error) {
	spreadsheet, err := c.service.Spreadsheets.Get(c.spreadsheetID).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve spreadsheet: %v", err)
	}
	return spreadsheet, nil
}

// GetSheetByName retrieves a sheet by name
func (c *Client) GetSheetByName(sheetName string) (*sheets.Sheet, error) {
	spreadsheet, err := c.GetSpreadsheet()
	if err != nil {
		return nil, err
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return sheet, nil
		}
	}

	return nil, fmt.Errorf("sheet %s not found", sheetName)
}

// CreateSheet creates a new sheet in the spreadsheet
func (c *Client) CreateSheet(sheetName string) (*sheets.Sheet, error) {
	// Check if sheet already exists
	existingSheet, _ := c.GetSheetByName(sheetName)
	if existingSheet != nil {
		return existingSheet, nil
	}

	// Create new sheet
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: sheetName,
					},
				},
			},
		},
	}

	resp, err := c.service.Spreadsheets.BatchUpdate(c.spreadsheetID, req).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create sheet: %v", err)
	}

	if len(resp.Replies) > 0 && resp.Replies[0].AddSheet != nil {
		return &sheets.Sheet{
			Properties: resp.Replies[0].AddSheet.Properties,
		}, nil
	}

	return nil, fmt.Errorf("failed to create sheet")
}

// WriteData writes data to a sheet starting at the specified range
func (c *Client) WriteData(sheetName string, startRange string, data [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: data,
	}

	rangeSpec := fmt.Sprintf("%s!%s", sheetName, startRange)

	_, err := c.service.Spreadsheets.Values.Update(
		c.spreadsheetID,
		rangeSpec,
		valueRange,
	).ValueInputOption("RAW").Do()

	if err != nil {
		return fmt.Errorf("unable to write data: %v", err)
	}

	return nil
}

// AppendData appends data to a sheet
func (c *Client) AppendData(sheetName string, data [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: data,
	}

	_, err := c.service.Spreadsheets.Values.Append(
		c.spreadsheetID,
		sheetName,
		valueRange,
	).ValueInputOption("RAW").Do()

	if err != nil {
		return fmt.Errorf("unable to append data: %v", err)
	}

	return nil
}

// ClearSheet clears all data from a sheet
func (c *Client) ClearSheet(sheetName string) error {
	clearRange := fmt.Sprintf("%s!A1:Z10000", sheetName)

	_, err := c.service.Spreadsheets.Values.Clear(
		c.spreadsheetID,
		clearRange,
		&sheets.ClearValuesRequest{},
	).Do()

	if err != nil {
		return fmt.Errorf("unable to clear sheet: %v", err)
	}

	return nil
}

// FormatHeader formats the header row with bold text and background color
func (c *Client) FormatHeader(sheetName string, numColumns int) error {
	sheet, err := c.GetSheetByName(sheetName)
	if err != nil {
		return err
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				RepeatCell: &sheets.RepeatCellRequest{
					Range: &sheets.GridRange{
						SheetId:          sheet.Properties.SheetId,
						StartRowIndex:    0,
						EndRowIndex:      1,
						StartColumnIndex: 0,
						EndColumnIndex:   int64(numColumns),
					},
					Cell: &sheets.CellData{
						UserEnteredFormat: &sheets.CellFormat{
							BackgroundColor: &sheets.Color{
								Red:   0.9,
								Green: 0.9,
								Blue:  0.9,
							},
							TextFormat: &sheets.TextFormat{
								Bold: true,
							},
						},
					},
					Fields: "userEnteredFormat(backgroundColor,textFormat)",
				},
			},
		},
	}

	_, err = c.service.Spreadsheets.BatchUpdate(c.spreadsheetID, req).Do()
	if err != nil {
		return fmt.Errorf("unable to format header: %v", err)
	}

	return nil
}

// ValidateCredentials validates the service account credentials
func ValidateCredentials(credentialsJSON string) error {
	var creds map[string]interface{}
	if err := json.Unmarshal([]byte(credentialsJSON), &creds); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	requiredFields := []string{"type", "project_id", "private_key_id", "private_key", "client_email"}
	for _, field := range requiredFields {
		if _, ok := creds[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	if creds["type"] != "service_account" {
		return fmt.Errorf("invalid credential type: expected service_account, got %s", creds["type"])
	}

	return nil
}

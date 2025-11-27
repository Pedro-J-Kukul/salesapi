# Google Sheets Integration

## Overview

The Sales API now supports automatic export of sales records to Google Sheets for financial auditing and treasurer access. This integration provides real-time sales data export, automated record keeping, and easy access for financial reviews.

## Features

- Export sales records to Google Sheets with customizable date ranges
- Multiple export types: daily, monthly, custom, or all-time
- Export history tracking with status monitoring
- Formatted exports with headers, summaries, and audit information
- Role-based access control for export operations
- Secure service account authentication

## Setup

### 1. Create a Google Cloud Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Sheets API for your project
4. Create a service account:
   - Go to IAM & Admin > Service Accounts
   - Click "Create Service Account"
   - Name it (e.g., "sales-api-sheets")
   - Grant it the "Editor" role
   - Create and download a JSON key file

### 2. Create a Google Spreadsheet

1. Create a new Google Spreadsheet
2. Share it with the service account email (found in the JSON key file)
3. Give it "Editor" permissions
4. Copy the Spreadsheet ID from the URL:
   - URL format: `https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/edit`

### 3. Configure the Sales API

Add the following environment variables to your `.env` file or pass them as flags:

```bash
# Enable Google Sheets integration
GOOGLE_SHEETS_ENABLED=true

# Service account credentials (JSON format)
GOOGLE_SERVICE_ACCOUNT_KEY='{"type":"service_account","project_id":"...","private_key_id":"...","private_key":"...","client_email":"...","client_id":"...","auth_uri":"...","token_uri":"...","auth_provider_x509_cert_url":"...","client_x509_cert_url":"..."}'

# Spreadsheet ID
GOOGLE_SHEETS_SPREADSHEET_ID=your_spreadsheet_id_here
```

Alternatively, use command-line flags:

```bash
./api -sheets-enabled=true \
      -sheets-service-account-key='...' \
      -sheets-spreadsheet-id='...'
```

### 4. Run Database Migrations

Apply the new migration to create the export_history table:

```bash
make migrate/up
```

### 5. Update Permissions

Run the updated seeds to add export permissions:

```bash
make seed/up
```

Grant export permissions to users who need them:
- `exports:create` - Create exports
- `exports:read` - View export history and sheets info

## API Endpoints

### 1. Export Sales to Google Sheets

**Endpoint:** `POST /v1/exports/sales`

**Authentication:** Required

**Permissions:** `exports:create`

**Request Body:**
```json
{
  "export_type": "daily",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "sheet_name": "January_2024_Sales"
}
```

**Parameters:**
- `export_type` (required): Type of export - `daily`, `monthly`, `custom`, or `all`
- `start_date` (optional): Start date in YYYY-MM-DD format
- `end_date` (optional): End date in YYYY-MM-DD format
- `sheet_name` (optional): Custom sheet name. If not provided, one will be generated automatically

**Response:**
```json
{
  "export": {
    "id": 1,
    "user_id": 123,
    "export_type": "daily",
    "spreadsheet_id": "abc123...",
    "sheet_name": "Daily_Sales_2024-01-15",
    "row_count": 150,
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z",
    "status": "completed",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "message": "Successfully exported 150 sales records to sheet 'January_2024_Sales'"
}
```

### 2. Get Export History

**Endpoint:** `GET /v1/exports/history`

**Authentication:** Required

**Permissions:** `exports:read`

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Results per page (default: 20)
- `user_id` (optional): Filter by user ID
- `export_type` (optional): Filter by export type
- `status` (optional): Filter by status (pending, completed, failed)
- `min_date` (optional): Filter by minimum creation date (YYYY-MM-DD)
- `max_date` (optional): Filter by maximum creation date (YYYY-MM-DD)
- `sort` (optional): Sort field (default: created_at)

**Example:**
```
GET /v1/exports/history?export_type=daily&status=completed&page=1&page_size=20
```

**Response:**
```json
{
  "exports": [
    {
      "id": 1,
      "user_id": 123,
      "export_type": "daily",
      "spreadsheet_id": "abc123...",
      "sheet_name": "Daily_Sales_2024-01-15",
      "row_count": 150,
      "status": "completed",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "metadata": {
    "current_page": 1,
    "page_size": 20,
    "first_page": 1,
    "last_page": 1,
    "total_records": 1
  }
}
```

### 3. Get Spreadsheet Information

**Endpoint:** `GET /v1/exports/sheets`

**Authentication:** Required

**Permissions:** `exports:read`

**Response:**
```json
{
  "sheets_info": {
    "spreadsheet_id": "abc123...",
    "spreadsheet_title": "Sales Data",
    "sheets": [
      "Daily_Sales_2024-01-15",
      "Monthly_Summary_2024-01"
    ],
    "sheet_count": 2
  }
}
```

## Export Data Format

Exported sheets include the following columns:

### Sales Records Export
- Transaction ID
- Date
- Time
- User Email
- User Name
- Product Name
- Unit Price
- Quantity
- Total Amount

### Summary Section (at the bottom)
- Total Transactions
- Total Items Sold
- Total Revenue

### Export Metadata
- Exported By (user name and email)
- Export Date

## Security Considerations

1. **Service Account Credentials**
   - Store credentials securely using environment variables
   - Never commit credentials to version control
   - Rotate credentials periodically

2. **Spreadsheet Access**
   - Only share spreadsheets with authorized service accounts
   - Use separate spreadsheets for different sensitivity levels
   - Regularly review access permissions

3. **API Permissions**
   - Grant `exports:create` only to trusted users (admins, managers)
   - Grant `exports:read` to users who need to view export history
   - Use authentication tokens with appropriate expiration

4. **Audit Trail**
   - All exports are logged in export_history table
   - Each export includes user information and timestamps
   - Monitor export activity regularly

## Troubleshooting

### "Google Sheets is not configured" Error
- Ensure `GOOGLE_SHEETS_ENABLED=true` is set
- Verify service account key JSON is valid
- Check that spreadsheet ID is correct

### "Failed to export to Google Sheets" Error
- Verify the service account has editor access to the spreadsheet
- Check that the Google Sheets API is enabled in your GCP project
- Ensure the service account credentials are valid and not expired

### Empty Export
- Verify there are sales records in the specified date range
- Check that the database connection is working
- Review export history for any error messages

## Example Usage

### Export Today's Sales
```bash
curl -X POST https://your-api.com/v1/exports/sales \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "export_type": "daily",
    "start_date": "2024-01-15",
    "end_date": "2024-01-15"
  }'
```

### Export Monthly Sales
```bash
curl -X POST https://your-api.com/v1/exports/sales \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "export_type": "monthly",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "sheet_name": "January_2024_Sales"
  }'
```

### View Export History
```bash
curl -X GET "https://your-api.com/v1/exports/history?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Best Practices

1. **Regular Exports**
   - Set up automated exports for daily/monthly financial reviews
   - Export data before month-end closing

2. **Naming Convention**
   - Use descriptive sheet names with dates
   - Examples: "Daily_Sales_2024-01-15", "Monthly_Summary_2024-01"

3. **Data Validation**
   - Review exported data for accuracy
   - Cross-check totals with database records

4. **Access Control**
   - Limit export permissions to financial staff and managers
   - Review and update permissions regularly

5. **Monitoring**
   - Monitor export_history for failed exports
   - Set up alerts for export failures
   - Review export activity logs periodically

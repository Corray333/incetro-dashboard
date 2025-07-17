package sheets

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Corray333/employee_dashboard/internal/domains/client/entities/client"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	"github.com/spf13/viper"
	"google.golang.org/api/sheets/v4"
)

type ClientSheetsRepository struct {
	client *gsheets.Client
}

func NewClientSheetsRepository(client *gsheets.Client) *ClientSheetsRepository {
	return &ClientSheetsRepository{
		client: client,
	}
}

func (r *ClientSheetsRepository) UpdateSheetsClients(ctx context.Context, sheetID string, clients []client.Client) error {
	if len(clients) == 0 {
		return nil
	}

	// Use default sheet ID if not provided
	if sheetID == "" {
		sheetID = viper.GetString("sheets.id")
	}

	appendRange := viper.GetString("sheets.clients_sheet") + "!A2:"
	rowLen := len(entityToSheetsClient(&clients[0]))
	lastColLetter := string(rune('A' + rowLen - 1))
	appendRange += lastColLetter

	clearValues := &sheets.ClearValuesRequest{}
	_, err := r.client.Svc().Spreadsheets.Values.Clear(sheetID, appendRange, clearValues).Do()
	if err != nil {
		slog.Error("Error clearing old values", "error", err)
		return err
	}

	var vr sheets.ValueRange

	for _, client := range clients {
		vr.Values = append(vr.Values, entityToSheetsClient(&client))
	}

	_, err = r.client.Svc().Spreadsheets.Values.Append(sheetID, appendRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", "error", err)
		return err
	}

	// Clear formatting to prevent inheriting bold text from header row
	err = r.clearFormatting(ctx, sheetID, appendRange, len(clients))
	if err != nil {
		slog.Error("Error clearing formatting", "error", err)
		// Don't return error here as data was successfully added
	}

	return nil
}

// getSheetIDByName retrieves the sheet ID by sheet name
func (r *ClientSheetsRepository) getSheetIDByName(ctx context.Context, spreadsheetID, sheetName string) (int64, error) {
	spreadsheet, err := r.client.Svc().Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return 0, err
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return sheet.Properties.SheetId, nil
		}
	}

	return 0, fmt.Errorf("sheet with name '%s' not found", sheetName)
}

// clearFormatting removes formatting from the specified range to prevent inheriting bold text from header
func (r *ClientSheetsRepository) clearFormatting(ctx context.Context, sheetID, appendRange string, rowCount int) error {
	// Get sheet name from config
	sheetName := viper.GetString("sheets.clients_sheet")

	// Get the actual sheet ID by name
	actualSheetID, err := r.getSheetIDByName(ctx, sheetID, sheetName)
	if err != nil {
		slog.Error("Error getting sheet ID by name", "error", err, "sheetName", sheetName)
		// Fallback to sheet ID 0 if we can't find the sheet
		actualSheetID = 0
	}

	// Create a request to clear formatting for the data rows
	request := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				RepeatCell: &sheets.RepeatCellRequest{
					Range: &sheets.GridRange{
						SheetId:          actualSheetID,
						StartRowIndex:    1,                   // Start from row 2 (index 1)
						EndRowIndex:      int64(1 + rowCount), // End at last data row
						StartColumnIndex: 0,                   // Column A
						EndColumnIndex:   5,                   // Up to column E (adjust based on your data)
					},
					Cell: &sheets.CellData{
						UserEnteredFormat: &sheets.CellFormat{
							TextFormat: &sheets.TextFormat{
								Bold: false,
							},
						},
					},
					Fields: "userEnteredFormat.textFormat.bold",
				},
			},
		},
	}

	_, err = r.client.Svc().Spreadsheets.BatchUpdate(sheetID, request).Do()
	return err
}

func entityToSheetsClient(client *client.Client) []interface{} {

	return []interface{}{
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(client.ID.String(), "-", "")), client.Name),
		string(client.Status),
		client.Source,
		client.UniqueID,
	}
}

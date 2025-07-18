package sheets

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	"github.com/spf13/viper"
	"google.golang.org/api/sheets/v4"
)

type TimeSheetsRepository struct {
	client *gsheets.Client
}

func NewTimeSheetsRepository(client *gsheets.Client) *TimeSheetsRepository {
	return &TimeSheetsRepository{
		client: client,
	}
}

// getSheetIDByName retrieves the sheet ID by sheet name
func (r *TimeSheetsRepository) getSheetIDByName(ctx context.Context, spreadsheetID, sheetName string) (int64, error) {
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

func (r *TimeSheetsRepository) UpdateSheetsTimes(ctx context.Context, sheetID string, times []entity_time.Time) error {
	if len(times) == 0 {
		return nil
	}

	sheetName := viper.GetString("sheets.time_sheet")
	appendRange := viper.GetString("sheets.time_sheet") + "!A3:"
	rowLen := len(entityToSheetsTime(&times[0]))
	lastColLetter := string(rune('A' + rowLen - 1))
	appendRange += lastColLetter

	// Get the actual sheet ID by name
	actualSheetID, err := r.getSheetIDByName(ctx, sheetID, sheetName)
	if err != nil {
		slog.Error("Error getting sheet ID by name", "error", err, "sheetName", sheetName)
		// Fallback to sheet ID 0 if we can't find the sheet
		actualSheetID = 0
	}

	clearValues := &sheets.ClearValuesRequest{}
	if _, err := r.client.Svc().Spreadsheets.Values.Clear(sheetID, sheetName+"!A2:"+lastColLetter, clearValues).Do(); err != nil {
		slog.Error("Error clearing old values", "error", err)
		return err
	}

	deleteRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDimension: &sheets.DeleteDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    actualSheetID,
						Dimension:  "ROWS",
						StartIndex: 2,                 // Row 3 (0-indexed)
						EndIndex:   int64(len(times)), // Delete all existing data rows
					},
				},
			},
		},
	}

	_, err = r.client.Svc().Spreadsheets.BatchUpdate(sheetID, deleteRequest).Do()
	if err != nil {
		slog.Error("Error deleting old rows", "error", err)
		return err
	}

	var vr sheets.ValueRange

	for _, time := range times {
		vr.Values = append(vr.Values, entityToSheetsTime(&time))
	}

	_, err = r.client.Svc().Spreadsheets.Values.Append(sheetID, appendRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", "error", err)
		return err
	}

	return nil
}

func entityToSheetsTime(time *entity_time.Time) []interface{} {
	return []interface{}{
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(time.ID.String(), "-", "")), strings.ReplaceAll(time.WhatDid, "\"", "\"\"")),
		time.TotalHours,
		time.WorkDate.Format("02/01/2006"),
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(time.TaskID.String(), "-", "")), strings.ReplaceAll(time.TaskName, "\"", "\"\"")),
		time.ProjectName,
		time.WhoDid,
		time.TaskID,
		time.Direction,
		time.TaskEstimate,
		time.CreatedAt.Format("02/01/2006 15:04:05"),
		time.ProjectID,
		time.BH,
		time.SH,
		time.DH,
		time.BHGS,
		time.ProjectStatus,
		time.ID,
		time.Expertise,
		time.PH,
		time.Overtime,
		time.Priority,
	}
}

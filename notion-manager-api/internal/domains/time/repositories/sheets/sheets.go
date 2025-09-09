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
	// Имя листа из конфигурации
	sheetName := viper.GetString("sheets.time_sheet")

	// Если данных нет, очищаем редактируемые столбцы и выходим
	if len(times) == 0 {
		rowLen := len(entityToSheetsTime(&entity_time.Time{}))
		lastColLetter := string(rune('A' + rowLen - 1))
		clearRange := sheetName + "!A2:" + lastColLetter
		if _, err := r.client.Svc().Spreadsheets.Values.Clear(sheetID, clearRange, &sheets.ClearValuesRequest{}).Do(); err != nil {
			slog.Error("Error clearing old values (empty times)", "error", err)
			return err
		}
		return nil
	}

	writeRange := sheetName + "!A2" // запись со 2-й строки

	// Формируем значения для записи
	var vr sheets.ValueRange
	vr.MajorDimension = "ROWS"
	for _, t := range times {
		// vr.Values = append(vr.Values, entityToSheetsTime(&t))
		temp := entityToSheetsTime(&t)
		fmt.Println("Time: ", temp)
		slog.Info("Loading time to sheets", "time", temp)
		vr.Values = append(vr.Values, temp)
	}

	// Получаем фактический ID листа по имени
	actualSheetID, err := r.getSheetIDByName(ctx, sheetID, sheetName)
	if err != nil {
		slog.Error("Error getting sheet ID by name", "error", err, "sheetName", sheetName)
		actualSheetID = 0
	}

	// Убедимся, что в листе достаточно строк, чтобы записать все данные
	spreadsheet, err := r.client.Svc().Spreadsheets.Get(sheetID).Do()
	if err != nil {
		slog.Error("Error getting spreadsheet properties", "error", err)
		return err
	}

	var currentRowCapacity int64
	for _, sh := range spreadsheet.Sheets {
		if sh.Properties.Title == sheetName {
			if sh.Properties.GridProperties != nil {
				currentRowCapacity = sh.Properties.GridProperties.RowCount
			}
			break
		}
	}

	// Требуемое количество строк = 1 (шапка) + количество строк с данными
	requiredRows := int64(1 + len(vr.Values))
	if currentRowCapacity < requiredRows {
		missing := requiredRows - currentRowCapacity
		appendReq := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					AppendDimension: &sheets.AppendDimensionRequest{
						SheetId:   actualSheetID,
						Dimension: "ROWS",
						Length:    missing,
					},
				},
			},
		}

		if _, err := r.client.Svc().Spreadsheets.BatchUpdate(sheetID, appendReq).Do(); err != nil {
			slog.Error("Error appending rows", "error", err)
			return err
		}
	}

	// Очистка диапазона A2:lastCol в редактируемых столбцах перед записью
	rowLen := len(entityToSheetsTime(&times[0]))
	lastColLetter := string(rune('A' + rowLen - 1))
	clearRange := sheetName + "!A2:" + lastColLetter
	if _, err := r.client.Svc().Spreadsheets.Values.Clear(sheetID, clearRange, &sheets.ClearValuesRequest{}).Do(); err != nil {
		slog.Error("Error clearing old values", "error", err)
		return err
	}

	// Перезаписываем значения без удаления строк
	_, err = r.client.Svc().Spreadsheets.Values.Update(sheetID, writeRange, &vr).ValueInputOption("USER_ENTERED").Do()
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

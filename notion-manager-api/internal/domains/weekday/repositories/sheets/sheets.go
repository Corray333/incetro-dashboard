package sheets

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	"github.com/spf13/viper"
	"google.golang.org/api/sheets/v4"
)

type WeekdaySheetsRepository struct {
	client *gsheets.Client
}

func NewWeekdaySheetsRepository(client *gsheets.Client) *WeekdaySheetsRepository {
	return &WeekdaySheetsRepository{
		client: client,
	}
}

// getSheetIDByName retrieves the sheet ID by sheet name
func (r *WeekdaySheetsRepository) getSheetIDByName(ctx context.Context, spreadsheetID, sheetName string) (int64, error) {
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

func (r *WeekdaySheetsRepository) UpdateSheetsWeekdays(ctx context.Context, sheetID string, weekdays []weekday.Weekday) error {
	// Имя листа из конфигурации
	sheetName := viper.GetString("sheets.weekday_sheet")

	// Если данных нет, очищаем редактируемые столбцы и выходим
	if len(weekdays) == 0 {
		rowLen := len(entityToSheetsWeekday(&weekday.Weekday{}))
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
	for _, t := range weekdays {
		// vr.Values = append(vr.Values, entityToSheetsTime(&t))
		temp := entityToSheetsWeekday(&t)
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
	rowLen := len(entityToSheetsWeekday(&weekdays[0]))
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

func entityToSheetsWeekday(weekday *weekday.Weekday) []interface{} {
	return []interface{}{
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(weekday.ID.String(), "-", "")), strings.ReplaceAll(weekday.Reason, "\"", "\"\"")),
		weekday.Employee.Username,
		weekday.Category,
		weekday.PeriodStart.Format("02/01/2006"),
		weekday.PeriodEnd.Format("02/01/2006"),
	}
}

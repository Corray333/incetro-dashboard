package sheets

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	"github.com/spf13/viper"
	"google.golang.org/api/sheets/v4"
)

type TaskSheetsRepository struct {
	client *gsheets.Client
}

func NewTaskSheetsRepository(client *gsheets.Client) *TaskSheetsRepository {
	return &TaskSheetsRepository{
		client: client,
	}
}

// getSheetIDByName retrieves the sheet ID by sheet name
func (r *TaskSheetsRepository) getSheetIDByName(ctx context.Context, spreadsheetID, sheetName string) (int64, error) {
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

func (r *TaskSheetsRepository) UpdateSheetsTasks(ctx context.Context, sheetID string, tasks []task.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	sheetName := viper.GetString("sheets.task_sheet")
	appendRange := sheetName + "!A3:"
	rowLen := len(entityToSheetsTask(&tasks[0]))
	lastColLetter := string(rune('A' + rowLen - 1))
	appendRange += lastColLetter

	// Get the actual sheet ID by name
	actualSheetID, err := r.getSheetIDByName(ctx, sheetID, sheetName)
	if err != nil {
		slog.Error("Error getting sheet ID by name", "error", err, "sheetName", sheetName)
		// Fallback to sheet ID 0 if we can't find the sheet
		actualSheetID = 0
	}

	// Get current sheet data to determine how many rows exist
	readRange := sheetName + "!A:A"
	resp, err := r.client.Svc().Spreadsheets.Values.Get(sheetID, readRange).Do()
	if err != nil {
		slog.Error("Error getting current sheet data", "error", err)
		return err
	}

	// Calculate how many rows currently exist (excluding header rows)
	currentRowCount := len(resp.Values)
	if currentRowCount > 2 { // Only delete if there are data rows (more than 2 header rows)
		deleteRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					DeleteDimension: &sheets.DeleteDimensionRequest{
						Range: &sheets.DimensionRange{
							SheetId:    actualSheetID,
							Dimension:  "ROWS",
							StartIndex: 2,                      // Row 3 (0-indexed)
							EndIndex:   int64(currentRowCount), // Delete only existing data rows
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
	}

	// Now append new data
	var vr sheets.ValueRange

	// Добавляем основные строки задач
	for _, t := range tasks {
		fmt.Printf("Task: %v\n", entityToSheetsTask(&t))
		vr.Values = append(vr.Values, entityToSheetsTask(&t))
	}

	// Генерируем дополнительные строки по статусам и месяцам
	monthStatusRows := generateMonthStatusRows(tasks)
	for _, row := range monthStatusRows {
		// vr.Values = append(vr.Values, row)
		fmt.Printf("MonthStatusRow: %v\n", row)
	}

	// Генерируем дополнительные строки для родительских задач
	parentTaskRows := generateParentTaskRows(tasks)
	for _, row := range parentTaskRows {
		// vr.Values = append(vr.Values, row)
		fmt.Printf("ParentTaskRow: %v\n", row)
	}

	_, err = r.client.Svc().Spreadsheets.Values.Append(sheetID, appendRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", "error", err)
		return err
	}

	return nil
}

// generateMonthStatusRows создает строки для каждой комбинации месяца и статуса
func generateMonthStatusRows(tasks []task.Task) [][]interface{} {
	var rows [][]interface{}

	// Собираем все уникальные комбинации месяц-статус
	monthStatusMap := make(map[string]task.Status)

	for _, t := range tasks {
		// Обрабатываем период от startDate до endDate
		if !t.Start.IsZero() {
			startDate := t.Start
			endDate := t.End
			if endDate.IsZero() {
				endDate = startDate
			}

			// Проходим по всем месяцам в периоде
			current := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
			end := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, endDate.Location())

			for current.Before(end) || current.Equal(end) {
				monthKey := current.Format("2006-01")
				monthStatusKey := monthKey + "|" + string(t.Status)
				monthStatusMap[monthStatusKey] = t.Status

				current = current.AddDate(0, 1, 0)
			}
		}
	}

	// Создаем строки для каждой комбинации
	for monthStatusKey, status := range monthStatusMap {
		parts := strings.Split(monthStatusKey, "|")
		if len(parts) != 2 {
			continue
		}

		monthStr := parts[0]
		parsedTime, err := time.Parse("2006-01", monthStr)
		if err != nil {
			continue
		}

		// Создаем строку с первым днем месяца
		firstDayOfMonth := parsedTime.Format("02/01/2006")

		row := []interface{}{
			"",              // Название задачи пустое
			"",              // Приоритет пустой
			string(status),  // Статус
			firstDayOfMonth, // startDate
			firstDayOfMonth, // endDate
			"",              // Родительская задача пустая
			"",              // Главная задача пустая
			"",              // Направление пустое
			"",              // Экспертиза пустая
			"",              // TotalHours пустое
			"",              // TotalEstimate пустое
		}

		rows = append(rows, row)
	}

	return rows
}

// generateParentTaskRows создает строки для родительских задач по месяцам
func generateParentTaskRows(tasks []task.Task) [][]interface{} {
	var rows [][]interface{}

	for _, t := range tasks {
		fmt.Printf("Childs count: %d\n", t.ChildCount)
		// Проверяем, является ли задача родительской
		if t.ChildCount > 0 && !t.Start.IsZero() {
			startDate := t.Start
			endDate := t.End
			if endDate.IsZero() {
				endDate = startDate
			}

			// Проходим по всем месяцам в периоде родительской задачи
			current := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
			end := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, endDate.Location())

			for current.Before(end) || current.Equal(end) {
				firstDayOfMonth := current.Format("02/01/2006")

				row := []interface{}{
					"",              // Название задачи пустое
					"",              // Приоритет пустой
					"",              // Статус пустой
					firstDayOfMonth, // startDate
					firstDayOfMonth, // endDate
					// Родительская задача с гиперссылкой
					fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(t.ID.String(), "-", "")), strings.ReplaceAll(t.Task, "\"", "\"\"")),
					"", // Главная задача пустая
					"", // Направление пустое
					"", // Экспертиза пустая
					"", // TotalHours пустое
					"", // TotalEstimate пустое
				}

				rows = append(rows, row)
				current = current.AddDate(0, 1, 0)
			}
		}
	}

	return rows
}

// func entityToSheetsTime(time *task.Task) []interface{} {
// 	return []interface{}{
// 		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(time.ID.String(), "-", "")), time.WhatDid),
// 		time.TotalHours,
// 		time.WorkDate.Format("02/01/2006"),
// 		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(time.TaskID.String(), "-", "")), time.TaskName),
// 		time.ProjectName,
// 		time.WhoDid,
// 		time.TaskID,
// 		time.Direction,
// 		time.TaskEstimate,
// 		time.CreatedAt.Format("02/01/2006 15:04:05"),
// 		time.ProjectID,
// 		time.BH,
// 		time.SH,
// 		time.DH,
// 		time.BHGS,
// 		time.ProjectStatus,
// 		time.ID,
// 		time.Expertise,
// 		time.PH,
// 		time.Overtime,
// 		time.Priority,
// 	}
// }

// type Task struct {
// 	ID             uuid.UUID `json:"id"`
// 	CreatedTime    time.Time `json:"created_time"`
// 	LastEditedTime time.Time `json:"last_edited_time"`

// 	Task        string      `json:"task"`
// 	Priority    string      `json:"priority"`
// 	Status      Status      `json:"status"`
// 	ParentID    uuid.UUID   `json:"parent_id"`
// 	Responsible []uuid.UUID `json:"responsible_ids"`
// 	CreatorID   uuid.UUID   `json:"creator_id"`
// 	ExecutorIDs []uuid.UUID `json:"executor_ids"`
// 	ProjectID   uuid.UUID   `json:"project_id"`
// 	Estimate    float64     `json:"estimate"`
// 	Tags        []string    `json:"tags"`
// 	Start       time.Time   `json:"start"`
// 	End         time.Time   `json:"end"`

// 	PreviousID    uuid.UUID   `json:"previous_ids"`
// 	NextID        uuid.UUID   `json:"next_ids"`
// 	TotalHours    float64     `json:"total_hours"`
// 	Subtasks      []uuid.UUID `json:"subtasks_ids"`
// 	TBH           float64     `json:"tbh"`
// 	CP            float64     `json:"cp"`
// 	TotalEstimate float64     `json:"total_estimate"`
// 	PlanFact      float64     `json:"plan_fact"`
// 	Duration      float64     `json:"duration"`
// 	CR            float64     `json:"cr"`
// 	IKP           string      `json:"ikp"`
// }

// ### Что нужно выгрузить из каждой задачи

// 1. Родительская задача (если есть)
// 2. Главная задача
// 3. Название задачи
// 4. Направление задачи
// 5. Приоритет
// 6. Статус
// 7. Дедлайн
// 8. Total оценка (часы)
// 9. Total затраченное время (часы)
// Если задачи не имеют одного из полей — ставить прочерк или n/a.

func entityToSheetsTask(task *task.Task) []interface{} {
	// Форматирование даты начала
	var startDate string
	if !task.Start.IsZero() {
		startDate = task.Start.Format("02/01/2006")
	}

	// Форматирование даты окончания
	var endDate string
	if !task.End.IsZero() {
		endDate = task.End.Format("02/01/2006")
	} else if !task.Start.IsZero() {
		// Если есть дата начала, но нет даты окончания, то дата окончания равна дате начала
		endDate = task.Start.Format("02/01/2006")
	}

	return []interface{}{
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(task.ID.String(), "-", "")), strings.ReplaceAll(task.Task, "\"", "\"\"")),
		task.Priority,
		string(task.Status),
		startDate,
		endDate,
		// task.ParentName, // TODO: change to name with link
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(task.ParentID.String(), "-", "")), strings.ReplaceAll(task.ParentName, "\"", "\"\"")),
		task.MainTask,
		task.GetDirection(),
		task.Expertise,
		task.TotalHours,
		task.TotalEstimate + task.SH,
	}
}

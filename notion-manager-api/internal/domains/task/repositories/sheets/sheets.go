package sheets

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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

func (r *TaskSheetsRepository) UpdateSheetsTasks(ctx context.Context, sheetID string, tasks []task.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	appendRange := viper.GetString("sheets.task_sheet") + "!A3:"
	rowLen := len(entityToSheetsTask(&tasks[0]))
	lastColLetter := string(rune('A' + rowLen - 1))
	appendRange += lastColLetter

	clearValues := &sheets.ClearValuesRequest{}
	_, err := r.client.Svc().Spreadsheets.Values.Clear(sheetID, appendRange, clearValues).Do()
	if err != nil {
		slog.Error("Error clearing old values", "error", err)
		return err
	}

	var vr sheets.ValueRange

	for _, time := range tasks {
		vr.Values = append(vr.Values, entityToSheetsTask(&time))
	}

	_, err = r.client.Svc().Spreadsheets.Values.Append(sheetID, appendRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", "error", err)
		return err
	}

	return nil
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
	return []interface{}{
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(task.ID.String(), "-", "")), task.Task),
		task.Priority,
		string(task.Status),
		task.Start.Format("02/01/2006"),
		task.End.Format("02/01/2006"),
		// task.ParentName, // TODO: change to name with link
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(task.ParentID.String(), "-", "")), task.ParentName),
		task.MainTask,
		task.GetDirection(),
		task.Expertise,
		task.TotalHours,
		task.TotalEstimate,
	}
}

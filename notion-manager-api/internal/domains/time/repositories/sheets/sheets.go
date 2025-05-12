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

func (r *TimeSheetsRepository) UpdateSheetsTimes(ctx context.Context, times []entity_time.Time) error {
	if len(times) == 0 {
		return nil
	}

	appendRange := viper.GetString("sheets.time_sheet") + "!A3:"
	rowLen := len(entityToSheetsTime(&times[0]))
	lastColLetter := string(rune('A' + rowLen - 1))
	appendRange += lastColLetter

	clearValues := &sheets.ClearValuesRequest{}
	_, err := r.client.Svc().Spreadsheets.Values.Clear(viper.GetString("sheets.id"), appendRange, clearValues).Do()
	if err != nil {
		slog.Error("Error clearing old values", "error", err)
		return err
	}

	var vr sheets.ValueRange

	for _, time := range times {
		vr.Values = append(vr.Values, entityToSheetsTime(&time))
	}

	_, err = r.client.Svc().Spreadsheets.Values.Append(viper.GetString("sheets.id"), appendRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", "error", err)
		return err
	}

	return nil
}

func (r *TimeSheetsRepository) UpdateTempSheetsTimes(ctx context.Context, times []entity_time.Time) error {
	if len(times) == 0 {
		return nil
	}

	appendRange := viper.GetString("temp_sheets.time_sheet") + "!A3:"
	rowLen := len(entityToSheetsTime(&times[0]))
	lastColLetter := string(rune('A' + rowLen - 1))
	appendRange += lastColLetter

	clearValues := &sheets.ClearValuesRequest{}
	_, err := r.client.Svc().Spreadsheets.Values.Clear(viper.GetString("temp_sheets.id"), appendRange, clearValues).Do()
	if err != nil {
		slog.Error("Error clearing old values", "error", err)
		return err
	}

	var vr sheets.ValueRange

	for _, time := range times {
		vr.Values = append(vr.Values, entityToSheetsTime(&time))
	}

	_, err = r.client.Svc().Spreadsheets.Values.Append(viper.GetString("temp_sheets.id"), appendRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", "error", err)
		return err
	}

	return nil
}

func entityToSheetsTime(time *entity_time.Time) []interface{} {
	return []interface{}{
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(time.ID.String(), "-", "")), time.WhatDid),
		time.TotalHours,
		time.WorkDate.Format("02/01/2006"),
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(time.TaskID.String(), "-", "")), time.TaskName),
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

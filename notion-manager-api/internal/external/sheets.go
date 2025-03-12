package external

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Corray333/employee_dashboard/internal/entities"
	"github.com/Corray333/employee_dashboard/pkg/notion"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var (
	ErrNoTime = errors.New("no last synced time found")
)

const (
	TimeLayout = "02/01/2006 15:04:05"
)

func GetLastSyncedTime(srv *sheets.Service, spreadsheetId string) (int64, error) {

	readRange := "Time!S2"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		slog.Error("Error getting Google Sheets", slog.String("error", err.Error()))
		return 0, err
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		slog.Error("No data found")
		return 0, ErrNoTime
	}

	lastSynced, err := time.ParseInLocation(TimeLayout, resp.Values[0][0].(string), time.Local)
	if err != nil {
		slog.Error("Error parsing last synced time", slog.String("error", err.Error()))
		return 0, err
	}

	return lastSynced.Unix(), nil
}

func SetLastSyncedTime(lastSyncedTimestamp int64, srv *sheets.Service, spreadsheetId string) error {
	writeRange := "Time!S2"

	lastSynced := time.Unix(lastSyncedTimestamp, 0)

	serialized := lastSynced.Format(TimeLayout)

	// Create a ValueRange with the single value
	var vr sheets.ValueRange
	myval := []interface{}{serialized} // Replace "Your Value" with the value you want to insert
	vr.Values = append(vr.Values, myval)

	// Update the cell with the value
	_, err := srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		slog.Error("Error setting last synced time", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (e *External) UpdateTimeSheet(srv *sheets.Service) error {
	lastSynced, err := GetLastSyncedTime(srv, spreadsheetId)
	if err != nil {
		slog.Error("Error getting last synced time", slog.String("error", err.Error()))
		return err
	}

	readRange := "Time!S:S"
	fullTable, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		slog.Error("Error getting Google Sheets", slog.String("error", err.Error()))
		return err
	}

	var vr sheets.ValueRange
	times, err := e.GetSheetsTimes(lastSynced, "", "")
	if err != nil {
		slog.Error("Error getting times from Notion", slog.String("error", err.Error()))
		return err
	}

	var updateVr []*sheets.ValueRange

	for _, timeRaw := range times {
		date, _ := time.Parse(notion.TIME_LAYOUT_IN, timeRaw.Properties.WorkDate.Date.Start)
		date2, _ := time.Parse("2006-01-02", timeRaw.Properties.WorkDate.Date.Start)
		if date.Before(date2) {
			date = date2
		}
		lastSyncedRaw, _ := time.Parse(notion.TIME_LAYOUT_IN, timeRaw.LastEditedTime)
		lastSynced = lastSyncedRaw.Unix()
		title := ""
		for _, name := range timeRaw.Properties.WhatDid.Title {
			title += name.PlainText
		}

		rawId, err := findRowIndexByID(fullTable, timeRaw.ID)

		if err != nil {
			slog.Error("Error finding row index", slog.String("error", err.Error()))
			return err
		}

		myval := []interface{}{
			fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, timeRaw.URL, title),
			timeRaw.Properties.TotalHours.Number,
			date.Format("02/01/2006"),
			func() string {
				if len(timeRaw.Properties.Task.Relation) == 0 {
					return ""
				}
				url := "https://www.notion.so/"
				id := strings.Join(strings.Split(timeRaw.Properties.Task.Relation[0].ID, "-"), "")
				return fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, url+id, timeRaw.Properties.TaskName.Formula.String)
			}(),
			timeRaw.Properties.ProjectName.Formula.String,
		}
		if len(timeRaw.Properties.WhoDid.People) > 0 {
			myval = append(myval, timeRaw.Properties.WhoDid.People[0].Name)
		} else {
			myval = append(myval, "")
		}
		myval = append(myval, []interface{}{
			// TODO:
			timeRaw.Properties.PayableHours.Formula.Number,
			func() string {
				if len(timeRaw.Properties.Task.Relation) == 0 {
					return ""
				}
				return timeRaw.Properties.Task.Relation[0].ID
			}(),
			timeRaw.Properties.Direction.Select.Name,
			timeRaw.Properties.EstimateHours.Formula.String,
			lastSyncedRaw.Format(TimeLayout),
			func() string {
				if len(timeRaw.Properties.Project.Rollup.Array) == 0 || len(timeRaw.Properties.Project.Rollup.Array[0].Relation) == 0 {
					return ""
				}
				return timeRaw.Properties.Project.Rollup.Array[0].Relation[0].ID
			}(),
			timeRaw.Properties.BH.Formula.Number,
			timeRaw.Properties.SH.Number,
			timeRaw.Properties.DH.Number,
			timeRaw.Properties.BHGS.Formula.String,
			timeRaw.Properties.ProjectStatus.Formula.String,
			timeRaw.ID,
		}...)

		if rawId != -1 {
			uvr := sheets.ValueRange{
				Values: [][]interface{}{myval},
				Range:  fmt.Sprintf("Time!A%d:S%d", rawId, rawId),
			}
			updateVr = append(updateVr, &uvr)
			continue
		}

		vr.Values = append(vr.Values, myval)

	}

	writeRange := "Time!A3:S3"

	_, err = srv.Spreadsheets.Values.BatchUpdate(spreadsheetId, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             updateVr,
	}).Do()

	if err != nil {
		slog.Error("Error updating Google Sheets", slog.String("error", err.Error()))
		return err
	}

	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", slog.String("error", err.Error()))
		return err
	}

	return SetLastSyncedTime(lastSynced, srv, spreadsheetId)
}

func (e *External) UpdateProjectsSheet(srv *sheets.Service, projects []entities.Project) error {

	var vr sheets.ValueRange

	var updateVr []*sheets.ValueRange

	for _, project := range projects {
		title := project.Name

		myval := []interface{}{
			fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", project.ID), title),
			project.Type,
			project.Manager,
			project.Status,
		}

		vr.Values = append(vr.Values, myval)

	}

	writeRange := "Projects!A2:S2"

	_, err := srv.Spreadsheets.Values.BatchUpdate(spreadsheetId, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             updateVr,
	}).Do()

	if err != nil {
		slog.Error("Error updating Google Sheets", slog.String("error", err.Error()))
		return err
	}

	// Clear all old values
	clearRange := "Projects!A2:D"
	clearValues := &sheets.ClearValuesRequest{}
	_, err = srv.Spreadsheets.Values.Clear(spreadsheetId, clearRange, clearValues).Do()
	if err != nil {
		slog.Error("Error clearing old values", slog.String("error", err.Error()))
		return err
	}

	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (e *External) UpdatePeopleSheet(srv *sheets.Service, people []entities.Employee) error {

	var vr sheets.ValueRange

	var updateVr []*sheets.ValueRange

	for _, person := range people {

		myval := []interface{}{
			fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(strings.ToLower(person.ProfileID), "-", "")), person.Username),
			person.ExpertiseName,
			person.Direction,
			person.Status,
			person.Email,
			person.Geo,
			person.Phone,
			fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://t.me/%s", person.Telegram), person.Telegram),
		}

		vr.Values = append(vr.Values, myval)

	}

	writeRange := "People!A2:S2"

	_, err := srv.Spreadsheets.Values.BatchUpdate(spreadsheetId, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             updateVr,
	}).Do()

	if err != nil {
		slog.Error("Error updating Google Sheets", slog.String("error", err.Error()))
		return err
	}

	// Clear all old values
	clearRange := "People!A2:H"
	clearValues := &sheets.ClearValuesRequest{}
	_, err = srv.Spreadsheets.Values.Clear(spreadsheetId, clearRange, clearValues).Do()
	if err != nil {
		slog.Error("Error clearing old values", slog.String("error", err.Error()))
		return err
	}

	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (e *External) UpdateExpertiseSheet(srv *sheets.Service, expertises []entities.Expertise) error {

	var vr sheets.ValueRange

	var updateVr []*sheets.ValueRange

	for _, expertise := range expertises {

		myval := []interface{}{
			expertise.Name,
			expertise.Direction,
			expertise.Description,
		}

		vr.Values = append(vr.Values, myval)

	}

	writeRange := "Expertise!A2:C2"

	_, err := srv.Spreadsheets.Values.BatchUpdate(spreadsheetId, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             updateVr,
	}).Do()

	if err != nil {
		slog.Error("Error updating Google Sheets", slog.String("error", err.Error()))
		return err
	}

	// Clear all old values
	clearRange := "Expertise!A2:C"
	clearValues := &sheets.ClearValuesRequest{}
	_, err = srv.Spreadsheets.Values.Clear(spreadsheetId, clearRange, clearValues).Do()
	if err != nil {
		slog.Error("Error clearing old values", slog.String("error", err.Error()))
		return err
	}

	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", slog.String("error", err.Error()))
		return err
	}

	return nil
}

const spreadsheetId = "1dStGuMfFU2Vq2V2xgXLyKUq_j3zYBeP15LA0eUQtTAQ"

func (e *External) NewSheetsClient() (*sheets.Service, error) {
	slog.Info("Updating Google Sheets")
	b, err := os.ReadFile("../secrets/credentials.json")
	if err != nil {
		slog.Error("Unable to read client secret file", slog.String("error", err.Error()))
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		slog.Error("Unable to parse client secret file to config", slog.String("error", err.Error()))
		return nil, err
	}
	client := GetClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		slog.Error("Unable to retrieve Sheets Client", slog.String("error", err.Error()))
		return nil, err
	}

	return srv, nil
}

type UpdateRequest struct {
	RawID int
	Value interface{}
}

const (
	MaxRequestsPerMinute = 60
)

func findRowIndexByID(table *sheets.ValueRange, id string) (int, error) {
	// Определяем диапазон, который будем получать (весь лист)

	// Ищем строку с нужным значением
	for i, row := range table.Values {
		if len(row) > 0 && row[0] == id {
			// Возвращаем индекс строки (в Google Sheets строки индексируются с 1)
			fmt.Println("Found: ", i+1)
			return i + 1, nil
		}
	}

	fmt.Println("Not found")
	// Если значение не найдено, возвращаем -1
	return -1, nil
}

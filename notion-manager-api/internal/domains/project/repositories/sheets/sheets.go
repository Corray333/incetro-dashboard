package sheets

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	"github.com/spf13/viper"
	"google.golang.org/api/sheets/v4"
)

type ProjectSheetsRepository struct {
	client *gsheets.Client
}

func NewProjectSheetsRepository(client *gsheets.Client) *ProjectSheetsRepository {
	return &ProjectSheetsRepository{
		client: client,
	}
}

func (r *ProjectSheetsRepository) UpdateSheetsProjects(ctx context.Context, sheetID string, projects []project.Project) error {
	if len(projects) == 0 {
		return nil
	}

	// Use default sheet ID if not provided
	if sheetID == "" {
		sheetID = viper.GetString("sheets.id")
	}

	appendRange := viper.GetString("sheets.projects_sheet") + "!A2:"
	rowLen := len(entityToSheetsProject(&projects[0]))
	lastColLetter := string(rune('A' + rowLen - 1))
	appendRange += lastColLetter

	clearValues := &sheets.ClearValuesRequest{}
	_, err := r.client.Svc().Spreadsheets.Values.Clear(sheetID, appendRange, clearValues).Do()
	if err != nil {
		slog.Error("Error clearing old values", "error", err)
		return err
	}

	var vr sheets.ValueRange

	for _, project := range projects {
		vr.Values = append(vr.Values, entityToSheetsProject(&project))
	}

	_, err = r.client.Svc().Spreadsheets.Values.Append(sheetID, appendRange, &vr).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		slog.Error("Error updating Google Sheets", "error", err)
		return err
	}

	return nil
}

func entityToSheetsProject(project *project.Project) []interface{} {
	clientName := ""
	if project.Client != nil {
		clientName = project.Client.Name
	}

	return []interface{}{
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(project.ID.String(), "-", "")), project.Name),
		project.Status,
		project.Type,
		clientName,
		project.SheetsLink,
	}
}

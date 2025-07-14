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

	appendRange := viper.GetString("sheets.clients_sheet") + "!A3:"
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

	return nil
}

func entityToSheetsClient(client *client.Client) []interface{} {
	projectNamesStr := ""
	if len(client.Projects) > 0 {
		projectNames := make([]string, len(client.Projects))
		for i, project := range client.Projects {
			projectNames[i] = project.Name
		}
		projectNamesStr = strings.Join(projectNames, ", ")
	}

	return []interface{}{
		fmt.Sprintf(`=HYPERLINK("%s"; "%s")`, fmt.Sprintf("https://notion.so/%s", strings.ReplaceAll(client.ID.String(), "-", "")), client.Name),
		string(client.Status),
		client.Source,
		client.UniqueID,
		projectNamesStr,
	}
}

package service

import (
	"context"
	"log/slog"
	"time"
	pkg_time "time"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	"github.com/Corray333/employee_dashboard/internal/utils"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type timeLastUpdateTimeGetter interface {
	GetTimeLastUpdateTime(ctx context.Context) (time.Time, error)
}

type timeLastUpdateTimeSetter interface {
	SetTimeLastUpdateTime(ctx context.Context, lastUpdateTime time.Time) error
}

type timeRawLister interface {
	ListTimes(ctx context.Context, lastUpdate pkg_time.Time) ([]entity_time.Time, error)
}

type timeSetter interface {
	SetTime(ctx context.Context, time *entity_time.Time) error
}

type timesLister interface {
	ListTimes(ctx context.Context, filter entity_time.TimeFilter, offset, limit int) ([]entity_time.Time, error)
}

func (s *TimeService) updateTimes(ctx context.Context) error {
	slog.Info("Updating times")
	lastUpdateTime, err := s.timeLastUpdateTimeGetter.GetTimeLastUpdateTime(ctx)
	if err != nil {
		return err
	}

	times, err := s.timeRawLister.ListTimes(ctx, lastUpdateTime)
	if err != nil {
		return err
	}
	if len(times) == 0 {
		return nil
	}

	lastTime := time.Time{}
	for _, time := range times {
		if err := s.timeSetter.SetTime(ctx, &time); err != nil {
			return err
		}
		if time.LastUpdate.After(lastTime) {
			lastTime = time.LastUpdate
		}
	}
	if err := s.timeLastUpdateTimeSetter.SetTimeLastUpdateTime(ctx, lastTime); err != nil {
		return err
	}

	return nil
}

func (s *TimeService) TimeSync(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		if err := s.updateTimes(ctx); err != nil {
			slog.Error("Notion time sync error", "error", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

type projectsLister interface {
	ListProjects(ctx context.Context) ([]project.Project, error)
}

func (s *TimeService) updateSheets(ctx context.Context) {

	times, err := s.timesLister.ListTimes(ctx, entity_time.TimeFilter{}, 0, 20000)
	if err != nil {
		slog.Error("Error getting times", "error", err)
		return
	}

	if err := s.sheetsRepository.UpdateSheetsTimes(ctx, viper.GetString("sheets.id"), times); err != nil {
		slog.Error("Error updating sheets", "error", err)
		return
	}

	// projects, err := s.projectsLister.ListProjects(ctx)
	// if err != nil {
	// 	slog.Error("Error getting projects", "error", err)
	// 	return
	// }

	// for _, project := range projects {
	// 	if project.SheetsLink == "" {
	// 		continue
	// 	}

	// 	sheetsID, err := utils.ExtractSpreadsheetID(project.SheetsLink)
	// 	if err != nil {
	// 		slog.Error("Error extracting spreadsheet ID", "error", err)
	// 		continue
	// 	}
	// 	times, err := s.timesLister.ListTimes(ctx, entity_time.TimeFilter{ProjectID: project.ID}, 0, 20000)
	// 	if err != nil {
	// 		slog.Error("Error getting times", "error", err)
	// 		return
	// 	}
	// 	if err := s.sheetsRepository.UpdateSheetsTimes(ctx, sheetsID, times); err != nil {
	// 		slog.Error("Error updating sheets", "error", err)
	// 		return
	// 	}
	// }

}

func (s *TimeService) UpdateProjectSheets(ctx context.Context, projectID uuid.UUID) error {
	times, err := s.timesLister.ListTimes(ctx, entity_time.TimeFilter{ProjectID: projectID}, 0, 20000)
	if err != nil {
		return err
	}

	projects, err := s.projectsLister.ListProjects(ctx)
	if err != nil {
		return err
	}

	var sheetID string
	for _, project := range projects {
		if project.ID == projectID {
			sheetID, err = utils.ExtractSpreadsheetID(project.SheetsLink)
			if err != nil {
				return err
			}
			break
		}
	}

	if sheetID == "" {
		return nil
	}

	if err := s.sheetsRepository.UpdateSheetsTimes(ctx, sheetID, times); err != nil {
		return err
	}

	return nil

}

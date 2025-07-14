package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/client/entities/client"
)

type clientsNotionLister interface {
	ListClients(ctx context.Context, lastUpdate time.Time) ([]client.Client, error)
}

type clientLastUpdateTimeGetter interface {
	GetClientsLastUpdateTime(ctx context.Context) (time.Time, error)
}

type clientSetter interface {
	SetClient(ctx context.Context, client *client.Client) error
}

type clientLister interface {
	ListClients(ctx context.Context, filter *client.Filter) ([]client.Client, error)
}

type sheetsClientsUpdater interface {
	UpdateSheetsClients(ctx context.Context, sheetID string, clients []client.Client) error
}

func (s *ClientService) ClientsSync(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		if err := s.updateClients(ctx); err != nil {
			slog.Error("Notion clients sync error", "error", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (s *ClientService) updateClients(ctx context.Context) error {
	lastUpdateTime, err := s.clientLastUpdateTimeGetter.GetClientsLastUpdateTime(ctx)
	if err != nil {
		return err
	}

	clients, err := s.clientsNotionLister.ListClients(ctx, lastUpdateTime)
	if err != nil {
		return err
	}
	if len(clients) == 0 {
		return nil
	}

	ctx, err = s.transactioner.Begin(ctx)
	if err != nil {
		return err
	}
	defer s.transactioner.Rollback(ctx)

	for _, c := range clients {
		if err := s.clientSetter.SetClient(ctx, &c); err != nil {
			return err
		}
	}
	if err := s.transactioner.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *ClientService) UpdateSheets(ctx context.Context) error {
	clients, err := s.clientLister.ListClients(ctx, &client.Filter{})
	if err != nil {
		slog.Error("Error getting clients", "error", err)
		return err
	}

	if s.sheetsClientsUpdater != nil {
		if err := s.sheetsClientsUpdater.UpdateSheetsClients(ctx, "", clients); err != nil {
			slog.Error("Error updating sheets", "error", err)
			return err
		}
	}

	return nil
}

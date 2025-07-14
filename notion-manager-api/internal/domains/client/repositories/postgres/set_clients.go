package postgres

import (
	"context"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/client/entities/client"
	"github.com/google/uuid"
)

type clientDB struct {
	ID         uuid.UUID `db:"client_id"`
	Name       string    `db:"name"`
	Status     string    `db:"status"`
	Source     string    `db:"source"`
	UniqueID   string    `db:"unique_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	ProjectIDs []string  `db:"project_ids"`
}

func clientDBFromEntity(c *client.Client) *clientDB {
	projectIDs := make([]string, len(c.ProjectIDs))
	for i, id := range c.ProjectIDs {
		projectIDs[i] = id.String()
	}

	return &clientDB{
		ID:         c.ID,
		Name:       c.Name,
		Status:     string(c.Status),
		Source:     c.Source,
		UniqueID:   c.UniqueID,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
		ProjectIDs: projectIDs,
	}
}

func (r *ClientPostgresRepository) SetClient(ctx context.Context, client *client.Client) error {
	clientDB := clientDBFromEntity(client)

	query := `
		INSERT INTO clients (client_id, name, status, source, unique_id, created_at, updated_at, project_ids)
		VALUES (:client_id, :name, :status, :source, :unique_id, :created_at, :updated_at, :project_ids)
		ON CONFLICT (client_id) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			source = EXCLUDED.source,
			unique_id = EXCLUDED.unique_id,
			updated_at = EXCLUDED.updated_at,
			project_ids = EXCLUDED.project_ids
	`

	_, err := r.DB().NamedExecContext(ctx, query, clientDB)
	if err != nil {
		return err
	}

	// Update last sync time
	updateQuery := `UPDATE system SET clients_db_last_sync = $1`
	_, err = r.DB().ExecContext(ctx, updateQuery, client.UpdatedAt)
	return err
}

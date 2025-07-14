package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/Corray333/employee_dashboard/internal/domains/client/entities/client"
	"github.com/google/uuid"
)

func (c *clientDB) ToEntity() *client.Client {
	projectIDs := make([]uuid.UUID, 0, len(c.ProjectIDs))
	for _, idStr := range c.ProjectIDs {
		if id, err := uuid.Parse(idStr); err == nil {
			projectIDs = append(projectIDs, id)
		}
	}

	return &client.Client{
		ID:         c.ID,
		Name:       c.Name,
		Status:     client.Status(c.Status),
		Source:     c.Source,
		UniqueID:   c.UniqueID,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
		ProjectIDs: projectIDs,
	}
}

func (r *ClientPostgresRepository) ListClients(ctx context.Context, filter *client.Filter) ([]client.Client, error) {
	query := `SELECT client_id, name, status, source, unique_id, created_at, updated_at, project_ids FROM clients`
	args := []interface{}{}
	conditions := []string{}
	argIndex := 1

	if filter != nil {
		if len(filter.IDs) > 0 {
			placeholders := make([]string, len(filter.IDs))
			for i, id := range filter.IDs {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, id)
				argIndex++
			}
			conditions = append(conditions, "client_id IN ("+strings.Join(placeholders, ",")+")")
		}

		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, string(*filter.Status))
			argIndex++
		}

		if filter.Source != nil {
			conditions = append(conditions, fmt.Sprintf("source = $%d", argIndex))
			args = append(args, *filter.Source)
			argIndex++
		}

		if filter.Name != nil {
			conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
			args = append(args, "%"+*filter.Name+"%")
			argIndex++
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY client_id DESC"

	var clientsDB []clientDB
	err := r.DB().SelectContext(ctx, &clientsDB, query, args...)
	if err != nil {
		return nil, err
	}

	clients := make([]client.Client, len(clientsDB))
	for i, c := range clientsDB {
		clients[i] = *c.ToEntity()
	}

	return clients, nil
}

func (r *ClientPostgresRepository) GetClientsByIDs(ctx context.Context, clientIDs []uuid.UUID) ([]client.Client, error) {
	if len(clientIDs) == 0 {
		return []client.Client{}, nil
	}

	query := `SELECT client_id, name, status, source, unique_id, created_at, updated_at, project_ids FROM clients WHERE client_id IN (`
	placeholders := make([]string, len(clientIDs))
	args := make([]interface{}, len(clientIDs))

	for i, id := range clientIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query += strings.Join(placeholders, ",") + ")"

	var clientsDB []clientDB
	err := r.DB().SelectContext(ctx, &clientsDB, query, args...)
	if err != nil {
		return nil, err
	}

	clients := make([]client.Client, len(clientsDB))
	for i, c := range clientsDB {
		clients[i] = *c.ToEntity()
	}

	return clients, nil
}

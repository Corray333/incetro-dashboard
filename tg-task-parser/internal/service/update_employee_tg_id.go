package service

import (
	"context"
)

type employeeTgIDUpdater interface {
	UpdateEmployeeTgID(ctx context.Context, username string, tgID int64) error
}

func (s *Service) UpdateEmployeeTgID(ctx context.Context, username string, tgID int64) error {
	return s.employeeTgIDUpdater.UpdateEmployeeTgID(ctx, username, tgID)
}

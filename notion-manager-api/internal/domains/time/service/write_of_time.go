package service

import (
	"context"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
)

type timeWriteOfCreater interface {
	CreateTimeWriteOf(ctx context.Context, time *entity_time.TimeOutboxMsg) error
}

func (s *TimeService) CreateTimeWriteOf(ctx context.Context, time *entity_time.TimeOutboxMsg) error {
	return s.timeWriteOfCreater.CreateTimeWriteOf(ctx, time)
}

package cron

import (
	"context"
	"log/slog"

	"github.com/robfig/cron/v3"
)

type service interface {
	SendIncorrectTimeNotifications(ctx context.Context) error
}

type CronService struct {
	service service
	cron    *cron.Cron
}

func NewCronService(service service) *CronService {
	return &CronService{
		service: service,
		cron:    cron.New(),
	}
}

func (c *CronService) Start() error {
	// Добавляем задачу на каждый день в 9:30 утра
	_, err := c.cron.AddFunc("30 6 * * *", func() {
		ctx := context.Background()
		if err := c.service.SendIncorrectTimeNotifications(ctx); err != nil {
			slog.Error("Failed to send incorrect time notifications", "error", err)
		}
	})
	if err != nil {
		return err
	}

	c.cron.Start()
	slog.Info("Cron service started")
	return nil
}

func (c *CronService) Stop() {
	c.cron.Stop()
	slog.Info("Cron service stopped")
}

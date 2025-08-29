package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/corray333/tg-task-parser/internal/entities/task"
	"github.com/corray333/tg-task-parser/internal/repositories/yatracker"
	"github.com/google/uuid"
)

type taskCreator interface {
	CreateTask(ctx context.Context, task *task.Task, projectID uuid.UUID) (string, error)
}

type yaTrackerTaskCreator interface {
	CreateTask(ctx context.Context, task *task.Task) (*yatracker.Issue, error)
}

type yaTrackerTaskSearcher interface {
	SearchTasksByName(ctx context.Context, t *task.Task) ([]yatracker.Issue, error)
}

type projectByChatIDGetter interface {
	GetProjectByChatID(ctx context.Context, chatID int64) (uuid.UUID, error)
}

// escapeMarkdownV2 escapes special characters for Telegram MarkdownV2
func escapeMarkdownV2(text string) string {
	// Characters that need to be escaped in MarkdownV2: _*[]()~`>#+-=|{}.!
	specialChars := regexp.MustCompile(`([_*\[\]()~` + "`" + `>#+=|{}.!-])`)
	return specialChars.ReplaceAllString(text, "\\$1")
}

func (s *Service) CreateTask(ctx context.Context, chatID int64, msg string, replyMessage string) (string, error) {
	projectID, err := s.projectByChatIDGetter.GetProjectByChatID(ctx, chatID)
	if err != nil {
		return "", err
	}

	newTask, err := s.taskMsgParser.ParseMessage(ctx, msg)
	if err != nil {
		return "", err
	}
	if newTask == nil {
		return "", nil
	}

	trackerProjectID, err := uuid.Parse("183dd045-c92a-42c3-83ba-5030fbb3451f")
	if err != nil {
		return "", err
	}

	// We'll capture a YaTracker issue (existing or newly created) to build a proper response later if needed
	var trackerIssue *yatracker.Issue

	if projectID == trackerProjectID {
		// try to find task in ya tracker
		tasks, err := s.yaTrackerTaskSearcher.SearchTasksByName(ctx, newTask)
		if err != nil {
			return "", err
		}
		if len(tasks) == 0 {
			// create task in ya tracker
			created, err := s.yaTrackerTaskCreator.CreateTask(ctx, newTask)
			if err != nil {
				return "", err
			}
			trackerIssue = created
		} else {
			// use the first matching issue
			issue := tasks[0]
			trackerIssue = &issue
		}
	}

	pageID, err := s.taskCreator.CreateTask(ctx, newTask, projectID)
	if err != nil {
		slog.Error("error while creating task in repository", "error", err)
		return "", err
	}

	notionLink := "https://notion.so/" + strings.ReplaceAll(pageID, "-", "")
	notionHyperlink := fmt.Sprintf("[%s](%s)", escapeMarkdownV2(newTask.Title), notionLink)
	quote := fmt.Sprintf("*Тело задачи:*\n\n```\n%s\n```", newTask.PlainBody)

	// Build response text
	if projectID == trackerProjectID && trackerIssue != nil {
		yaLink := fmt.Sprintf("https://tracker.yandex.ru/%s", trackerIssue.Key)
		text := fmt.Sprintf("Задача *%s: %s* создана:\n\n• Яндекс\\.Трекер: [%s](%s)\n• Notion: %s\n\n%s", escapeMarkdownV2(trackerIssue.Key), escapeMarkdownV2(trackerIssue.Summary), escapeMarkdownV2(trackerIssue.Key), yaLink, notionHyperlink, quote)
		return text, nil
	}

	// General case: only Notion link
	return fmt.Sprintf("Задача создана: %s\n\n%s", notionHyperlink, quote), nil
}

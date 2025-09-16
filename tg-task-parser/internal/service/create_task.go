package service

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"github.com/corray333/tg-task-parser/internal/entities/task"
	"github.com/corray333/tg-task-parser/internal/repositories/temp_storage"
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
	return s.CreateTaskFromCombinedMessage(ctx, chatID, msg, replyMessage, nil)
}

func (s *Service) CreateTaskFromCombinedMessage(ctx context.Context, chatID int64, msg string, replyMessage string, images []temp_storage.ImageInfo) (string, error) {
	projectID, err := s.projectByChatIDGetter.GetProjectByChatID(ctx, chatID)
	if err != nil {
		return "", err
	}

	msg += "\n\n" + replyMessage

	// Создаем каналы для параллельного выполнения
	type taskResult struct {
		task *task.Task
		err  error
	}

	type imagesResult struct {
		imageURLs []string
		err       error
	}

	taskChan := make(chan taskResult, 1)
	imagesChan := make(chan imagesResult, 1)

	// Параллельно парсим сообщение
	go func() {
		newTask, err := s.taskMsgParser.ParseMessage(ctx, msg)
		taskChan <- taskResult{task: newTask, err: err}
	}()

	// Параллельно обрабатываем изображения
	go func() {
		if len(images) == 0 {
			imagesChan <- imagesResult{imageURLs: nil, err: nil}
			return
		}

		imageURLs, err := s.processImages(ctx, images)
		imagesChan <- imagesResult{imageURLs: imageURLs, err: err}
	}()

	// Ждем результаты
	taskRes := <-taskChan
	if taskRes.err != nil {
		return "", taskRes.err
	}
	if taskRes.task == nil {
		return "", nil
	}

	imageRes := <-imagesChan
	if imageRes.err != nil {
		slog.Error("Failed to process images", "error", imageRes.err)
		// Продолжаем создание задачи без изображений
	} else {
		// Добавляем ссылки на изображения в задачу
		taskRes.task.Images = imageRes.imageURLs
	}

	trackerProjectID, err := uuid.Parse("183dd045-c92a-42c3-83ba-5030fbb3451f")
	if err != nil {
		return "", err
	}

	// We'll capture a YaTracker issue (existing or newly created) to build a proper response later if needed
	var trackerIssue *yatracker.Issue

	if projectID == trackerProjectID {
		// try to find task in ya tracker
		tasks, err := s.yaTrackerTaskSearcher.SearchTasksByName(ctx, taskRes.task)
		if err != nil {
			return "", err
		}
		if len(tasks) == 0 {
			// create task in ya tracker
			created, err := s.yaTrackerTaskCreator.CreateTask(ctx, taskRes.task)
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

	pageID, err := s.taskCreator.CreateTask(ctx, taskRes.task, projectID)
	if err != nil {
		slog.Error("error while creating task in repository", "error", err)
		return "", err
	}

	notionLink := "https://notion.so/" + strings.ReplaceAll(pageID, "-", "")
	notionHyperlink := fmt.Sprintf("[%s](%s)", escapeMarkdownV2(taskRes.task.Title), notionLink)
	// quote := fmt.Sprintf(">*Тело задачи:*\n>%s", escapeMarkdownV2(taskRes.task.PlainBody))

	// Build response text
	if projectID == trackerProjectID && trackerIssue != nil {
		yaLink := fmt.Sprintf("https://tracker.yandex.ru/%s", trackerIssue.Key)
		text := fmt.Sprintf("Задача создана: \n\n• Яндекс\\.Трекер: [%s](%s)\n• Notion: %s", escapeMarkdownV2(trackerIssue.Key), yaLink, notionHyperlink)
		return text, nil
	}

	// General case: only Notion link
	return fmt.Sprintf("Задача создана: \n\n%s", notionHyperlink), nil
}

// processImages скачивает и загружает изображения параллельно, возвращает ссылки на них
func (s *Service) processImages(ctx context.Context, images []temp_storage.ImageInfo) ([]string, error) {
	if len(images) == 0 {
		return nil, nil
	}

	var wg sync.WaitGroup
	imageURLs := make([]string, len(images))
	errors := make([]error, len(images))

	for i, img := range images {
		wg.Add(1)
		go func(idx int, imageInfo temp_storage.ImageInfo) {
			defer wg.Done()

			// Скачиваем изображение из Telegram
			imageData, err := s.tgFileDownloader.DownloadFile(ctx, imageInfo.FileID)
			if err != nil {
				slog.Error("Failed to download image from Telegram", "fileID", imageInfo.FileID, "error", err)
				errors[idx] = err
				return
			}

			// Генерируем уникальное имя для файла
			fileName := fmt.Sprintf("task_image_%s_%d.jpg", imageInfo.FileUniqueID, idx)

			// Сохраняем в файловое хранилище
			err = s.fileManager.SaveFile(ctx, imageData, fileName)
			if err != nil {
				slog.Error("Failed to save image to file storage", "fileName", fileName, "error", err)
				errors[idx] = err
				return
			}

			// Получаем URL для доступа к файлу
			url, err := s.fileManager.GetFileURL(ctx, fileName)
			if err != nil {
				slog.Error("Failed to get file URL", "fileName", fileName, "error", err)
				errors[idx] = err
				return
			}

			imageURLs[idx] = url
		}(i, img)
	}

	wg.Wait()

	// Собираем только успешно обработанные URL
	var successfulURLs []string
	for i, url := range imageURLs {
		if errors[i] == nil && url != "" {
			successfulURLs = append(successfulURLs, url)
		}
	}

	// Возвращаем ошибку только если ни одно изображение не удалось обработать
	if len(successfulURLs) == 0 && len(images) > 0 {
		return nil, fmt.Errorf("failed to process any images")
	}

	return successfulURLs, nil
}

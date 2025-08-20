package message

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/corray333/tg-task-parser/internal/entities/task"
	"github.com/corray333/tg-task-parser/internal/errs"
)

type Hashtag string

const (
	HashtagTask     = Hashtag("задача")
	HashtagFeedback = Hashtag("ос")
)

type Mention string

type Message struct {
	ChatID    int64  `json:"chat_id"`
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	FromID    int64  `json:"from_id"`

	Hashtags []Hashtag `json:"hashtags"`
	Mentions []Mention `json:"mentions"`

	Raw string `json:"raw"`
}

func ParseMessage(mainText, replyText string) (*Message, error) {
	task := &Message{}

	hashtagRegex, err := regexp.Compile(`#([\p{L}\d_]+)`)
	if err != nil {
		return nil, errors.Join(errs.ErrCompileHashingRegex, err)
	}

	mentionRegex, err := regexp.Compile(`@([\p{L}\d_]+)`)
	if err != nil {
		return nil, errors.Join(errs.ErrCompileMentionRegex, err)
	}

	fullText := mainText
	if replyText != "" {
		fullText = replyText + " " + mainText
	}

	// Найдем все хэштэги и упоминания без префиксов
	hashtags := hashtagRegex.FindAllStringSubmatch(mainText, -1)
	for _, hashtag := range hashtags {
		task.Hashtags = append(task.Hashtags, Hashtag(hashtag[1])) // Добавляем без "#"
	}

	mentions := mentionRegex.FindAllStringSubmatch(fullText, -1)
	for _, mention := range mentions {
		task.Mentions = append(task.Mentions, Mention(mention[1])) // Добавляем без "@"
	}

	// Обработка основного текста без хэштэгов и упоминаний
	taskText := replyText
	if taskText == "" {
		taskText = mainText
	}

	// Убираем хэштэги и упоминания из текста
	taskText = hashtagRegex.ReplaceAllString(taskText, "")
	taskText = mentionRegex.ReplaceAllString(taskText, "")
	taskText = strings.TrimSpace(taskText)

	// Приводим первую букву к верхнему регистру корректно для Unicode
	runes := []rune(taskText)
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
		taskText = string(runes)
	}

	task.Text = taskText

	task.Raw = fullText

	return task, nil
}

// ParseTask parses a message and returns a *task.Task if HashtagTask is present, otherwise returns nil
func ParseTask(mainText, replyText string) (*task.Task, error) {
	parsedMessage, err := ParseMessage(mainText, replyText)
	if err != nil {
		return nil, err
	}

	// Check if HashtagTask is present
	if !slices.Contains(parsedMessage.Hashtags, HashtagTask) {
		return nil, nil
	}

	// Convert message.Message to task.Task
	taskEntity := &task.Task{
		Text: parsedMessage.Text,
		Hashtags: func() []task.Tag {
			res := make([]task.Tag, 0, len(parsedMessage.Hashtags))
			for _, h := range parsedMessage.Hashtags {
				res = append(res, task.Tag(h))
			}
			return res
		}(),
		Mentions: func() []task.Mention {
			res := make([]task.Mention, 0, len(parsedMessage.Mentions))
			for _, m := range parsedMessage.Mentions {
				res = append(res, task.Mention(m))
			}
			return res
		}(),
	}

	return taskEntity, nil
}

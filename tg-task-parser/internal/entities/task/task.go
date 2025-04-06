package task

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/corray333/tg-task-parser/internal/errs"
)

type Hashtag string

type Mention string

type Task struct {
	Text     string    `json:"text"`
	Hashtags []Hashtag `json:"hashtags"`
	Mentions []Mention `json:"mentions"`
}

func TaskFromMessage(mainText, replyText string) (*Task, error) {
	task := &Task{}

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

	hashtags := hashtagRegex.FindAllString(fullText, -1)
	fmt.Println(fullText, hashtags)
	for _, hashtag := range hashtags {
		task.Hashtags = append(task.Hashtags, Hashtag(hashtag))
	}

	mentions := mentionRegex.FindAllString(fullText, -1)
	for _, mention := range mentions {
		task.Mentions = append(task.Mentions, Mention(mention))
	}

	taskText := replyText
	if taskText == "" {
		taskText = mainText
	}

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

	fmt.Printf("|%s|", task.Text)

	return task, nil
}

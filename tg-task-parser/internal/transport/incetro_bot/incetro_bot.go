package incetro_bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"github.com/corray333/tg-task-parser/internal/entities/feedback"
	message "github.com/corray333/tg-task-parser/internal/entities/message"
	"github.com/corray333/tg-task-parser/internal/entities/project"
	"github.com/corray333/tg-task-parser/internal/entities/topic"
	"github.com/google/uuid"
)

type IncetroTelegramBot struct {
	service    service
	bot        *gotgbot.Bot
	dispatcher *ext.Dispatcher
	updater    *ext.Updater
}

const (
	MsgTextChooseProject  = "Выберите проект:"
	MsgTextChooseFeedback = "Выберите обратную связь:"
)

const (
	CallbackTypeChooseProject  = "0"
	CallbackTypeChooseFeedback = "1"
)

type service interface {
	CreateTask(ctx context.Context, chatID int64, message string, replyMessage string) (string, error)
	GetProjects(ctx context.Context) ([]project.Project, error)
	LinkChatToProject(ctx context.Context, chatID int64, projectID uuid.UUID) error

	RequestActiveFeedbacks(ctx context.Context, chatID int64, messageID int64, msg *message.Message) ([]feedback.Feedback, error)
	AnswerFeedback(ctx context.Context, chatID, messageID int64, feedbackID uuid.UUID) error
	CreateFeedback(ctx context.Context, chatID, messageID int64) (uuid.UUID, error)
	SaveTgMessage(ctx context.Context, msg message.Message) error

	GetTopics(ctx context.Context) ([]topic.Topic, error)
}

func NewIncetroBot(service service) *IncetroTelegramBot {
	token := os.Getenv("BOT_TOKEN")
	bot, err := gotgbot.NewBot(token, nil)
	if err != nil {
		panic("failed to create bot: " + err.Error())
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(bot *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			slog.Error("Handler error", "error", err)
			return ext.DispatcherActionNoop
		},
	})

	stickers, _ := bot.GetForumTopicIconStickers(nil)
	for _, s := range stickers {
		fmt.Println(s.CustomEmojiId, s.Emoji)
	}

	updater := ext.NewUpdater(dispatcher, nil)

	tr := &IncetroTelegramBot{
		service:    service,
		bot:        bot,
		dispatcher: dispatcher,
		updater:    updater,
	}

	tr.registerHandlers()

	return tr
}

func (t *IncetroTelegramBot) registerHandlers() {

	t.dispatcher.AddHandler(handlers.NewCommand("setup", func(b *gotgbot.Bot, ctx *ext.Context) error {
		chatID := ctx.EffectiveChat.Id
		var err error

		defer func() {
			if err != nil {
				_, err := b.SendMessage(chatID, "Произошла ошибка при создании топиков. Пожалуйста, попробуйте еще раз.", &gotgbot.SendMessageOpts{
					MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
				})
				if err != nil {
					slog.Error("Error sending error message", "error", err)
				}
			}
		}()

		topics, err := t.service.GetTopics(context.Background())
		if err != nil {
			slog.Error("Get topics failed", "err", err)
			return err
		}

		for _, tp := range topics {
			_, err := b.CreateForumTopic(chatID, tp.Name, &gotgbot.CreateForumTopicOpts{
				IconCustomEmojiId: tp.Icon,
			})
			if err != nil && !strings.Contains(err.Error(), "MESSAGE_THREAD_ALREADY_EXISTS") {
				slog.Error("create topic failed", "topic", tp.Name, "err", err)
				return err
			}

		}

		_, err = b.SendMessage(chatID, "Топики успешно созданы.", nil)
		if err != nil {
			slog.Error("Error sending success message", "error", err)
		}
		err = nil

		return err
	}))

	t.dispatcher.AddHandler(handlers.NewCommand("pinapp", func(b *gotgbot.Bot, ctx *ext.Context) error {
		chatId := ctx.EffectiveChat.Id

		// Извлекаем текст после команды
		args := strings.TrimSpace(ctx.Message.Text[len("/pinapp"):])
		if args == "" {
			args = "Запусти наше мини-приложение!"
		}

		// Создаём inline-кнопку
		inlineKeyboard := [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text: "Открыть",
					Url:  "https://t.me/incetro_bot/management",
				},
			},
		}

		// Получаем информацию о чате
		chat, err := b.GetChat(chatId, nil)
		if err != nil {
			slog.Error("Не удалось получить информацию о чате", "error", err)
			return err
		}

		// Проверяем наличие закреплённого сообщения
		if chat.PinnedMessage != nil && chat.PinnedMessage.From != nil && chat.PinnedMessage.From.Id == b.Id && chat.PinnedMessage.MessageThreadId == ctx.EffectiveMessage.MessageThreadId {
			// Пытаемся изменить текст существующего закреплённого сообщения
			_, _, err = b.EditMessageText(args, &gotgbot.EditMessageTextOpts{
				ChatId:    chatId,
				MessageId: chat.PinnedMessage.MessageId,
				ReplyMarkup: gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: inlineKeyboard,
				},
			})
			if err != nil {
				slog.Error("Не удалось изменить текст закреплённого сообщения", "error", err)
				return err
			}

		} else {
			// Отправляем новое сообщение с кнопкой
			msg, err := b.SendMessage(chatId, args, &gotgbot.SendMessageOpts{
				ReplyMarkup:     gotgbot.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard},
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			})
			if err != nil {
				slog.Error("Не удалось отправить сообщение", "error", err)
				return err
			}

			// Закрепляем отправленное сообщение
			_, err = b.PinChatMessage(chatId, msg.MessageId, nil)
			if err != nil {
				slog.Error("Не удалось закрепить сообщение", "error", err)
				return err
			}
		}

		// Удалить сообщение с командой
		_, err = b.DeleteMessage(chatId, ctx.Message.MessageId, nil)
		if err != nil {
			slog.Error("Не удалось удалить сообщение с командой", "error", err)
			return err
		}

		return nil
	}))

	// Хендлер сообщений
	t.dispatcher.AddHandler(handlers.NewMessage(nil, func(bot *gotgbot.Bot, ctx *ext.Context) error {
		msg := ctx.EffectiveMessage
		if msg == nil {
			return nil
		}

		// save each message
		if err := t.service.SaveTgMessage(context.Background(), message.Message{
			ChatID:    msg.Chat.Id,
			MessageID: msg.MessageId,
			Text:      msg.Text,
			FromID:    msg.From.Id,
		}); err != nil {
			slog.Error("Error saving message", "error", err, "message", msg, "chat", msg.Chat, "user", msg.From)
		}

		if msg.NewChatMembers != nil {
			for _, member := range msg.NewChatMembers {
				if member.Id == bot.User.Id {
					return t.handleBotAddedToChat(bot, ctx)
				}
			}
			return nil
		}

		var reply string
		if msg.ReplyToMessage != nil {
			reply = msg.ReplyToMessage.Text
		}

		parsedMsg, err := message.ParseMessage(msg.Text, reply)
		if err != nil {
			slog.Error("Error parsing message", "error", err)
			_, _ = msg.Reply(bot, "Не удалось разобрать сообщение", nil)
			return nil
		}

		if slices.Contains(parsedMsg.Hashtags, message.HashtagTask) {
			pageID, err := t.service.CreateTask(context.Background(), msg.Chat.Id, msg.Text, reply)
			if err != nil {
				slog.Error("Error creating task", "error", err)
				_, _ = msg.Reply(bot, "Не удалось создать задачу", nil)
				return nil
			}

			if pageID == "" {
				return nil
			}

			_, err = msg.Reply(bot, fmt.Sprintf("Задача создана: https://notion.so/%s", strings.ReplaceAll(pageID, "-", "")), nil)
			if err != nil {
				slog.Error("Error sending confirmation", "error", err)
			}
		} else if slices.Contains(parsedMsg.Hashtags, message.HashtagFeedback) {
			feedbacks, err := t.service.RequestActiveFeedbacks(context.Background(), msg.Chat.Id, msg.MessageId, parsedMsg)
			if err != nil {
				slog.Error("Error listing feedbacks", "error", err)
				_, _ = msg.Reply(bot, "Не удалось получить список обратной связи", nil)
				return nil
			}

			var keyboard [][]gotgbot.InlineKeyboardButton

			for _, f := range feedbacks {
				btn := gotgbot.InlineKeyboardButton{
					Text:         f.Text,
					CallbackData: CallbackTypeChooseFeedback + "|" + f.ID.String() + "|" + strconv.Itoa(int(msg.GetMessageId())),
				}
				keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{btn})
			}

			keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
				gotgbot.InlineKeyboardButton{
					Text:         "Новая обратная связь",
					CallbackData: CallbackTypeChooseFeedback + "|" + uuid.Nil.String() + "|" + strconv.Itoa(int(msg.GetMessageId())),
				},
			})

			_, err = msg.Reply(bot, MsgTextChooseFeedback, &gotgbot.SendMessageOpts{
				ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: keyboard,
				},
			})
			if err != nil {
				slog.Error("Error sending feedbacks", "error", err)
			}
		}

		return nil
	}))

	// Хендлер callback-кнопок
	t.dispatcher.AddHandler(handlers.NewCallback(nil, func(bot *gotgbot.Bot, ctx *ext.Context) error {
		cb := ctx.CallbackQuery
		cbData := strings.Split(cb.Data, "|")
		if len(cbData) < 2 {
			slog.Error("Invalid callback data", "data", cb.Data)
			return nil
		}

		fmt.Println(cbData)

		switch cbData[0] {
		case CallbackTypeChooseProject:
			projectID, err := uuid.Parse(cbData[1])
			if err != nil {
				slog.Error("Invalid project UUID", "data", cb.Data)
				return nil
			}

			err = t.service.LinkChatToProject(context.Background(), cb.Message.GetChat().Id, projectID)
			if err != nil {
				slog.Error("Error linking chat to project", "error", err)
				return nil
			}
			_, _ = bot.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
				Text: "Проект выбран",
			})
		case CallbackTypeChooseFeedback:
			feedbackID, err := uuid.Parse(cbData[1])
			if err != nil {
				slog.Error("Invalid feedback UUID", "data", cb.Data)
				return nil
			}
			if len(cbData) != 3 {
				slog.Error("Invalid callback data", "data", cb.Data)
				return nil
			}
			messageID, err := strconv.Atoi(cbData[2])
			if err != nil {
				slog.Error("Invalid message ID", "data", cb.Data)
				return nil
			}

			if feedbackID == uuid.Nil {
				feedbackID, err := t.service.CreateFeedback(context.Background(), cb.Message.GetChat().Id, int64(messageID))
				if err != nil {
					slog.Error("Error creating feedback", "error", err)
					return nil
				}

				if _, err := bot.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
					Text: "Обратная связь создана",
				}); err != nil {
					slog.Error("Error answering callback query", "error", err)
					return nil
				}

				opts := &gotgbot.SendMessageOpts{
					MessageThreadId: ctx.Message.MessageThreadId,
				}

				_, err = bot.SendMessage(cb.Message.GetChat().Id,
					fmt.Sprintf("Новая обратная связь: https://notion.so/%s", strings.ReplaceAll(feedbackID.String(), "-", "")),
					opts)
				if err != nil {
					slog.Error("Error sending feedback link", "error", err)
				}

			} else {
				if err := t.service.AnswerFeedback(context.Background(), cb.Message.GetChat().Id, int64(messageID), feedbackID); err != nil {
					slog.Error("Error updating feedback", "error", err)
					return nil
				}

				if _, err := bot.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
					Text: "Обратная связь выбрана",
				}); err != nil {
					slog.Error("Error answering callback query", "error", err)
					return nil
				}
			}
		}

		_, _ = bot.DeleteMessage(cb.Message.GetChat().Id, cb.Message.GetMessageId(), nil)
		return nil
	}))
}

func (t *IncetroTelegramBot) Run() {
	slog.Info("Bot is running...")
	err := t.updater.StartPolling(t.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 10,
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	t.updater.Idle()
}

func (t *IncetroTelegramBot) handleBotAddedToChat(bot *gotgbot.Bot, ctx *ext.Context) error {
	projects, err := t.service.GetProjects(context.Background())
	if err != nil {
		slog.Error("Failed to get projects", "error", err)
		_, _ = ctx.EffectiveMessage.Reply(bot, "Ошибка при получении проектов", nil)
		return nil
	}

	var keyboard [][]gotgbot.InlineKeyboardButton
	var row []gotgbot.InlineKeyboardButton

	for _, p := range projects {
		btn := gotgbot.InlineKeyboardButton{
			Text:         p.Name,
			CallbackData: CallbackTypeChooseProject + "|" + p.ID.String(),
		}
		row = append(row, btn)
		if len(row) == 2 {
			keyboard = append(keyboard, row)
			row = nil
		}
	}
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}

	_, err = ctx.EffectiveMessage.Reply(bot, MsgTextChooseProject, &gotgbot.SendMessageOpts{
		ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	})

	if err != nil {
		slog.Error("Failed to send inline keyboard", "error", err)
	}
	return nil
}

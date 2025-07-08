package tgclient

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type MTProtoClient struct {
	Client *telegram.Client
	API    *tg.Client
	Ctx    context.Context

	cancel context.CancelFunc
}

func NewMTProtoClient(appID int, appHash string, phone string) *MTProtoClient {
	sessionFile := "../secrets/tg/session.dat"

	storage := &session.FileStorage{
		Path: sessionFile,
	}

	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: storage,
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	c := &MTProtoClient{
		Client: client,
		API:    client.API(),
		Ctx:    ctx,
		cancel: cancel,
	}

	go func() {
		err := client.Run(context.Background(), func(ctx context.Context) error {

			codePrompt := func(ctx context.Context, _ *tg.AuthSentCode) (string, error) {
				fmt.Print("Введите код из SMS/Telegram: ")
				code, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				return strings.TrimSpace(code), nil
			}

			flow := auth.NewFlow(
				auth.Constant(phone, "jyd1iq-sowkog-hesHi6", auth.CodeAuthenticatorFunc(codePrompt)),
				auth.SendCodeOptions{},
			)

			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return fmt.Errorf("auth: %w", err)
			}

			wg.Done()
			<-ctx.Done()
			return nil

		})
		if err != nil {
			slog.Error("Error running mtproto client", "error", err)
		}
	}()

	wg.Wait()
	return c
}

func (r *MTProtoClient) GetAccessHashByChannelID(ctx context.Context, channelID int64) (int64, error) {
	dialogs, err := r.API.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      20,
	})
	if err != nil {
		slog.Error("Error getting dialogs", "error", err)
		return 0, err
	}

	switch d := dialogs.(type) {
	case *tg.MessagesDialogs:
		for _, chat := range d.Chats {
			switch ch := chat.(type) {
			case *tg.Chat:
				switch mt := ch.MigratedTo.(type) {
				case *tg.InputChannel:
					if mt.ChannelID == channelID {
						return mt.AccessHash, nil
					}
				}
			}
		}
	}
	return 0, fmt.Errorf("channel not found")
}

func (r *MTProtoClient) GetGroupByTitle(ctx context.Context, title string) (int64, int64, error) {
	dialogs, err := r.API.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      20,
	})
	if err != nil {
		slog.Error("Error getting dialogs", "error", err)
		return 0, 0, err
	}

	switch d := dialogs.(type) {
	case *tg.MessagesDialogs:
		for _, chat := range d.Chats {
			switch ch := chat.(type) {
			case *tg.Chat:
				if ch.Title == title && ch.MigratedTo == nil {
					return ch.ID, 0, nil
				}
				switch mt := ch.MigratedTo.(type) {
				case *tg.InputChannel:
					return mt.ChannelID, mt.AccessHash, nil
				}
			}
		}
	}
	return 0, 0, fmt.Errorf("channel not found")
}

type Topic struct {
	ID    int
	Title string
	Icon  int64
}

func (r *MTProtoClient) CreateTopics(ctx context.Context, channelID int64, accessHash int64, topics []Topic) error {
	if accessHash == 0 {
		u, err := r.API.MessagesMigrateChat(ctx, channelID)
		if err != nil && !strings.Contains(err.Error(), tg.ErrChatIDInvalid) {
			slog.Error("Error migrating chat", "error", err)
			return err
		}

		channelID = int64(0)
		accessHash = int64(0)

		switch up := u.(type) {
		case *tg.Updates:
			for _, chatObj := range up.Chats {
				switch ch := chatObj.(type) {
				case *tg.Chat:
					if ch.MigratedTo != nil {
						if dst, ok := ch.MigratedTo.(*tg.InputChannel); ok {
							channelID = dst.ChannelID
							accessHash = dst.AccessHash
						}
					}
				}
			}
		default:
			log.Printf("unhandled concrete type %T", up)
			return nil
		}
	}

	for _, topic := range topics {
		_, err := r.API.ChannelsCreateForumTopic(context.Background(), &tg.ChannelsCreateForumTopicRequest{
			Channel: &tg.InputChannel{
				ChannelID:  channelID,
				AccessHash: accessHash,
			},
			Title:       topic.Title,
			IconEmojiID: topic.Icon,
			RandomID:    time.Now().Unix(),
		})
		if err != nil {
			slog.Error("Error creating topic", "error", err)
			return err
		}
	}

	return nil
}

func (r *MTProtoClient) GetPinnedMessageIDInTopic(ctx context.Context, channelID int64, accessHash int64, topicID int) (int, error) {
	msgs, err := r.API.MessagesSearch(context.Background(), &tg.MessagesSearchRequest{
		Peer: &tg.InputPeerChannel{
			ChannelID:  channelID,
			AccessHash: accessHash,
		},
		Filter: &tg.InputMessagesFilterPinned{},
		Limit:  100,
	})
	if err != nil {
		slog.Error("Error getting channel info", "error", err)
	}
	switch msgs := msgs.(type) {
	case *tg.MessagesChannelMessages:
		for _, msg := range msgs.Messages {
			switch msg := msg.(type) {
			case *tg.Message:
				switch reply := msg.ReplyTo.(type) {
				case *tg.MessageReplyHeader:
					if reply.ReplyToTopID != 0 {
						if reply.ReplyToTopID == topicID {
							return msg.ID, nil
						}
						continue
					}
					if reply.ReplyToMsgID == topicID {
						return msg.ID, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("pinned message not found")
}

var topics = []Topic{
	{
		Title: "Backend разработка",
		Icon:  5783078953308655968,
	},
	{
		Title: "Мобильная разработка",
		Icon:  5974453277155135447,
	},
	{
		Title: "Тестирование",
		Icon:  5976544483846654540,
	},
	{
		Title: "Сборки приложения",
		Icon:  5974053797951967293,
	},
	{
		Title: "Менеджерская",
		Icon:  6012499476147604376,
	},
	{
		Title: "Общий",
		Icon:  6001526766714227911,
	},
	{
		Title: "Дизайн",
		Icon:  5974572969303739894,
	},
	{
		Title: "От клиента",
		Icon:  5974416568069655298,
	},
	{
		Title: "Аналитика",
		Icon:  5976377521287990495,
	},
	{
		Title: "Web разработка",
		Icon:  5974475701179387553,
	},
	{
		Title: "Вопросы",
		Icon:  5974229895906069525,
	},
}

func (r *MTProtoClient) GetTopics(ctx context.Context, channelID int64, accessHash int64) ([]Topic, error) {
	info, err := r.API.ChannelsGetForumTopics(context.Background(), &tg.ChannelsGetForumTopicsRequest{
		Channel: &tg.InputChannel{
			ChannelID:  channelID,
			AccessHash: accessHash,
		},
	})
	if err != nil {
		slog.Error("Error getting forum topics", "error", err)
	}
	result := make([]Topic, 0, len(info.Topics))
	for _, topic := range info.Topics {
		switch topic := topic.(type) {
		case *tg.ForumTopic:
			result = append(result, Topic{
				ID:    topic.ID,
				Title: topic.Title,
				Icon:  topic.IconEmojiID,
			})
		}
	}
	return result, nil
}

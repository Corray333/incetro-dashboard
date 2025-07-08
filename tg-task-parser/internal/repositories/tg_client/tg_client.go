package tg_client

import (
	"os"
	"strconv"

	"github.com/corray333/tg-task-parser/pkg/tgclient"
)

type TgClient struct {
	client *tgclient.MTProtoClient
}

func NewTgClient() *TgClient {
	appID, err := strconv.Atoi(os.Getenv("TG_CLIENT_APP_ID"))
	if err != nil {
		panic(err)
	}
	client := tgclient.NewMTProtoClient(appID, os.Getenv("TG_CLIENT_APP_HASH"), os.Getenv("TG_CLIENT_PHONE"))
	return &TgClient{
		client: client,
	}
}

package tg_client

import (
	"os"
	"strconv"

	"github.com/corray333/tg-task-parser/pkg/tgclient"
)

type MtProtoRepository struct {
	client *tgclient.MTProtoClient
}

func NewTgClient() *MtProtoRepository {
	appID, err := strconv.Atoi(os.Getenv("TG_CLIENT_APP_ID"))
	if err != nil {
		panic(err)
	}
	client := tgclient.NewMTProtoClient(appID, os.Getenv("TG_CLIENT_APP_HASH"), os.Getenv("TG_CLIENT_PHONE"))
	return &MtProtoRepository{
		client: client,
	}
}

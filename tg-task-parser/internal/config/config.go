package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/corray333/tg-task-parser/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type YaTrackerConfig struct {
	QueueID   int    `mapstructure:"queue_id"`
	ProjectID int    `mapstructure:"project_id"`
	OrgID     string `mapstructure:"org_id"`
	APIURL    string `mapstructure:"api_url"`
}

func MustInit() {
	if err := godotenv.Load("../.env"); err != nil {
		panic("error while loading .env file: " + err.Error())
	}
	// Setup viper getting file in ./configs directory. File has name like local.yaml or prod.yaml, depends on env variable ENV
	viper.SetConfigName(os.Getenv("ENV"))
	viper.AddConfigPath("../configs")
	if err := viper.ReadInConfig(); err != nil {
		panic("error while reading config file: " + err.Error())
	}
	fmt.Println(viper.AllKeys())
	SetupLogger()
}

func SetupLogger() {
	handler := logger.NewHandler(nil)
	log := slog.New(handler)
	slog.SetDefault(log)
}

func GetYaTrackerConfig() YaTrackerConfig {
	var config YaTrackerConfig
	if err := viper.UnmarshalKey("yatracker", &config); err != nil {
		panic("error while unmarshaling yatracker config: " + err.Error())
	}
	return config
}

func GetYaTrackerToken() string {
	return os.Getenv("YA_TRACKER_TOKEN")
}

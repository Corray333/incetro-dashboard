package config

import (
	"log/slog"
	"os"

	"github.com/Corray333/employee_dashboard/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

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
	SetupLogger()
}

func SetupLogger() {
	handler := logger.NewHandler(nil)
	log := slog.New(handler)
	slog.SetDefault(log)
}

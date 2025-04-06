package main

import (
	"fmt"

	"github.com/corray333/tg-task-parser/internal/app"
	"github.com/corray333/tg-task-parser/internal/config"
)

func main() {
	config.MustInit()
	fmt.Println("Start")

	app.New().Run()
}

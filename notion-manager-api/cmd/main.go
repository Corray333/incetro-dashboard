package main

import (
	"fmt"
	"time"

	"github.com/Corray333/employee_dashboard/internal/app"
	"github.com/Corray333/employee_dashboard/internal/config"
)

func main() {
	fmt.Println((int(time.Now().Month())-1)/3 + 1)
	config.MustInit()
	fmt.Println("Start")

	app.New().Run()
}

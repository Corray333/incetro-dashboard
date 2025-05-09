package main

import (
	"fmt"

	"github.com/Corray333/employee_dashboard/internal/app"
	"github.com/Corray333/employee_dashboard/internal/config"
)

func main() {

	config.MustInit()
	fmt.Println("Start")

	app.New().Run()
}

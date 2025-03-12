package app

import (
	"fmt"

	"github.com/Corray333/employee_dashboard/internal/external"
	"github.com/Corray333/employee_dashboard/internal/repositories"
	"github.com/Corray333/employee_dashboard/internal/service"
	"github.com/Corray333/employee_dashboard/internal/transport"
)

type app struct {
	store     *repositories.Storage
	service   *service.Service
	transport *transport.Transport
}

func New() *app {

	storage := repositories.New()
	external := external.New()
	// data, _ := external.GetSheetsTimes(1741718524, "", "")
	// for i := range data {
	// 	fmt.Println(data[i].Properties.WhatDid.Title, data[i].Properties.BHGS.Formula.String)
	// }
	val, _ := storage.GetEmployees()
	fmt.Printf("%+v\n", val)
	service := service.New(storage, external)

	transport := transport.New(service)
	transport.RegisterRoutes()

	app := &app{
		store:     storage,
		service:   service,
		transport: transport,
	}

	return app
}

func (app *app) Run() {
	go app.service.Run()
	app.transport.Run()
}

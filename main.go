package main

import (
	"API/db"
	"API/server"
)

func main() {
	db.RunMigrations()
	dataBase := db.ConnectGORM()
	server.Init(dataBase)
	server.InitRoutes()
	server.Run()
}

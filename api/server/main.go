package main

import (
	"git.houseofkummer.com/lior/home-dns/api/services"
)

func main() {
	app := services.NewApp(services.Options{})
	app.Run()
}

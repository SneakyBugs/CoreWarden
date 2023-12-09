package main

import (
	"context"
	"fmt"

	"git.houseofkummer.com/lior/home-dns/api/services"
)

func main() {
	app := services.NewApp(services.Options{})
	fmt.Println("Hello world!")
	err := app.Start(context.Background())
	if err != nil {
		panic(err)
	}
}

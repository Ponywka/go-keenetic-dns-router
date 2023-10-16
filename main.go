package main

import (
	"github.com/Ponywka/go-keenetic-dns-router/internal/app"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	config := app.Config{}
	if err := env.Parse(&config); err != nil {
		panic(err)
	}

	if err := app.New(&config); err != nil {
		panic(err)
	}
}

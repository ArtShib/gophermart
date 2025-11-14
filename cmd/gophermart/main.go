package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ArtShib/gophermart.git/internal/app"
	"github.com/ArtShib/gophermart.git/internal/config"
	"github.com/ArtShib/gophermart.git/internal/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoadConfig()
	store, err := storage.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	newApp := app.NewApp(cfg, &store)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go newApp.Run(ctx)

	<-quit

	newApp.Stop(ctx)

}

//./gophermarttest -accrual-binary-path "/home/artem/GolandProjects/gophermart/cmd/accrual/accrual_linux_amd64" -accrual-database-uri "host=localhost port=5432 user=postgres password=mysecretpassword dbname=postgres sslmode=disable" -accrual-port "8081" -gophermart-binary-path "/home/artem/GolandProjects/gophermart/cmd/gophermart/gophermart" -gophermart-database-uri "host=localhost port=5432 user=postgres password=mysecretpassword dbname=postgres sslmode=disable" -gophermart-port "8080" -gophermart-host "localhost" -accrual-host "localhost"

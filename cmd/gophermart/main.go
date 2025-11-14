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

package logger

import (
	"log/slog"
	"os"
)

//package logger
//
//import (
//"log/slog"
//"os"
//)

func New() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	return logger
}

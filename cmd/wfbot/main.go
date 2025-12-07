package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"team-workflow-bot/internal/app"
	"team-workflow-bot/internal/config"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer log.Print("Team workflow bot application finished")

	cfg := config.LoadConfig()

	application := app.NewApp(cfg)

	application.Start(ctx).Wait()

	select {
	case <-time.After(5 * time.Second):
		log.Print("Forced shutdown after timeout")
	case <-ctx.Done():
	}
}

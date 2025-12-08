package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"team-workflow-bot/internal/app"
	"team-workflow-bot/internal/config"
	"team-workflow-bot/internal/environment"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer log.Print("Team workflow bot application finished")

	log.SetOutput(os.Stdout)

	environment.InitGlobal()
	cfg := config.Load("team-workflow-bot")

	application := app.NewApp(cfg)

	application.Start(ctx)

	select {
	case <-time.After(5 * time.Second):
		log.Print("Forced shutdown after timeout")
	case <-ctx.Done():
	}
}

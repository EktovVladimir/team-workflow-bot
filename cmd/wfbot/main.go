package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"team-workflow-bot/internal/app"
	"team-workflow-bot/internal/config"
	"team-workflow-bot/internal/global"
	"team-workflow-bot/internal/logger"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer logrus.Info("Team workflow bot application finished")

	log.SetOutput(os.Stdout)

	global.InitGlobal()
	cfg := config.Load("team-workflow-bot")

	logger.Init(cfg)

	application := app.NewApp(cfg)

	application.Start(ctx)

	select {
	case <-time.After(5 * time.Second):
		logrus.Info("Forced shutdown after timeout")
	case <-ctx.Done():
	}
}

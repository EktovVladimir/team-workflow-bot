package logger

import (
	"log"
	"os"
	"path/filepath"
	"team-workflow-bot/internal/config"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

func Init(cfg *config.Config) {
	fileFormatter := &logrus.TextFormatter{}

	fileWriter, _ := rotatelogs.New(
		filepath.Join(cfg.Logger.Dir, "%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(time.Duration(cfg.Logger.MaxAge)*24*time.Hour),
	)

	level, err := logrus.ParseLevel(cfg.Logger.Level)
	if err != nil {
		level = logrus.InfoLevel
	}

	logrus.SetLevel(level)

	logrus.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.DebugLevel: fileWriter,
			logrus.InfoLevel:  fileWriter,
			logrus.WarnLevel:  fileWriter,
			logrus.ErrorLevel: fileWriter,
			logrus.FatalLevel: fileWriter,
			logrus.PanicLevel: fileWriter,
		},
		fileFormatter,
	))

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	logrus.SetOutput(os.Stdout)
}

func NewStdLogger(level logrus.Level, prefix string) *log.Logger {
	return log.New(logrus.StandardLogger().WriterLevel(level), prefix, log.Lshortfile|log.LstdFlags)
}

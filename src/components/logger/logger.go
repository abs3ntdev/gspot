package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/lmittmann/tint"
	"go.uber.org/fx"

	"github.com/abs3ntdev/gspot/src/config"
)

type LoggerResult struct {
	fx.Out
	Logger *slog.Logger
}

type LoggerParams struct {
	fx.In

	Config *config.Config
}

func NewLogger(p LoggerParams) LoggerResult {
	lvl := slog.LevelInfo
	configLevel := strings.ToUpper(p.Config.LogLevel)
	switch configLevel {
	case "INFO":
		lvl = slog.LevelInfo
	case "WARN":
		lvl = slog.LevelWarn
	case "ERROR":
		lvl = slog.LevelError
	case "DEBUG":
		lvl = slog.LevelDebug
	}
	if strings.ToUpper(p.Config.LogOutput) == "FILE" {
		fp := ""
		p, err := os.UserConfigDir()
		if err != nil {
			p, err := os.UserHomeDir()
			if err != nil {
				os.Exit(1)
			}
			fp = filepath.Join(p, ".config", "gspot", "gspot.log")
		} else {
			fp = filepath.Join(p, "gspot", "gspot.log")
		}
		f, err := os.Create(fp)
		if err != nil {
			os.Exit(1)
		}
		return LoggerResult{
			Logger: slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
				Level: lvl.Level(),
			})),
		}
	}
	return LoggerResult{
		Logger: slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level:      lvl.Level(),
			TimeFormat: "[15:04:05.000]",
		})),
	}
}

package server

import (
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/yylego/kratos-cron/cronkratos"
	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
)

// NewCronServer creates a new cron server and registers cron jobs with stage locking
// 创建新的 cron server 并注册带 stage 读锁协调的定时任务
func NewCronServer(cronService *service.CronService, logger *slog.Logger) *cronkratos.Server {
	srv := cronkratos.NewServer(
		cron.New(
			cron.WithSeconds(),
			cron.WithLocation(time.FixedZone("CST", 8*60*60)), // UTC+8
		),
		logger,
	)
	cronService.RegisterCron(srv)
	return srv
}

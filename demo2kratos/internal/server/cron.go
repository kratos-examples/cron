package server

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
	"github.com/yylego/kratos-cron/cronkratos"
	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
)

// NewCronServer creates a new cron server and registers cron jobs with locker
// 创建新的 cron server 并注册带锁的定时任务
func NewCronServer(cronService *service.CronService, logger log.Logger) *cronkratos.Server {
	srv := cronkratos.NewServer(
		cron.New(
			cron.WithSeconds(),
			cron.WithLocation(time.FixedZone("CST", 8*60*60)), // UTC+8
		),
		logger,
	)
	cronkratos.RegisterCronServerL(srv, cronService)
	return srv
}

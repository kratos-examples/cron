package service

import (
	"context"
	"log/slog"

	"github.com/yylego/kratos-cron/cronkratos"
	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
	"github.com/yylego/rese"
)

// CronService handles cron job registration
// 处理定时任务注册
type CronService struct {
	task *biz.TaskUsecase
	slog *slog.Logger
}

// NewCronService creates a new CronService instance
// 创建新的 CronService 实例
func NewCronService(task *biz.TaskUsecase, logger *slog.Logger) *CronService {
	return &CronService{task: task, slog: logger}
}

// RegisterCron registers cron jobs on the cron server
// 注册定时任务到 cron server
func (s *CronService) RegisterCron(srv *cronkratos.Server) {
	// Sync data every minute
	// 每分钟同步数据
	rese.C1(srv.AddFunc("0 * * * * *", func(ctx context.Context, stage *cronkratos.Stage) {
		if erk := s.task.SyncData(ctx); erk != nil {
			s.slog.Error("sync data task error", "err", erk)
		} else {
			s.slog.Info("sync data task success")
		}
	}))

	// Cleanup data every second
	// 每秒清理数据
	rese.C1(srv.AddFunc("* * * * * *", func(ctx context.Context, stage *cronkratos.Stage) {
		if erk := s.task.CleanupData(ctx); erk != nil {
			s.slog.Error("cleanup data task error", "err", erk)
		} else {
			s.slog.Info("cleanup data task success")
		}
	}))
}

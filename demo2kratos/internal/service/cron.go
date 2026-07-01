package service

import (
	"context"
	"log/slog"

	"github.com/yylego/kratos-cron/cronkratos"
	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
	"github.com/yylego/rese"
)

// CronService handles cron job registration with stage locking
// 处理带 stage 读锁协调的定时任务注册
type CronService struct {
	task *biz.TaskUsecase
	slog *slog.Logger
}

// NewCronService creates a new CronService instance
// 创建新的 CronService 实例
func NewCronService(task *biz.TaskUsecase, logger *slog.Logger) *CronService {
	return &CronService{task: task, slog: logger}
}

// RegisterCron registers cron jobs; each task drives the stage read-lock for safe shutdown
// 注册定时任务，每个 task 用 stage 读锁配合安全退出
func (s *CronService) RegisterCron(srv *cronkratos.Server) {
	// Sync data every minute, lock each iteration via stage
	// 每分钟同步数据，在业务层循环中每次迭代用 stage 加锁
	rese.C1(srv.AddFunc("0 * * * * *", func(ctx context.Context, stage *cronkratos.Stage) {
		if erk := s.task.SyncData(ctx, stage); erk != nil {
			s.slog.Error("sync data task error", "err", erk)
		} else {
			s.slog.Info("sync data task success")
		}
	}))

	// Cleanup data every second, lock via stage
	// 每秒清理数据，用 stage 加锁
	rese.C1(srv.AddFunc("* * * * * *", func(ctx context.Context, stage *cronkratos.Stage) {
		if erk := s.task.CleanupData(ctx, stage); erk != nil {
			s.slog.Error("cleanup data task error", "err", erk)
		} else {
			s.slog.Info("cleanup data task success")
		}
	}))
}

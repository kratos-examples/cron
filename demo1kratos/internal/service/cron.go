package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
	"github.com/yylego/rese"
)

// CronService handles cron job registration
// 处理定时任务注册
type CronService struct {
	task *biz.TaskUsecase
	slog *log.Helper
}

// NewCronService creates a new CronService instance
// 创建新的 CronService 实例
func NewCronService(task *biz.TaskUsecase, logger log.Logger) *CronService {
	return &CronService{task: task, slog: log.NewHelper(logger)}
}

// RegisterCron registers cron jobs to cron instance
// 注册定时任务到 cron 实例
func (s *CronService) RegisterCron(ctx context.Context, c *cron.Cron) {
	// Sync data every minute
	// 每分钟同步数据
	rese.C1(c.AddFunc("0 * * * * *", func() {
		if erk := s.task.SyncData(ctx); erk != nil {
			s.slog.Errorf("sync data task error: %v", erk)
		} else {
			s.slog.Info("sync data task success")
		}
	}))

	// Cleanup data every second
	// 每秒清理数据
	rese.C1(c.AddFunc("* * * * * *", func() {
		if erk := s.task.CleanupData(ctx); erk != nil {
			s.slog.Errorf("cleanup data task error: %v", erk)
		} else {
			s.slog.Info("cleanup data task success")
		}
	}))
}

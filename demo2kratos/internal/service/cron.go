package service

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
	"github.com/yylego/rese"
)

// CronService handles cron job registration with locker
// 处理带锁的定时任务注册
type CronService struct {
	task *biz.TaskUsecase
	slog *log.Helper
}

// NewCronService creates a new CronService instance
// 创建新的 CronService 实例
func NewCronService(task *biz.TaskUsecase, logger log.Logger) *CronService {
	return &CronService{task: task, slog: log.NewHelper(logger)}
}

// RegisterCron registers cron jobs with locker for safe shutdown
// 注册定时任务，使用锁支持安全退出
func (s *CronService) RegisterCron(ctx context.Context, c *cron.Cron, locker sync.Locker) {
	// Sync data every minute, lock each iteration in biz loop
	// 每分钟同步数据，在业务层循环中每次迭代加锁
	rese.C1(c.AddFunc("0 * * * * *", func() {
		if erk := s.task.SyncData(ctx, locker); erk != nil {
			s.slog.Errorf("sync data task error: %v", erk)
		} else {
			s.slog.Info("sync data task success")
		}
	}))

	// Cleanup data every second, lock at biz layer
	// 每秒清理数据，在业务层加锁
	rese.C1(c.AddFunc("* * * * * *", func() {
		if erk := s.task.CleanupData(ctx, locker); erk != nil {
			s.slog.Errorf("cleanup data task error: %v", erk)
		} else {
			s.slog.Info("cleanup data task success")
		}
	}))
}

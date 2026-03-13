package biz

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
)

// TaskUsecase handles scheduled task business logic
// 处理定时任务的业务逻辑
type TaskUsecase struct {
	slog *log.Helper
}

// NewTaskUsecase creates a new TaskUsecase instance
// 创建新的 TaskUsecase 实例
func NewTaskUsecase(logger log.Logger) *TaskUsecase {
	return &TaskUsecase{
		slog: log.NewHelper(logger),
	}
}

// SyncData performs data sync in loop, lock each iteration
// 循环同步数据，每次迭代加锁
func (uc *TaskUsecase) SyncData(ctx context.Context, locker sync.Locker) *errors.Error {
	for i := 0; i < 10; i++ {
		if erk := uc.syncOnce(ctx, locker); erk != nil {
			return erk
		}
	}
	uc.slog.Info("SyncData complete")
	return nil
}

// syncOnce performs single data sync with lock protection
// 执行单次数据同步，带锁保护
func (uc *TaskUsecase) syncOnce(ctx context.Context, locker sync.Locker) *errors.Error {
	locker.Lock()
	defer locker.Unlock()
	// Check ctx validity after acquiring lock, exit if cancelled
	// 获取锁后检查 ctx 是否有效，已取消则退出
	if ctx.Err() != nil {
		return pb.ErrorUnknown("context error=%v", ctx.Err())
	}
	uc.slog.WithContext(ctx).Infof("syncOnce executed at %s", time.Now().Format(time.RFC3339))
	return nil
}

// CleanupData performs data cleanup with lock at biz layer
// 在业务层加锁执行数据清理
func (uc *TaskUsecase) CleanupData(ctx context.Context, locker sync.Locker) *errors.Error {
	locker.Lock()
	defer locker.Unlock()
	// Check ctx validity after acquiring lock, exit if cancelled
	// 获取锁后检查 ctx 是否有效，已取消则退出
	if ctx.Err() != nil {
		return pb.ErrorUnknown("context error=%v", ctx.Err())
	}
	uc.slog.WithContext(ctx).Infof("CleanupData executed at %s", time.Now().Format(time.RFC3339))
	return nil
}

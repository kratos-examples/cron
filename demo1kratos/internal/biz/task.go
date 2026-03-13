package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
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

// SyncData performs data sync in loop
// 循环同步数据
func (uc *TaskUsecase) SyncData(ctx context.Context) *errors.Error {
	for i := 0; i < 10; i++ {
		if erk := uc.syncOnce(ctx); erk != nil {
			return erk
		}
	}
	uc.slog.Info("SyncData complete")
	return nil
}

// syncOnce performs single data sync
// 执行单次数据同步
func (uc *TaskUsecase) syncOnce(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return pb.ErrorUnknown("context error=%v", ctx.Err())
	}
	uc.slog.WithContext(ctx).Infof("syncOnce executed at %s", time.Now().Format(time.RFC3339))
	return nil
}

// CleanupData performs data cleanup task
// 执行数据清理任务
func (uc *TaskUsecase) CleanupData(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return pb.ErrorUnknown("context error=%v", ctx.Err())
	}
	uc.slog.WithContext(ctx).Infof("CleanupData executed at %s", time.Now().Format(time.RFC3339))
	return nil
}

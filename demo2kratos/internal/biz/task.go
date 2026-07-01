package biz

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/yylego/kratos-cron/cronkratos"
	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
)

// TaskUsecase handles scheduled task business logic
// 处理定时任务的业务逻辑
type TaskUsecase struct {
	slog *slog.Logger
}

// NewTaskUsecase creates a new TaskUsecase instance
// 创建新的 TaskUsecase 实例
func NewTaskUsecase(logger *slog.Logger) *TaskUsecase {
	return &TaskUsecase{
		slog: logger,
	}
}

// SyncData performs data sync in loop, each iteration inside the stage read-lock
// 循环同步数据，每次迭代在 stage 读锁内执行
func (uc *TaskUsecase) SyncData(ctx context.Context, stage *cronkratos.Stage) *errors.Error {
	for i := 0; i < 10; i++ {
		if erk := uc.syncOnce(ctx, stage); erk != nil {
			return erk
		}
	}
	uc.slog.Info("SyncData complete")
	return nil
}

// syncOnce performs single data sync inside the stage read-lock
// 在 stage 读锁内执行单次数据同步
func (uc *TaskUsecase) syncOnce(ctx context.Context, stage *cronkratos.Stage) *errors.Error {
	var erk *errors.Error
	stage.Do(ctx, func(ctx context.Context) {
		// Check ctx validity inside the read-lock, exit if cancelled
		// 在读锁内检查 ctx 是否有效，已取消则退出
		if ctx.Err() != nil {
			erk = pb.ErrorUnknown("context error=%v", ctx.Err())
			return
		}
		uc.slog.InfoContext(ctx, "syncOnce executed", "time", time.Now().Format(time.RFC3339))
	})
	return erk
}

// CleanupData performs data cleanup inside the stage read-lock
// 在 stage 读锁内执行数据清理
func (uc *TaskUsecase) CleanupData(ctx context.Context, stage *cronkratos.Stage) *errors.Error {
	var erk *errors.Error
	stage.Do(ctx, func(ctx context.Context) {
		if ctx.Err() != nil {
			erk = pb.ErrorUnknown("context error=%v", ctx.Err())
			return
		}
		uc.slog.InfoContext(ctx, "CleanupData executed", "time", time.Now().Format(time.RFC3339))
	})
	return erk
}

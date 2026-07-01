# Changes

Code differences compared to source project.

## cmd/demo2kratos/main.go (+3 -1)

```diff
@@ -13,6 +13,7 @@
 	"github.com/go-kratos/kratos/v3/transport/grpc"
 	"github.com/go-kratos/kratos/v3/transport/http"
 	"github.com/yylego/done"
+	"github.com/yylego/kratos-cron/cronkratos"
 	"github.com/yylego/kratos-examples/demo2kratos/internal/conf"
 	"github.com/yylego/must"
 	"github.com/yylego/rese"
@@ -34,7 +35,7 @@
 	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
 }
 
-func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
+func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server, cs *cronkratos.Server) *kratos.App {
 	return kratos.New(
 		kratos.ID(done.VCE(os.Hostname()).Omit()),
 		kratos.Name(Name),
@@ -44,6 +45,7 @@
 		kratos.Server(
 			gs,
 			hs,
+			cs,
 		),
 	)
 }
```

## cmd/demo2kratos/wire_gen.go (+4 -1)

```diff
@@ -36,7 +36,10 @@
 	articleService := service.NewArticleService(articleUsecase)
 	grpcServer := server.NewGRPCServer(confServer, articleService, logger)
 	httpServer := server.NewHTTPServer(confServer, articleService, logger)
-	app := newApp(logger, grpcServer, httpServer)
+	taskUsecase := biz.NewTaskUsecase(logger)
+	cronService := service.NewCronService(taskUsecase, logger)
+	cronkratosServer := server.NewCronServer(cronService, logger)
+	app := newApp(logger, grpcServer, httpServer, cronkratosServer)
 	return app, func() {
 		cleanup()
 	}, nil
```

## internal/biz/biz.go (+1 -1)

```diff
@@ -2,4 +2,4 @@
 
 import "github.com/google/wire"
 
-var ProviderSet = wire.NewSet(NewArticleUsecase)
+var ProviderSet = wire.NewSet(NewArticleUsecase, NewTaskUsecase)
```

## internal/biz/task.go (+67 -0)

```diff
@@ -0,0 +1,67 @@
+package biz
+
+import (
+	"context"
+	"log/slog"
+	"time"
+
+	"github.com/go-kratos/kratos/v3/errors"
+	"github.com/yylego/kratos-cron/cronkratos"
+	pb "github.com/yylego/kratos-examples/demo2kratos/api/article"
+)
+
+// TaskUsecase handles scheduled task business logic
+// 处理定时任务的业务逻辑
+type TaskUsecase struct {
+	slog *slog.Logger
+}
+
+// NewTaskUsecase creates a new TaskUsecase instance
+// 创建新的 TaskUsecase 实例
+func NewTaskUsecase(logger *slog.Logger) *TaskUsecase {
+	return &TaskUsecase{
+		slog: logger,
+	}
+}
+
+// SyncData performs data sync in loop, each iteration inside the stage read-lock
+// 循环同步数据，每次迭代在 stage 读锁内执行
+func (uc *TaskUsecase) SyncData(ctx context.Context, stage *cronkratos.Stage) *errors.Error {
+	for i := 0; i < 10; i++ {
+		if erk := uc.syncOnce(ctx, stage); erk != nil {
+			return erk
+		}
+	}
+	uc.slog.Info("SyncData complete")
+	return nil
+}
+
+// syncOnce performs single data sync inside the stage read-lock
+// 在 stage 读锁内执行单次数据同步
+func (uc *TaskUsecase) syncOnce(ctx context.Context, stage *cronkratos.Stage) *errors.Error {
+	var erk *errors.Error
+	stage.Do(ctx, func(ctx context.Context) {
+		// Check ctx validity inside the read-lock, exit if cancelled
+		// 在读锁内检查 ctx 是否有效，已取消则退出
+		if ctx.Err() != nil {
+			erk = pb.ErrorUnknown("context error=%v", ctx.Err())
+			return
+		}
+		uc.slog.InfoContext(ctx, "syncOnce executed", "time", time.Now().Format(time.RFC3339))
+	})
+	return erk
+}
+
+// CleanupData performs data cleanup inside the stage read-lock
+// 在 stage 读锁内执行数据清理
+func (uc *TaskUsecase) CleanupData(ctx context.Context, stage *cronkratos.Stage) *errors.Error {
+	var erk *errors.Error
+	stage.Do(ctx, func(ctx context.Context) {
+		if ctx.Err() != nil {
+			erk = pb.ErrorUnknown("context error=%v", ctx.Err())
+			return
+		}
+		uc.slog.InfoContext(ctx, "CleanupData executed", "time", time.Now().Format(time.RFC3339))
+	})
+	return erk
+}
```

## internal/server/cron.go (+24 -0)

```diff
@@ -0,0 +1,24 @@
+package server
+
+import (
+	"log/slog"
+	"time"
+
+	"github.com/robfig/cron/v3"
+	"github.com/yylego/kratos-cron/cronkratos"
+	"github.com/yylego/kratos-examples/demo2kratos/internal/service"
+)
+
+// NewCronServer creates a new cron server and registers cron jobs with stage locking
+// 创建新的 cron server 并注册带 stage 读锁协调的定时任务
+func NewCronServer(cronService *service.CronService, logger *slog.Logger) *cronkratos.Server {
+	srv := cronkratos.NewServer(
+		cron.New(
+			cron.WithSeconds(),
+			cron.WithLocation(time.FixedZone("CST", 8*60*60)), // UTC+8
+		),
+		logger,
+	)
+	cronService.RegisterCron(srv)
+	return srv
+}
```

## internal/server/server.go (+1 -1)

```diff
@@ -5,4 +5,4 @@
 )
 
 // ProviderSet is server providers.
-var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)
+var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewCronServer)
```

## internal/service/cron.go (+47 -0)

```diff
@@ -0,0 +1,47 @@
+package service
+
+import (
+	"context"
+	"log/slog"
+
+	"github.com/yylego/kratos-cron/cronkratos"
+	"github.com/yylego/kratos-examples/demo2kratos/internal/biz"
+	"github.com/yylego/rese"
+)
+
+// CronService handles cron job registration with stage locking
+// 处理带 stage 读锁协调的定时任务注册
+type CronService struct {
+	task *biz.TaskUsecase
+	slog *slog.Logger
+}
+
+// NewCronService creates a new CronService instance
+// 创建新的 CronService 实例
+func NewCronService(task *biz.TaskUsecase, logger *slog.Logger) *CronService {
+	return &CronService{task: task, slog: logger}
+}
+
+// RegisterCron registers cron jobs; each task drives the stage read-lock for safe shutdown
+// 注册定时任务，每个 task 用 stage 读锁配合安全退出
+func (s *CronService) RegisterCron(srv *cronkratos.Server) {
+	// Sync data every minute, lock each iteration via stage
+	// 每分钟同步数据，在业务层循环中每次迭代用 stage 加锁
+	rese.C1(srv.AddFunc("0 * * * * *", func(ctx context.Context, stage *cronkratos.Stage) {
+		if erk := s.task.SyncData(ctx, stage); erk != nil {
+			s.slog.Error("sync data task error", "err", erk)
+		} else {
+			s.slog.Info("sync data task success")
+		}
+	}))
+
+	// Cleanup data every second, lock via stage
+	// 每秒清理数据，用 stage 加锁
+	rese.C1(srv.AddFunc("* * * * * *", func(ctx context.Context, stage *cronkratos.Stage) {
+		if erk := s.task.CleanupData(ctx, stage); erk != nil {
+			s.slog.Error("cleanup data task error", "err", erk)
+		} else {
+			s.slog.Info("cleanup data task success")
+		}
+	}))
+}
```

## internal/service/service.go (+1 -1)

```diff
@@ -2,4 +2,4 @@
 
 import "github.com/google/wire"
 
-var ProviderSet = wire.NewSet(NewArticleService)
+var ProviderSet = wire.NewSet(NewArticleService, NewCronService)
```


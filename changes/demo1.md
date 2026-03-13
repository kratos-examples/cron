# Changes

Code differences compared to source project.

## cmd/demo1kratos/main.go (+3 -1)

```diff
@@ -12,6 +12,7 @@
 	"github.com/go-kratos/kratos/v2/transport/grpc"
 	"github.com/go-kratos/kratos/v2/transport/http"
 	"github.com/yylego/done"
+	"github.com/yylego/kratos-cron/cronkratos"
 	"github.com/yylego/kratos-examples/demo1kratos/internal/conf"
 	"github.com/yylego/must"
 	"github.com/yylego/rese"
@@ -31,7 +32,7 @@
 	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
 }
 
-func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
+func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, cs *cronkratos.Server) *kratos.App {
 	return kratos.New(
 		kratos.ID(done.VCE(os.Hostname()).Omit()),
 		kratos.Name(Name),
@@ -41,6 +42,7 @@
 		kratos.Server(
 			gs,
 			hs,
+			cs,
 		),
 	)
 }
```

## cmd/demo1kratos/wire_gen.go (+4 -1)

```diff
@@ -27,7 +27,10 @@
 	studentService := service.NewStudentService(studentUsecase)
 	grpcServer := server.NewGRPCServer(confServer, studentService, logger)
 	httpServer := server.NewHTTPServer(confServer, studentService, logger)
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
 
-var ProviderSet = wire.NewSet(NewStudentUsecase)
+var ProviderSet = wire.NewSet(NewStudentUsecase, NewTaskUsecase)
```

## internal/biz/task.go (+56 -0)

```diff
@@ -0,0 +1,56 @@
+package biz
+
+import (
+	"context"
+	"time"
+
+	"github.com/go-kratos/kratos/v2/errors"
+	"github.com/go-kratos/kratos/v2/log"
+	pb "github.com/yylego/kratos-examples/demo1kratos/api/student"
+)
+
+// TaskUsecase handles scheduled task business logic
+// 处理定时任务的业务逻辑
+type TaskUsecase struct {
+	slog *log.Helper
+}
+
+// NewTaskUsecase creates a new TaskUsecase instance
+// 创建新的 TaskUsecase 实例
+func NewTaskUsecase(logger log.Logger) *TaskUsecase {
+	return &TaskUsecase{
+		slog: log.NewHelper(logger),
+	}
+}
+
+// SyncData performs data sync in loop
+// 循环同步数据
+func (uc *TaskUsecase) SyncData(ctx context.Context) *errors.Error {
+	for i := 0; i < 10; i++ {
+		if erk := uc.syncOnce(ctx); erk != nil {
+			return erk
+		}
+	}
+	uc.slog.Info("SyncData complete")
+	return nil
+}
+
+// syncOnce performs single data sync
+// 执行单次数据同步
+func (uc *TaskUsecase) syncOnce(ctx context.Context) *errors.Error {
+	if ctx.Err() != nil {
+		return pb.ErrorUnknown("context error=%v", ctx.Err())
+	}
+	uc.slog.WithContext(ctx).Infof("syncOnce executed at %s", time.Now().Format(time.RFC3339))
+	return nil
+}
+
+// CleanupData performs data cleanup task
+// 执行数据清理任务
+func (uc *TaskUsecase) CleanupData(ctx context.Context) *errors.Error {
+	if ctx.Err() != nil {
+		return pb.ErrorUnknown("context error=%v", ctx.Err())
+	}
+	uc.slog.WithContext(ctx).Infof("CleanupData executed at %s", time.Now().Format(time.RFC3339))
+	return nil
+}
```

## internal/server/cron.go (+24 -0)

```diff
@@ -0,0 +1,24 @@
+package server
+
+import (
+	"time"
+
+	"github.com/go-kratos/kratos/v2/log"
+	"github.com/robfig/cron/v3"
+	"github.com/yylego/kratos-cron/cronkratos"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/service"
+)
+
+// NewCronServer creates a new cron server and registers cron jobs
+// 创建新的 cron server 并注册定时任务
+func NewCronServer(cronService *service.CronService, logger log.Logger) *cronkratos.Server {
+	srv := cronkratos.NewServer(
+		cron.New(
+			cron.WithSeconds(),
+			cron.WithLocation(time.FixedZone("CST", 8*60*60)), // UTC+8
+		),
+		logger,
+	)
+	cronkratos.RegisterCronServer(srv, cronService)
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
+
+	"github.com/go-kratos/kratos/v2/log"
+	"github.com/robfig/cron/v3"
+	"github.com/yylego/kratos-examples/demo1kratos/internal/biz"
+	"github.com/yylego/rese"
+)
+
+// CronService handles cron job registration
+// 处理定时任务注册
+type CronService struct {
+	task *biz.TaskUsecase
+	slog *log.Helper
+}
+
+// NewCronService creates a new CronService instance
+// 创建新的 CronService 实例
+func NewCronService(task *biz.TaskUsecase, logger log.Logger) *CronService {
+	return &CronService{task: task, slog: log.NewHelper(logger)}
+}
+
+// RegisterCron registers cron jobs to cron instance
+// 注册定时任务到 cron 实例
+func (s *CronService) RegisterCron(ctx context.Context, c *cron.Cron) {
+	// Sync data every minute
+	// 每分钟同步数据
+	rese.C1(c.AddFunc("0 * * * * *", func() {
+		if erk := s.task.SyncData(ctx); erk != nil {
+			s.slog.Errorf("sync data task error: %v", erk)
+		} else {
+			s.slog.Info("sync data task success")
+		}
+	}))
+
+	// Cleanup data every second
+	// 每秒清理数据
+	rese.C1(c.AddFunc("* * * * * *", func() {
+		if erk := s.task.CleanupData(ctx); erk != nil {
+			s.slog.Errorf("cleanup data task error: %v", erk)
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
 
-var ProviderSet = wire.NewSet(NewStudentService)
+var ProviderSet = wire.NewSet(NewStudentService, NewCronService)
```


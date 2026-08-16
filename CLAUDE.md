# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> 本仓库是 `aitrading` 工作区的子模块。整体架构、本地开发环境、跨仓库依赖见上一级 `../CLAUDE.md`。本文件只补充 `go-infra` 内部的工作方式。

## 这个库是什么

`github.com/ranxx/go-infra` —— 共享基础设施库，**纯库，无 main 包**。Go 1.26.1。

后端 `ai-trading-server` 通过 `replace github.com/ranxx/go-infra => ../go-infra` 本地引用，因此**改动这里会立即影响后端构建，无需发版**。`make release` 仅用于给需要按 tag 拉取的外部消费者打版本。

## 命令

```bash
go build ./...                              # 编译整个库
go test ./...                               # 跑全部测试
go test -race ./...                         # 运行竞态检测
go vet ./...                                 # 运行标准静态检查
go test -v ./config/...                     # 跑单个包
go test -count=1 -run '^TestName$' ./config/ # 跑单个测试并跳过缓存
gofmt -w path/to/changed.go                 # 格式化修改过的 Go 文件

make release                                # ⚠️ 危险：自动 git add . && commit && push，并把最新 vX.Y.Z 的 patch+1 打 tag 推到远程。
                                            # 不要随手执行——它会提交工作区里所有未暂存改动。
```

当前测试覆盖很薄，只有 `config/`（defaults、loader）、`dedup/`、`eventbus/` 有测试。改动核心逻辑时需要手动验证。

## 两套配置系统（最容易混淆的点）

仓库里**并存两种配置加载机制**，新代码默认用后者：

1. **Provider 链式加载** —— `config/config.go` 的 `Config` 结构体。面向远程配置中心（Apollo 风格）：`Provider.GetValue(key)` 返回一段 JSON 字符串，按传入顺序回退取第一个非空的。用法是链式调用 `NewConfig(p...).LoadMySQL().LoadRedis().LoadLog().Error()`。底层走 `LoadByKey`（先 `ApplyDefaults` 填 `default` tag，再 `json.Unmarshal`）。

2. **YAML + 环境变量** —— `config/loader.go` 的 `Load(path, &cfg)` / `LoadOrDefault(path, &cfg)`。这是后端各服务实际在用的方式，优先级 **env > yaml > default tag > 零值**。识别三种 struct tag：`yaml:"x"`、`env:"ENV_VAR"`、`default:"x"`。`LoadOrDefault` 在文件缺失时不报错、退回 default+env。

两者共用 `config/defaults.go` 的 `ApplyDefaults`：按 `default` tag 递归填零值字段，支持 `time.Duration`（如 `default:"24h"`）和指针字段。**所有基础组件的 `Config` 都靠 `default` tag 提供默认值**，新增配置字段时记得带上。

## 贯穿全库的约定（动手前先理解）

- **单例 `Init(cfg)` + `Get()` + `sync.Once`**：`redis`、`postgres`、`nats`、`dlock`、`topic`、`task` 都是这个模式——服务启动时 `Init` 一次，业务代码用 `Get()` 取全局实例。同时通常也暴露 `NewXxx(cfg)` 构造非全局实例。新增有状态组件时沿用此约定。
- **函数式选项**：`logger.WithLevel(...)`、`task.Option`、`network.WithAddress(...)`、`middleware.WithCORS(...)`、`rate.WithCleanupInterval(...)` 等。`logger.Config` 同时提供 struct-tag 配置和 `Options` 两套入口。
- **代理拨号**：存储客户端（如 `redis`）的 `Config` 带 `Proxy bool` 字段，为 true 时用 `proxy.Wrap()` 注入自定义 `Dialer`，代理地址从 `ALL_PROXY` / `no_proxy` 环境变量读取。这是为了从境内连境外交易所/服务。新增需要外连的客户端时复用 `proxy` 包。
- **Trace ID 传播**：`tracer` 包在 `context` 里传 traceId（gRPC header 默认 `x-trace-id`）。`grpc.Server` 的 unary 拦截器自动「从 metadata 取或新生成 traceId → 注入出站 metadata → 带进结构化日志」，并打印方法名+耗时（不打请求开始日志）。HTTP 侧 `middleware.Logger` 用 `X-Trace-ID` 做同样的事。`interceptor.TraceUnaryInterceptor()` 是 gRPC **客户端**侧的对应物。

## 包速览（按职责分组）

- **存储客户端** `redis` `mysql` `postgres` `mongo` `elasticsearch`：每个都是 `config.go`（带 `default` tag）+ 客户端封装，单例 Init/Get。`mysql`/`postgres` 用 GORM。`redis` 暴露 `RedisClient` 接口（不是裸 `*redis.Client`），并支持 proxy。
- **消息/事件**：
  - `nats`：NATS Core 连接（带重试、`ConnectWithContext`）+ JetStream 辅助（`NewJetStream` 建流、`Subscribe` 持久订阅 ManualAck/MaxDeliver=3）。交易平台的 kline 事件流走这里。
  - `topic`：**另一套** NATS 抽象——用 `key.Keyer` 给 subject 加前缀，外面包一层 `{Id, Data}` JSON 信封。比 `nats` 包更老，注意别和它混用。
  - `eventbus`：进程内、基于泛型的类型安全事件总线（`Publish[T]` 非阻塞 / `Subscribe[T]` 返回只读 channel）。
  - `message`：消息编解码，`JSON` 与 `Protobuf` 两种实现。
- **HTTP** `middleware` + `limiter`：`middleware` 提供 `CORS`/`Logger`/`Recovery`/`RateLimit` 及 `Stack` 组合器（`NewStack(WithRecovery(), WithLogger(), ...).Then(mux)`，先加的在最外层）。`limiter/rate` 是令牌桶 `RateLimiter`（按 key 分桶、自动清理过期桶），被 `middleware.RateLimit` 消费。
- **gRPC** `grpc` `interceptor` `tracer`：`grpc.NewServer(cfg, logger, services...)` 自动注册标准 Health Check（K8s gRPC 探针）+ reflection + trace 拦截器；业务服务实现 `Register` 接口。
- **网络** `network`：TCP / WebSocket 统一抽象。核心接口 `Connection` / `MessageHandler` / `ConnectionManager` / `Server` 在 `network.go`，实现分在 `tcp/`、`ws/` 子包，配 `Coder`（编解码）+ `Packer`（粘包处理）。
- **协调/并发** `dlock` `dedup` `task`：
  - `dlock`：基于 Redis 的分布式锁，SETNX 取锁、Lua 脚本保证「值匹配才释放/续期」防误删。`TryAcquirePeriodLock` 用于「定时任务每周期只让一个 Pod 跑一次」。另有 `redis/try_lock.go` 是更轻量的单文件实现。
  - `dedup`：**为 collector 场景定制**——多个采集实例并行时，用 Redis `SET NX EX` 保证同一条 tick 只发布一次到 NATS；`TryPublish` 在 Redis 出错时返回 error 让调用方降级（直接发布）。键格式见 `dedup.Key`（含 exchange/instType/symbol 隔离）。
  - `task`：基于 `robfig/cron` 的调度器，`Schedule` 接口 + `DurationSchedule` 间隔调度，`sync.Once` 单例。
- **基础工具** `logger`（Logrus，可挂 ES Hook，console/file/both 输出）、`key`（带前缀分隔符的 `Keyer`，给缓存 key / NATS subject 加命名空间）、`proxy`（SOCKS5/HTTP）、`utils`（hash/math/rand/uuid/time/stack/notify）。

## 注意事项

- 给任意 `Config` 加字段时，三处要对齐：`json` tag、`yaml` tag、`default` tag——两套配置系统分别依赖 json 和 yaml。
- `nats` 与 `topic` 功能重叠，确认要改的是哪一套；新功能优先用 `nats` 包。
- 这是被 9+ 个后端服务共享的库，改动公共接口（如 `redis.RedisClient`、`config.Provider`、`network.Connection`）会波及所有消费者，先评估影响面。

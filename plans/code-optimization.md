# 代码优化计划

## 背景

对 webtools 项目 12 个 Go 源文件进行代码审查后，整理出按优先级排序的优化建议，涵盖正确性修复、性能优化、代码质量问题。

---

## P0 — 正确性修复（优先修复）

### 1. timing_wheel: 删除或实现空桩方法

**文件**: `external/pkg/utils/timing_wheel.go`

`CancelTask`（L288）和 `UpdateTask`（L294）仅有 `return nil`，且 `task` 结构体无 `taskId` 字段，这两个方法无法工作。

**改动**:
- a) 删除这两个空桩方法及 `taskID string` 参数
- 或 b) 给 `task` 添加 ID 字段，并实现真正的取消/更新逻辑（用 map 索引任务）

**推荐方案 a**，即删除空桩方法，保持 API 最小化。

### 2. gnet: 修复 PongState 竞态条件

**文件**: `external/pkg/utils/gnet.go`、`external/pkg/utils/gnet_test.go`

`WSContext.PongState` 在 `read()`（L240-243）和 `startPing()`（gnet_test.go L79）中并发读写，未加同步保护。虽然 `Write` 有 mutex，但 `PongState` 未纳入保护范围。

**改动**:
- 将 `PongState bool` 改为 `pongState atomic.Bool`
- `read()` 中用 `w.pongState.Store(true)` 替代 `w.PongState = true`
- `startPing()` 中用 `ctx.pongState.Load()` 和 `ctx.pongState.Store(false)` 替代直接访问
- `WSContext` 中 `PongState` 移除，新增小写字段 `pongState atomic.Bool`

**验证**: `go test -race ./external/pkg/utils/ -run TestWs`

### 3. logger: 修复日志双重输出

**文件**: `external/pkg/utils/logger.go`

`CustomHandler.Handle()` 直接 `fmt.Print` 自定义格式的输出，但未阻止底层 handler 的输出。当 slog 框架调用 `Handle` 时，底层 `slog.TextHandler` 也会输出，造成每条日志打印两次。

**改动**:
- `Handle()` 中只做自定义格式输出，不再调用底层 handler 的 `Handle`
- `WithAttrs` 和 `WithGroup` 返回的 `CustomHandler` 也应保持自定义格式，而非仅包装底层 handler

---

## P1 — 缺陷修复

### 4. http: 修复重试逻辑

**文件**: `external/pkg/utils/http.go`

两个问题：
- L67-72：当 `req.GetBody == nil` 时重试，后续请求 body 已空。修复：在 `request()` 中统一用 `bytes.Buffer` 构造 body，确保 `GetBody` 可用
- L74-81：5xx 重试与网络错误重试逻辑混在一起，5xx 时也只应在可重试场景下继续

**改动**:
- `request()` 中将 `bytes.NewReader(body)` 改为 `bytes.NewBuffer(body)`，确保 `http.NewRequest` 自动生成 `GetBody`
- `doWithRetry()` 中 5xx 分支也调用 `isRetriableError`（当前5xx总是可重试，逻辑上合理但应显式处理）

---

## P2 — 代码质量

### 5. dispatcher: 添加优雅关闭

**文件**: `external/pkg/utils/dispatcher.go`

**改动**:
- 添加 `Stop()` 方法，关闭所有 shard channel，等待 goroutine 退出
- 添加 `sync.WaitGroup` 跟踪 `processTasks` goroutine

### 6. response: 抽取重复代码

**文件**: `external/pkg/utils/response.go`

`Fail`、`ValidateError`、`ServerError` 中 message 提取逻辑重复 3 次。

**改动**: 抽取 `func extractMessage(messages ...string) string` 辅助函数。

### 7. config: 修复泛型类型检查

**文件**: `external/pkg/utils/config.go`

L44 `reflect.TypeOf(cfg).Kind() != reflect.Struct` 对指针类型等会误判。

**改动**: 改为 `reflect.TypeOf(new(T)).Elem().Kind() != reflect.Struct`

### 8. page: 修复注释

**文件**: `external/pkg/utils/page.go`

L15 注释说"初始容量设为0"但代码是 `make([]T, 0, req.PageSize)`，实际预分配了容量。

**改动**: 修正注释为准确描述，或直接删除注释。

---

## 验证方案

1. `go vet ./...` — 无新增 warning
2. `go build ./...` — 编译通过
3. `go test -race ./...` — 无竞态检测报错
4. 手动启动 gnet WebSocket 服务端，用 wscat 连接验证 PongState 修复
5. 手动验证日志无重复输出

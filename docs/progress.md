# codex-watch 项目进度

> 更新时间：2026-04-14

## 已完成

### 核心功能

- [x] **PTY 包装 codex 命令** — 通过 `creack/pty` 启动 codex，支持 raw stdin、stdout 透传
- [x] **会话文件发现与匹配** — 扫描 `~/.codex/sessions/*.jsonl`，支持 fresh / resume / fork / explicit session / `-C`/`--cd` 切目录等模式
- [x] **JSONL 事件解析** — 解析 `session_meta`、`thread.started`、`turn_context`、`error`、`event_msg`（含 `token_count` 和 `task_complete`）
- [x] **实时状态栏** — 底部预留一行，250ms 刷新，显示 model、elapsed、token 统计、context%、rate limit%、estimated cost
- [x] **会话 tail** — 实时追踪 JSONL 文件新增行，持续更新 State
- [x] **Summary 持久化** — 会话结束后保存到 `~/.local/state/codex-watch/sessions/`，支持 XDG_STATE_HOME
- [x] **Summary 加载与排序** — `LoadAll` 按 StartedAt 降序返回所有历史 summary
- [x] **跳过空 summary** — 无有效数据（无 token、仅 lifecycle）时 `ErrSkipSave`
- [x] **report 命令** — 支持 `--latest`、`--session`、`--status`、`--model`、`--cwd`、`--limit`、`--json` 过滤与输出
- [x] **report 文本输出** — 单条详细视图 / 多条紧凑列表视图
- [x] **pricing 估算** — 静态价格表支持 gpt-5、gpt-5-mini、o4-mini，前缀匹配，未知模型返回 `cost unknown`
- [x] **debug 日志** — `CODEX_WATCH_DEBUG=1` 输出会话匹配过程诊断信息
- [x] **信号处理** — SIGINT/SIGTERM 转发给 codex 子进程，SIGWINCH 动态调整 PTY 尺寸
- [x] **非 TTY 降级** — 非终端环境跳过状态栏渲染、跳过 raw mode

### 测试

- [x] `events_test.go` — token_count fixture、turn_context、readSnapshot、非法 JSON 处理、空行/畸形行跳过
- [x] `store_test.go` — Save/LoadAll 生命周期、空 summary 跳过、error 退出仍持久化
- [x] `tailer_test.go` — DetectMatchOptions（resume --last、explicit session、-C）、FindCandidate 排序、looksLikeSessionID、损坏文件跳过、FindCandidateWithDebug 日志内容
- [x] `run_test.go` — statusRenderer.Finish 清除底部行并打印 summary、render 注入 row/width 及截断
- [x] `report_test.go` — filterSummaries 各过滤条件、formatElapsed 优先级

### 项目基础

- [x] Go module 初始化（Go 1.26）
- [x] CLI 入口 + 子命令分发（codex / report / help）
- [x] README 使用说明、构建方式、debug 指南、已知限制
- [x] roadmap 路线图文档

---

## 待办事项

### Phase 1: 基础可靠性（最高优先级）

- [ ] **pricing 单元测试** — pricing 包无测试，需覆盖精确匹配、前缀匹配、未知模型、空字符串等边界
- [ ] **事件 fixture 扩充** — 增加更多 codex-cli 版本变体的 JSONL 样例，验证兼容性解析
- [ ] **summary 落盘语义明确化** — 失败时不覆盖有效 summary、异常退出标记、空摘要跳过策略
- [ ] **终端渲染稳定性** — 解决 flicker、异常退出后底部状态残留问题
- [ ] **debug 输出增强** — 按阶段说明"为什么选/不选某个候选"，提供更多诊断信息

### Phase 2: 报表与分析

- [ ] **时间范围过滤** — report 支持 `--since` / `--until` 日期区间过滤
- [ ] **聚合摘要视图** — 总 token、总成本、按模型占比、按 cwd 排行、失败会话统计
- [ ] **JSON schema 固化** — report `--json` 输出定义为稳定 schema，字段只追加不重命名
- [ ] **价格表版本化** — 建立模型价格维护机制，降低更新成本
- [ ] **历史 summary 向后兼容** — 新字段缺失不破坏旧记录加载

### Phase 3: 配置与发布

- [ ] **配置文件层** — 统一管理 session 根目录、状态栏开关、价格覆盖、默认输出格式、debug 开关
- [ ] **安装与发布** — 版本号、release 构建、变更日志、shell completion
- [ ] **平台兼容性** — Linux/macOS 常见终端一致行为验证
- [ ] **兼容性声明** — 文档化支持的 codex-cli 版本范围和已知限制

### Phase 4: 可扩展（延后）

- [ ] **导出格式扩展** — CSV / Markdown 摘要导出
- [ ] **hook / exporter 接口** — 预留扩展点，默认本地文件中心
- [ ] **高级看板** — 目录级汇总、多机器同步（待评估需求）

---

## 当前代码统计

| 包 | 源文件 | 测试文件 | 说明 |
|---|---|---|---|
| `cmd/codex-watch` | `main.go` | 无 | CLI 入口 |
| `internal/codexwatch` | `run.go` | `run_test.go` | PTY 包装 + 状态栏 |
| `internal/session` | `types.go` `events.go` `store.go` `tailer.go` | `events_test.go` `store_test.go` `tailer_test.go` | 核心会话逻辑 |
| `internal/pricing` | `pricing.go` | **无** | 价格估算 |
| `internal/report` | `report.go` | `report_test.go` | 报表输出 |

## 下一步

优先完成 Phase 1 中的 **pricing 单元测试**，这是当前唯一无测试覆盖的包，且是基础可靠性的一部分。

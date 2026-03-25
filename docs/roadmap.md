# codex-watch 未来发展路线

## Summary

将 `codex-watch` 明确定位为“面向个人开发者的本地 Codex 会话观察与成本分析 CLI”，短期不做团队平台。未来 1-2 个月优先级放在“稳定性与正确性”，先把会话识别、事件解析、状态栏展示、历史数据可靠性做扎实，再扩展报表与易用性。

成功标准：

- 日常 `codex` / `resume --last` 使用时，状态栏能稳定跟上真实会话
- 本地落盘摘要可信，可作为后续报表唯一数据源
- `report` 能回答“我最近用了多少 token、花了多少钱、在哪些项目里消耗最多”
- 项目能跟随 `codex-cli` 版本变化较低成本维护

## 路线图

### Phase 1: 基础可靠性（最高优先级，先做）

- 强化 session 匹配逻辑，覆盖 fresh / resume / explicit session / fork / `-C` 切目录等路径，降低误匹配和空状态栏概率。
- 把事件解析从“基于当前观察到的 JSONL”提升到“兼容多个已知变体”，引入 fixture 驱动测试，单独验证 `session_meta`、`turn_context`、`token_count`、`task_complete`、`error` 等事件。
- 明确 summary 落盘语义：什么时候保存、失败会不会覆盖、异常退出如何标记、空摘要如何跳过。
- 改善终端渲染稳定性，至少解决 flicker、异常退出后底部状态残留、非 TTY 场景的降级输出。
- 增加可诊断性：debug 输出按阶段说明“匹配了哪个会话、为什么选它、为什么没选其他候选”。

交付完成后的门槛：

- 针对本地不同启动方式有稳定回归测试
- 出现状态栏空白或数据异常时，用户能靠 debug 信息自查

### Phase 2: 报表与分析（第二优先级）

- 扩展 `report`，补齐时间维度和聚合能力：最近 N 次、按日期区间、按模型、按项目目录、按状态聚合。
- 增加摘要型视图，而不只是一条条列出会话：总 token、总成本、按模型占比、按 cwd 排行、失败会话统计。
- 固化 JSON 输出为稳定 schema，保证后续 shell 脚本或外部工具可依赖。
- 给 price table 建立版本化维护方式，减少模型价格更新时的改动成本；未知模型应显式标记为 `cost unknown`，不能静默给错价。
- 为历史 summary 增加向后兼容策略，未来字段新增不破坏已有存量记录。

这一阶段的目标是让 `report` 从“查看记录”升级为“本地会话分析入口”。

### Phase 3: 配置与发布体验（第三优先级）

- 引入轻量配置层，统一管理 session 根目录、状态栏开关、价格覆盖、默认 report 输出格式、debug 开关。
- 保持命令面简洁：顶层仍保留 `codex-watch codex` 和 `codex-watch report`，不扩张成多层复杂子命令树。
- 做安装与发布基础设施：版本号、release 构建、README 示例、变更日志、shell completion。
- 明确平台支持策略，至少在 Linux/macOS 的常见终端场景保持一致行为。
- 增加面向用户的“兼容性声明”：支持哪些 `codex-cli` 版本、已知限制是什么。

### Phase 4: 可扩展但保持本地优先（延后）

- 增加导出能力，如 `report --json` 之外的 CSV/Markdown 摘要，方便本地记账或周报。
- 预留 hook / exporter 接口，但默认仍以本地文件为中心，不引入常驻服务和远端依赖。
- 如果个人使用价值已经跑通，再评估是否做“目录级汇总看板”或“多机器同步”，但不提前进入团队平台路线。

## Public APIs / Interfaces

保持现有公共入口稳定：

- `codex-watch codex [codex args...]`
- `codex-watch report [...]`

下一阶段允许扩展但不破坏的接口方向：

- `report` 新增时间范围和聚合类 flags，例如“时间区间”“summary 模式”“按字段分组”
- `report --json` 输出定义为稳定 schema；新增字段只能追加，不能重命名已有字段
- summary 仍作为本地持久化核心对象，继续围绕 `session_id`、`thread_id`、模型、cwd、token、rate limit、cost、status 扩展
- 可选配置文件作为增量能力引入，不要求现有用户迁移

## Test Plan

- 事件 fixture 测试：不同 `codex-cli` JSONL 样例都能解析出一致 summary。
- 匹配逻辑测试：fresh、resume、explicit session、`-C/--cd`、多候选竞争、损坏文件跳过。
- 状态栏测试：TTY 与非 TTY、窗口 resize、中断退出、子进程异常退出。
- 存储兼容测试：旧 summary 文件可加载，新字段缺失不报错。
- 报表测试：文本输出、JSON 输出、时间过滤、cwd/model/status 聚合、未知价格模型。
- 端到端 smoke test：启动包装命令、读取样例 session、产出 summary、再用 `report` 查询到结果。

## Assumptions

- 产品定位默认是“个人效率工具”，不是团队 SaaS 或集中式成本平台。
- 近期开发资源有限，优先把正确性和维护性拉高，而不是快速堆更多 flags。
- 现有本地 summary 文件将继续作为唯一事实来源，后续分析能力都建立在它之上。
- 兼容 `codex-cli` 的演进会是长期成本项，因此 fixture 和兼容测试必须尽早建设。

# Origin Check

AI API 中转站评测平台：后台定时探测各中转站的**模型真伪**与**性能指标**，前端展示排行榜与详细报告。

## 架构

- **Go 1.25+** — API / Scheduler / Worker
- **PostgreSQL** — 探测结果与鉴定报告
- **Redis** — 任务队列
- **Next.js 14** — 展示站

## 快速开始

### 1. 启动基础设施

```bash
cp .env.example .env
make up
```

### 2. 配置 API Key（可选，无 Key 时仅入库不探测）

在 `.env` 中填入各中转站 Key：

```
LIAOBOTS_API_KEY=your-login-code
LINGYAAI_API_KEY=sk-...
ASIAI_API_KEY=sk-...
NICEGOAL_API_KEY=sk-...
```

### 3. 导入种子数据

```bash
go run ./cmd/seed
```

### 4. 启动服务

```bash
# 终端 1
go run ./cmd/api

# 终端 2
go run ./cmd/worker

# 终端 3
go run ./cmd/scheduler

# 终端 4
cd web && npm install && npm run dev
```

访问 http://localhost:3000

### Docker Compose 全栈

```bash
docker compose up --build
```

## 评测模型（主流旗舰）

性能与真伪探测覆盖以下模型（不含 DeepSeek 等国内模型及 gpt-4o / sonnet-4.5 / gemini-2.0 等旧版）：

| 系列 | 模型 ID |
|------|---------|
| GPT-5.5 | `gpt-5.5` |
| Claude Opus 4.7 | `claude-opus-4-7` |
| Gemini 3.5 | `gemini-3.5-flash` |
| 可用性探测（轻量） | `gpt-5.4-mini` |

| 名称 | 官网 | API Base |
|------|------|----------|
| Liaobots | https://liaobots.work/zh | https://ai.liaobots.work/v1 |
| 灵芽 API | https://api.lingyaai.cn/ | https://api.lingyaai.cn/v1 |
| Asiai Cloud | https://api.asiai.cloud/ | https://api.asiai.cloud/v1 |
| NiceGoal | https://www.nicegoal.ai | https://www.nicegoal.ai/v1 |

## API 端点

- `GET /api/v1/leaderboard` — 排行榜
- `GET /api/v1/relays/{id}` — 中转站详情
- `GET /api/v1/relays/{id}/metrics` — 性能指标
- `GET /api/v1/relays/{id}/reports` — 真伪报告

## 调度策略

通过环境变量配置（Go duration 格式，如 `15m`、`24h`）：

| 变量 | 默认 | 说明 |
|------|------|------|
| `PROBE_HEALTH_INTERVAL` | `15m` | 可用性轻量探测 |
| `PROBE_PERFORMANCE_INTERVAL` | `24h` | 性能探测（默认每天一次） |
| `PROBE_AUTHENTICITY_INTERVAL` | `24h` | 真伪鉴定 |
| `PROBE_*_ON_STARTUP` | health 开，performance/auth 关 | 服务启动时是否立即跑一轮 |

示例：`.env` 中设置 `PROBE_PERFORMANCE_INTERVAL=12h` 改为每 12 小时测一次性能。

## 手动触发探测（CLI）

服务运行后，可立即将探测任务入队并由 worker 执行（无需等定时调度）。

### 本地

```bash
# 列出中转站
go run ./cmd/trigger -list

# 对 Liaobots 跑全部 claimed models 真伪鉴定（默认 -wait 阻塞直到完成）
go run ./cmd/trigger -relay Liaobots -type authenticity

# 只测单个模型，不入队后等待
go run ./cmd/trigger -relay Liaobots -type authenticity -model gpt-5.5 -wait=false

# 性能 / 可用性
go run ./cmd/trigger -relay Liaobots -type performance
go run ./cmd/trigger -relay Liaobots -type health
```

### Docker（worker 容器内已包含 `/bin/trigger`）

```bash
# 查看收录的中转站
docker exec origin-check-worker-1 /bin/trigger -list

# 立即跑 Liaobots 真伪鉴定（3 个旗舰模型各一条任务）
docker exec origin-check-worker-1 /bin/trigger -relay Liaobots -type authenticity

# 只测 Claude Opus
docker exec origin-check-worker-1 /bin/trigger -relay "Liaobots" -type auth -model claude-opus-4-7
```

`-type` 可选：`authenticity`（别名 `auth`）、`performance`（`perf`）、`health`。  
默认对所有 `claimed_models` 各 enqueue 一条任务；worker 必须处于运行状态。

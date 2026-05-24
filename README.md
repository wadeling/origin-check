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
| GPT-5.5 | `gpt-5.5`, `gpt-5.5-pro` |
| Claude Opus 4.7 | `claude-opus-4-7` |
| Gemini 3.x | `gemini-3.1-pro-preview` |
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

- Health：每 15 分钟
- Performance：每 6 小时（整点）
- Authenticity：每 6 小时（整点后 30 分），启动时立即跑一轮

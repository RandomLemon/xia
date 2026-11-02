# xia

基于 [kei](https://github.com/RandomLemon/kei) 的聊天机器人宿主：kei 负责事件引擎、适配器与插件装配，[chat-agent](https://github.com/RandomLemon/kei-plugin-chat-agent) 提供对话、长期记忆与知识库 RAG。仓库自带一个 OpenAI 兼容的本地 stub，不配置任何 API key 也能跑通完整链路。

## 特性

- **双平台**：mock（本地 HTTP 控制面，便于联调）与 OneBot v11 反向 WebSocket（QQ）。
- **对话策略**：私聊直接回复；群聊通过 @ 或触发词“小霞”唤醒；默认启用低频随机插话。
- **持久化**：SQLite 保存历史、长期记忆与向量索引，重启不丢。
- **模型可替换**：走 OpenAI 兼容协议，换真实模型只改配置或环境变量。

## 目录结构

```
cmd/xia          宿主入口：装配适配器/插件，处理信号
cmd/llmstub      本地 OpenAI 兼容 stub（仅开发用）
configs/         运行配置
nix/             flake 与 nix-shell 共用的 devShell 定义
data/            SQLite 运行数据（已 gitignore）
```

## 环境准备

`go.mod` 通过 `replace` 引用同级仓库，工作区必须是：

```
code/
├── kei/
├── kei-plugin-chat-agent/
└── xia/
```

推荐使用 Nix，工具链已固定：

```bash
direnv allow          # 已启用 nix-direnv（推荐）
# 或
nix develop
# 或（不使用 flakes）
nix-shell
```

NixOS 可在 `configuration.nix` 中启用：

```nix
programs.direnv.enable = true;
programs.direnv.nix-direnv.enable = true;
```

若不用 Nix，需要 Go 1.25+、gcc 与 cgo（SQLite 驱动依赖 cgo）。

## 本地运行

默认 `configs/config.yaml` 使用两个局域网服务：

- 聊天：`http://192.168.200.101:8001/v1`，模型 `qwen3.8-flash-next`
- embedding：`http://192.168.200.101:8010/v1`，模型 `qwen3-vl-embedding-2b-q4km`，2048 维

直接启动机器人，不需要本地 stub：

```bash
go run ./cmd/xia -config configs/config.yaml
```

如果不想访问局域网服务，可完全本地运行。终端 1 启动 stub：

```bash
go run ./cmd/llmstub
```

终端 2 用环境变量把聊天和 embedding 都切到 stub：

```bash
KEI_PLUGINS_CHAT_AGENT_LLM_BASE_URL=http://127.0.0.1:18100/v1 \
KEI_PLUGINS_CHAT_AGENT_LLM_MODEL=stub-chat \
KEI_PLUGINS_CHAT_AGENT_EMBEDDING_BASE_URL=http://127.0.0.1:18100/v1 \
KEI_PLUGINS_CHAT_AGENT_EMBEDDING_MODEL=stub-embed \
KEI_PLUGINS_CHAT_AGENT_EMBEDDING_DIMENSIONS=8 \
go run ./cmd/xia -config configs/config.yaml
```

下面的 `stub:` 固定输出以“完全本地”模式为准；默认模型会返回自然语言。

启动后可用的本地端口：

| 地址 | 用途 |
|---|---|
| `127.0.0.1:18080` | mock 控制面：`POST /inject`、`GET /sent`、`GET /healthz` |
| `127.0.0.1:18082` | OneBot v11 反向 WebSocket：`/onebot/v11/ws` |
| `127.0.0.1:18100` | llmstub：`/v1/chat/completions`、`/v1/embeddings` |
| `127.0.0.1:19090` | Prometheus metrics：`/metrics` |

### 发一条私聊

```bash
curl -sS -X POST http://127.0.0.1:18080/inject \
  -H 'Content-Type: application/json' \
  -d '{"text":"你好，介绍一下你自己","user_id":"u1","user_name":"Tester","kind":"private"}'
```

等待约 2.5 秒（消息合并窗口）后查看发送记录：

```bash
curl -sS http://127.0.0.1:18080/sent \
  | jq -r '.[] | .Request.Message.Segments[] | select(.Type == "text") | .Data.text'
```

默认模型会返回自然语言；完全本地模式下预期输出：

```
stub: 你好，介绍一下你自己
```

### 发一条群聊

```bash
curl -sS -X POST http://127.0.0.1:18080/inject \
  -H 'Content-Type: application/json' \
  -d '{"text":"小霞你好","user_id":"u2","user_name":"Tester2","channel_id":"g1","kind":"group"}'
```

群聊必须 @ 机器人或包含触发词“小霞”；完全本地模式下上例会回复 `stub: 小霞你好`。

按 `Ctrl-C` 停止机器人；SQLite 数据保留在 `data/xia.db`。

## 配置

主配置是 `configs/config.yaml`，常用项：

- `bots[]`：平台实例、监听地址、OneBot 模式与密钥。
- `plugins.chat-agent`：人设、触发词、随机插话、并发/限流、模型地址、记忆与 RAG。
- `storage`：默认 `sqlite`，DSN 为 `./data/xia.db`。

所有配置键都可用 `KEI_` 前缀的环境变量覆盖，路径段以下划线连接，例如：

```bash
KEI_LOG_LEVEL=debug
KEI_STORAGE_DSN=/tmp/xia.db
KEI_PLUGINS_CHAT_AGENT_RANDOM_ENABLED=false
```

## 模型配置

默认聊天模型是 `http://192.168.200.101:8001/v1` 的 `qwen3.8-flash-next`，embedding 是
`http://192.168.200.101:8010/v1` 的 `qwen3-vl-embedding-2b-q4km`（2048 维）。聊天服务
本身不提供 embeddings，所以 embedding 单独指向 8010。换成其他 OpenAI 兼容服务：

```bash
export KEI_PLUGINS_CHAT_AGENT_LLM_BASE_URL=https://api.example.com/v1
export KEI_PLUGINS_CHAT_AGENT_LLM_MODEL=your-chat-model
export KEI_PLUGINS_CHAT_AGENT_LLM_API_KEY=sk-...

export KEI_PLUGINS_CHAT_AGENT_EMBEDDING_BASE_URL=https://api.example.com/v1
export KEI_PLUGINS_CHAT_AGENT_EMBEDDING_MODEL=your-embedding-model
export KEI_PLUGINS_CHAT_AGENT_EMBEDDING_DIMENSIONS=1536
export KEI_PLUGINS_CHAT_AGENT_EMBEDDING_API_KEY=sk-...
```

`embedding_dimensions` 必须与 embedding 服务返回的维度一致（当前真实服务是 2048，完全本地 stub 是 8）。修改维度后建议删除 `data/xia.db` 重建向量索引。也可以直接改 `configs/config.yaml` 中对应的 `llm_*` / `embedding_*` 键。如果暂时没有可用的 embedding 服务，可以关闭 `memory_enabled` / `rag_enabled`，避免检索时反复报错。

## 接入 QQ

配置里已声明 OneBot v11 反向 WebSocket 实例 `qq-main`。让 OneBot 实现（如 NapCat、Lagrange 等）连接到：

```
ws://127.0.0.1:18082/onebot/v11/ws
```

部署时注入账号信息，不要把密钥写进仓库：

```bash
export KEI_BOTS_QQ_MAIN_SECRET="your-onebot-secret"
export KEI_BOTS_QQ_MAIN_SELF_ID="123456789"
export KEI_PLUGINS_CHAT_AGENT_SELF_IDS="onebot:123456789"
```

`self_ids` 用于判定 @ 是否指向机器人本身；留空时任意 @ 都会触发回复。

管理命令默认关闭。需要时再配置管理员：

```bash
export KEI_PLUGINS_CHAT_AGENT_SUPER_ADMINS="onebot:10001"  # /chat
export KEI_AUTH_ADMIN_USERS="onebot:10001"                 # /chat-admin
```

## 数据与重置

- 默认数据库：`data/xia.db`（SQLite，已 gitignore）。
- 删除该文件并重启即可清空历史、长期记忆与向量索引。
- 临时改路径：`KEI_STORAGE_DSN=/tmp/xia.db`。

## 开发检查

```bash
gofmt -l .
go vet ./...
go build ./...
```

使用 Nix 时还可以：

```bash
nix flake check --no-build
nix-shell --run 'go build ./...'
```

## License

MIT，见 [LICENSE](LICENSE)。

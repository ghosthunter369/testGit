# 聊天历史功能改造方案（评审稿）

## 1. 目标与范围

本次需求目标：

1. 聊天页改为“左侧会话列表 + 右侧聊天窗口”布局。
2. 支持新建会话、切换会话、重命名会话、删除会话。
3. 会话消息持久化到 MySQL，不再依赖 `localStorage` 作为主存储。
4. 继续支持流式回复（SSE），并且与会话存储打通。
5. 所有非流式接口统一返回：`code` / `data` / `message`。
6. 继续遵循当前项目约束：仅允许 `GET` 和 `POST`。

## 2. 当前系统现状

已确认数据库 `clawdb` 当前只有一张表：`users`。

`users` 字段如下：

- `id` bigint PK
- `username` varchar(50) unique
- `password` varchar(255)
- `email` varchar(100)
- `created_at` datetime(3)
- `updated_at` datetime(3)

当前聊天接口是 `POST /api/chat` 流式返回，但没有会话维度持久化。

## 3. 数据库设计

建议新增 **2 张表**，满足 ChatGPT 类会话管理：

1. `chat_sessions`：会话主表（左侧列表）
2. `chat_messages`：消息明细表（右侧消息流）

### 3.1 表一：chat_sessions

用途：管理“一个会话”的元信息（标题、最后活跃时间、删除状态等）。

核心字段建议：

- `id`：会话 ID（bigint）
- `user_id`：所属用户 ID（关联 `users.id`）
- `title`：会话标题（可重命名）
- `last_message_at`：最后一条消息时间（用于左侧排序）
- `created_at` / `updated_at`
- `deleted_at`：软删除时间（支持“删除会话”）

### 3.2 表二：chat_messages

用途：存储每条消息，支持按会话回放历史。

核心字段建议：

- `id`：消息 ID（bigint）
- `session_id`：会话 ID（关联 `chat_sessions.id`）
- `user_id`：所属用户 ID（冗余字段，便于鉴权和查询）
- `role`：`user` / `assistant` / `system`
- `content`：消息内容（longtext）
- `seq`：会话内顺序号（保证消息顺序）
- `status`：消息状态（如 `1=normal, 2=error`）
- `created_at` / `updated_at`
- `deleted_at`：软删除时间

## 4. 建表 SQL（MySQL 8）

```sql
USE clawdb;

CREATE TABLE IF NOT EXISTS chat_sessions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  title VARCHAR(120) NOT NULL,
  last_message_at DATETIME(3) NULL,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  CONSTRAINT fk_chat_sessions_user
    FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_chat_sessions_user_updated
  ON chat_sessions(user_id, updated_at DESC);

CREATE INDEX idx_chat_sessions_user_deleted_last
  ON chat_sessions(user_id, deleted_at, last_message_at DESC);

CREATE TABLE IF NOT EXISTS chat_messages (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  session_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role VARCHAR(20) NOT NULL,
  content LONGTEXT NOT NULL,
  seq INT NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME(3) NULL,
  updated_at DATETIME(3) NULL,
  deleted_at DATETIME(3) NULL,
  CONSTRAINT fk_chat_messages_session
    FOREIGN KEY (session_id) REFERENCES chat_sessions(id),
  CONSTRAINT fk_chat_messages_user
    FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT uk_chat_messages_session_seq UNIQUE (session_id, seq)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_chat_messages_session_created
  ON chat_messages(session_id, created_at ASC);

CREATE INDEX idx_chat_messages_user_session
  ON chat_messages(user_id, session_id);
```

## 5. 接口设计（仅 GET/POST）

以下接口都放在鉴权组 `/api` 下，统一鉴权。

### 5.1 会话管理（左侧）

1. `GET /api/chat/sessions`
- 入参：无
- 出参：当前用户的会话列表（按 `last_message_at` / `updated_at` 倒序）

2. `POST /api/chat/session/create`
- 入参：`title`（可选，不传则默认“新对话 + 日期”）
- 出参：新建会话信息

3. `POST /api/chat/session/rename`
- 入参：`session_id`, `title`
- 出参：更新后的会话信息

4. `POST /api/chat/session/delete`
- 入参：`session_id`
- 行为：软删除会话（同时软删除该会话消息）

### 5.2 消息查询（右侧）

5. `GET /api/chat/messages?session_id=xxx&page=1&page_size=50`
- 入参：`session_id` 必填，分页可选
- 出参：会话消息列表（按 `seq` 升序）

### 5.3 流式聊天（写入消息）

6. `POST /api/chat/stream`
- 入参：`session_id`（可空，空则服务端自动建会话）、`message`
- 出参：SSE 流

SSE `data` 负载仍使用统一结构包装：

```json
{"code":0,"data":{"type":"chunk","content":"..."},"message":"ok"}
```

建议增加两个字段便于前端状态同步：

- `data.session_id`：用于“新会话自动创建”时前端拿到会话 ID
- `data.message_id`：用于消息落库后追踪（可选）

## 6. 关键后端改造点

## 6.1 模型与迁移

新增模型文件：

- `models/chat_session.go`
- `models/chat_message.go`

更新迁移入口：

- `models/user.go` 中 `AutoMigrate` 增加 `ChatSession`、`ChatMessage`

## 6.2 Handler 与路由

建议新增文件：

- `handlers/chat_session.go`（列表/创建/重命名/删除）
- `handlers/chat_message.go`（消息查询）

改造现有：

- `handlers/chat.go`
  - 入参改为包含 `session_id`
  - 发送前写入 user 消息
  - 流式过程中拼接 assistant 文本
  - 完成后落库 assistant 消息
  - 异常时在 `message` 写明错误

路由新增（`main.go`）：

- `GET /api/chat/sessions`
- `POST /api/chat/session/create`
- `POST /api/chat/session/rename`
- `POST /api/chat/session/delete`
- `GET /api/chat/messages`
- `POST /api/chat/stream`

## 6.3 查询与一致性

建议事务边界：

1. 写 user 消息 + 更新 session `last_message_at` 放一个事务。
2. assistant 完成后写 assistant 消息 + 更新 session，独立事务。
3. 如果流中断，assistant 消息可写 `status=2`（失败）或不写入（二选一，评审决定）。

## 7. 前端改造点

目标页面结构：

- 左侧：会话列表面板（新建、重命名、删除）
- 右侧：消息流 + 输入框 + 发送按钮

建议改造：

1. `frontend/src/views/Chat.vue` 拆为布局 + 子组件（可选）：
- `ChatSidebar`：会话列表
- `ChatWindow`：消息和输入

2. 前端状态来源改为接口：
- 初始化加载 `GET /api/chat/sessions`
- 切换会话加载 `GET /api/chat/messages`
- 发送消息走 `POST /api/chat/stream`

3. 错误统一展示：
- 所有失败提示优先展示后端 `message`

4. `localStorage` 仅用于 UI 辅助（如当前选中的 `session_id`），不保存消息主数据。

## 8. 配置与环境变量建议

可选新增配置（`.env`）：

- `CHAT_DEFAULT_TITLE=新对话`
- `CHAT_CONTEXT_MAX_MESSAGES=20`（组装模型上下文时最多取最近 N 条）
- `CHAT_MESSAGE_PAGE_SIZE=50`

说明：本地 DB 配置维持现状（`localhost:3306`, `root/2771651667`, `clawdb`）。

## 9. 实施顺序建议

1. 先落库：新增 2 张表 + GORM 模型 + 迁移。
2. 先做会话基础接口（创建/列表/删除/重命名/消息查询）。
3. 再改造流式聊天接口接入会话写入。
4. 最后改前端布局和交互（左侧会话 + 右侧流式聊天）。
5. 联调验收：重点看“切换会话、刷新后恢复、删除后状态、错误 message 展示”。

## 10. 验收标准（建议）

1. 用户可新建多个会话，并在左侧看到按时间排序的列表。
2. 点击不同会话，右侧消息正确切换。
3. 发送消息时可流式显示，并落库。
4. 刷新页面后可完整恢复会话与消息。
5. 重命名、删除会话即时生效。
6. 所有失败提示都来自后端 `message` 字段。


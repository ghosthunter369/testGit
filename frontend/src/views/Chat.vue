<template>
  <div class="chat-page">
    <div class="chat-layout">
      <aside class="chat-sidebar">
        <div class="sidebar-header">
          <router-link to="/home" class="back-btn" aria-label="返回主页">←</router-link>
          <button class="new-btn" @click="createSession(true)" :disabled="loading">+ 新对话</button>
        </div>

        <div v-if="sessionError" class="sidebar-error">{{ sessionError }}</div>

        <div class="session-list">
          <button
            v-for="session in sessions"
            :key="session.id"
            class="session-item"
            :class="{ active: session.id === currentSessionId }"
            @click="selectSession(session.id)"
          >
            <div class="session-main">
              <p class="session-title">{{ session.title || '新对话' }}</p>
              <p class="session-time">{{ formatTime(session.last_message_at || session.updated_at || session.created_at) }}</p>
            </div>
            <div class="session-actions" @click.stop>
              <button class="mini-btn" @click="renameSession(session)">改名</button>
              <button class="mini-btn danger" @click="deleteSession(session)">删</button>
            </div>
          </button>
        </div>
      </aside>

      <section class="chat-panel">
        <header class="chat-header">
          <h1>{{ currentSessionTitle }}</h1>
          <p>与 AI 进行连续对话</p>
        </header>

        <div v-if="errorMessage" class="error-banner">{{ errorMessage }}</div>

        <main class="chat-messages" ref="msgArea">
          <div v-for="(msg, i) in msgs" :key="i" class="message" :class="msg.role">
            <div class="role-tag">{{ msg.role === 'user' ? '你' : 'AI' }}</div>
            <div class="message-bubble">{{ msg.content }}</div>
          </div>
          <div v-if="!msgs.length" class="empty-state">点击左侧“新对话”开始聊天</div>
        </main>

        <form class="chat-input" @submit.prevent="send">
          <input
            v-model="input"
            type="text"
            placeholder="输入消息，按 Enter 发送"
            :disabled="loading"
          />
          <button type="submit" :disabled="loading || !input.trim()">
            {{ loading ? '生成中' : '发送' }}
          </button>
        </form>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'

const SESSION_CACHE_KEY = 'chat_selected_session_id'

const sessions = ref([])
const currentSessionId = ref(null)
const msgs = ref([])
const input = ref('')
const loading = ref(false)
const errorMessage = ref('')
const sessionError = ref('')
const msgArea = ref(null)

const currentSessionTitle = computed(() => {
  const target = sessions.value.find((s) => s.id === currentSessionId.value)
  return target?.title || 'AI 聊天'
})

const getToken = () => sessionStorage.getItem('token') || ''

const withAuthHeaders = () => {
  const token = getToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

const parseErrorMessage = async (res, fallback) => {
  try {
    const body = await res.json()
    return body?.message || fallback
  } catch (_) {
    return fallback
  }
}

const scrollBottom = () => {
  nextTick(() => {
    if (msgArea.value) {
      msgArea.value.scrollTop = msgArea.value.scrollHeight
    }
  })
}

const formatTime = (value) => {
  if (!value) return ''
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const loadSessions = async () => {
  sessionError.value = ''
  const res = await fetch('/api/chat/sessions', {
    method: 'GET',
    headers: withAuthHeaders()
  })
  if (!res.ok) {
    throw new Error(await parseErrorMessage(res, '加载会话失败'))
  }
  const body = await res.json()
  sessions.value = body?.data?.sessions || []
}

const loadMessages = async (sessionId) => {
  const res = await fetch(`/api/chat/messages?session_id=${sessionId}&page=1&page_size=200`, {
    method: 'GET',
    headers: withAuthHeaders()
  })
  if (!res.ok) {
    throw new Error(await parseErrorMessage(res, '加载消息失败'))
  }
  const body = await res.json()
  msgs.value = (body?.data?.messages || []).map((m) => ({
    role: m.role,
    content: m.content
  }))
  scrollBottom()
}

const selectSession = async (sessionId) => {
  currentSessionId.value = sessionId
  localStorage.setItem(SESSION_CACHE_KEY, String(sessionId))
  errorMessage.value = ''
  try {
    await loadMessages(sessionId)
  } catch (err) {
    errorMessage.value = err.message || '加载会话消息失败'
  }
}

const createSession = async (autoSelect = true) => {
  sessionError.value = ''
  try {
    const res = await fetch('/api/chat/session/create', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...withAuthHeaders()
      },
      body: JSON.stringify({ title: '新对话' })
    })
    if (!res.ok) {
      throw new Error(await parseErrorMessage(res, '新建会话失败'))
    }
    const body = await res.json()
    const created = body?.data?.session
    await loadSessions()
    if (autoSelect && created?.id) {
      msgs.value = []
      await selectSession(created.id)
    }
  } catch (err) {
    sessionError.value = err.message || '新建会话失败'
  }
}

const renameSession = async (session) => {
  const title = window.prompt('输入新的会话名称', session.title || '新对话')
  if (!title || !title.trim()) return

  const res = await fetch('/api/chat/session/rename', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...withAuthHeaders()
    },
    body: JSON.stringify({ session_id: session.id, title: title.trim() })
  })
  if (!res.ok) {
    errorMessage.value = await parseErrorMessage(res, '重命名失败')
    return
  }
  await loadSessions()
}

const deleteSession = async (session) => {
  const ok = window.confirm(`确认删除会话「${session.title || '新对话'}」吗？`)
  if (!ok) return

  const res = await fetch('/api/chat/session/delete', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...withAuthHeaders()
    },
    body: JSON.stringify({ session_id: session.id })
  })
  if (!res.ok) {
    errorMessage.value = await parseErrorMessage(res, '删除会话失败')
    return
  }

  if (currentSessionId.value === session.id) {
    currentSessionId.value = null
    msgs.value = []
    localStorage.removeItem(SESSION_CACHE_KEY)
  }
  await loadSessions()
  if (!currentSessionId.value && sessions.value.length > 0) {
    await selectSession(sessions.value[0].id)
  }
}

const parseStreamEvent = (block) => {
  const dataLine = block.split('\n').find((line) => line.startsWith('data:'))
  if (!dataLine) return null
  try {
    return JSON.parse(dataLine.replace(/^data:\s*/, ''))
  } catch (_) {
    return null
  }
}

const send = async () => {
  if (!input.value.trim() || loading.value) return

  errorMessage.value = ''
  const userText = input.value.trim()
  input.value = ''
  loading.value = true

  msgs.value.push({ role: 'user', content: userText })
  const assistantIndex = msgs.value.push({ role: 'assistant', content: '' }) - 1
  scrollBottom()

  try {
    const token = getToken()
    const response = await fetch('/api/chat/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {})
      },
      body: JSON.stringify({
        session_id: currentSessionId.value,
        message: userText
      })
    })

    if (!response.ok) {
      throw new Error(await parseErrorMessage(response, '发送失败'))
    }
    if (!response.body) {
      throw new Error('后端未返回流数据')
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      let idx = buffer.indexOf('\n\n')
      while (idx !== -1) {
        const block = buffer.slice(0, idx)
        buffer = buffer.slice(idx + 2)
        idx = buffer.indexOf('\n\n')

        const payload = parseStreamEvent(block)
        if (!payload) continue

        if (payload.code !== 0) {
          throw new Error(payload.message || '流式响应失败')
        }

        if (payload.data?.session_id && !currentSessionId.value) {
          currentSessionId.value = payload.data.session_id
          localStorage.setItem(SESSION_CACHE_KEY, String(payload.data.session_id))
          await loadSessions()
        }

        if (payload.data?.type === 'chunk') {
          msgs.value[assistantIndex].content += payload.data.content || ''
          scrollBottom()
        }
      }
    }

    if (!msgs.value[assistantIndex].content) {
      msgs.value[assistantIndex].content = '未收到有效回复，请重试。'
    }
    await loadSessions()
  } catch (err) {
    const msg = err.message || '发送失败'
    errorMessage.value = msg
    if (!msgs.value[assistantIndex].content) {
      msgs.value[assistantIndex].content = `错误：${msg}`
    }
  } finally {
    loading.value = false
    scrollBottom()
  }
}

onMounted(async () => {
  try {
    await loadSessions()
    if (!sessions.value.length) return

    const cached = Number(localStorage.getItem(SESSION_CACHE_KEY))
    const target = sessions.value.find((s) => s.id === cached) || sessions.value[0]
    if (target?.id) {
      await selectSession(target.id)
    }
  } catch (err) {
    sessionError.value = err.message || '初始化会话失败'
  }
})
</script>

<style scoped>
.chat-page {
  position: fixed;
  inset: 0;
  padding: 16px;
  background:
    radial-gradient(circle at 15% 10%, rgba(249, 146, 86, 0.18) 0%, transparent 35%),
    radial-gradient(circle at 100% 90%, rgba(53, 126, 255, 0.12) 0%, transparent 35%),
    #0f1624;
}

.chat-layout {
  height: 100%;
  display: grid;
  grid-template-columns: 300px 1fr;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 18px;
  overflow: hidden;
  background: rgba(10, 16, 29, 0.92);
}

.chat-sidebar {
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  flex-direction: column;
  min-height: 0;
  background: rgba(15, 22, 36, 0.94);
}

.sidebar-header {
  padding: 14px;
  display: flex;
  gap: 10px;
}

.back-btn,
.new-btn {
  border: none;
  border-radius: 10px;
  padding: 10px 12px;
  color: #0f1624;
  background: linear-gradient(135deg, #ffb27b 0%, #ff8758 100%);
  font-weight: 700;
  text-decoration: none;
  cursor: pointer;
}

.new-btn {
  flex: 1;
}

.sidebar-error {
  margin: 0 14px 12px;
  padding: 10px;
  border-radius: 10px;
  color: #ffc9c9;
  font-size: 12px;
  background: rgba(156, 40, 40, 0.3);
  border: 1px solid rgba(255, 120, 120, 0.45);
}

.session-list {
  overflow-y: auto;
  padding: 0 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.session-item {
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.02);
  padding: 10px;
  color: #edf4ff;
  text-align: left;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  gap: 8px;
}

.session-item.active {
  border-color: rgba(255, 175, 122, 0.85);
  background: rgba(255, 175, 122, 0.1);
}

.session-main {
  min-width: 0;
}

.session-title {
  margin: 0;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-time {
  margin: 4px 0 0;
  font-size: 11px;
  color: rgba(237, 244, 255, 0.6);
}

.session-actions {
  display: flex;
  gap: 6px;
}

.mini-btn {
  border: 1px solid rgba(255, 255, 255, 0.22);
  border-radius: 8px;
  background: transparent;
  color: #dfe9fb;
  font-size: 11px;
  padding: 2px 6px;
  cursor: pointer;
}

.mini-btn.danger {
  color: #ffc9c9;
  border-color: rgba(255, 115, 115, 0.45);
}

.chat-panel {
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
  overflow: hidden;
}

.chat-header {
  padding: 18px 22px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.chat-header h1 {
  margin: 0;
  color: #f6fbff;
  font-size: 20px;
  font-family: 'Outfit', sans-serif;
}

.chat-header p {
  margin: 4px 0 0;
  color: rgba(246, 251, 255, 0.7);
  font-size: 13px;
}

.error-banner {
  margin: 12px 22px 0;
  padding: 10px 12px;
  border-radius: 10px;
  color: #ffd4d4;
  border: 1px solid rgba(255, 99, 99, 0.55);
  background: rgba(102, 23, 23, 0.45);
}

.chat-messages {
  padding: 16px 22px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  flex: 1;
}

.empty-state {
  margin: auto;
  color: rgba(237, 244, 255, 0.55);
  font-size: 14px;
}

.message {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-width: 78%;
}

.message.user {
  margin-left: auto;
  align-items: flex-end;
}

.message.assistant {
  margin-right: auto;
  align-items: flex-start;
}

.role-tag {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.64);
}

.message-bubble {
  padding: 11px 13px;
  border-radius: 12px;
  line-height: 1.52;
  font-size: 14px;
  white-space: pre-wrap;
  word-break: break-word;
}

.message.user .message-bubble {
  color: #192237;
  background: linear-gradient(135deg, #ffc497 0%, #ff9f6e 100%);
}

.message.assistant .message-bubble {
  color: #edf4ff;
  border: 1px solid rgba(126, 167, 255, 0.34);
  background: rgba(40, 57, 97, 0.72);
}

.chat-input {
  padding: 14px 18px 18px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  display: grid;
  grid-template-columns: minmax(0, 1fr) 96px;
  gap: 10px;
  flex-shrink: 0;
  align-items: center;
}

.chat-input input {
  border: 1px solid rgba(136, 179, 255, 0.45);
  border-radius: 10px;
  background: rgba(11, 17, 31, 0.95);
  color: #edf4ff;
  padding: 12px 13px;
  height: 46px;
  width: 100%;
  min-width: 0;
}

.chat-input button {
  border: none;
  border-radius: 10px;
  height: 46px;
  padding: 0;
  color: #0f1624;
  font-weight: 700;
  background: linear-gradient(135deg, #8cc8ff 0%, #ff9f6e 100%);
  cursor: pointer;
  width: 96px;
  min-width: 96px;
}

.chat-input button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

@media (max-width: 900px) {
  .chat-page {
    padding: 0;
  }

  .chat-layout {
    border-radius: 0;
    border: none;
    grid-template-columns: 1fr;
    grid-template-rows: 220px 1fr;
  }

  .chat-sidebar {
    border-right: none;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }
}
</style>

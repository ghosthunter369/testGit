<template>
  <div class="glass-card home-card">
    <div class="header-row">
      <div class="welcome-section">
        <h2>欢迎回来</h2>
        <p class="subtitle">{{ user?.username || '用户' }}</p>
      </div>
      <div class="header-buttons">
        <button @click="goToChat" class="chat-btn">AI 聊天</button>
        <button @click="logout" class="logout-btn">退出登录</button>
      </div>
    </div>

    <div v-if="errorMessage" class="error">{{ errorMessage }}</div>

    <div class="user-profile">
      <div class="avatar">
        {{ (user?.username || '?').charAt(0).toUpperCase() }}
      </div>
      <div class="user-details">
        <h3>{{ user?.username || '未登录' }}</h3>
        <p class="user-email">{{ user?.email || '未设置邮箱' }}</p>
      </div>
    </div>

    <div class="info-grid" v-if="userInfo.id">
      <div class="info-card">
        <span class="info-label">用户ID</span>
        <span class="info-value">#{{ userInfo.id }}</span>
      </div>
      <div class="info-card">
        <span class="info-label">用户名</span>
        <span class="info-value">{{ userInfo.username }}</span>
      </div>
      <div class="info-card full-width">
        <span class="info-label">邮箱</span>
        <span class="info-value">{{ userInfo.email || '未设置' }}</span>
      </div>
      <div class="info-card full-width">
        <span class="info-label">注册时间</span>
        <span class="info-value">{{ formatDate(userInfo.created_at) }}</span>
      </div>
    </div>

    <div class="status-bar">
      <div class="status-dot"></div>
      <span>已连接</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const router = useRouter()
const errorMessage = ref('')

const parseUser = () => {
  const raw = sessionStorage.getItem('user')
  if (!raw) return {}
  try {
    return JSON.parse(raw)
  } catch (_) {
    return {}
  }
}

const user = ref(parseUser())
const userInfo = ref({})

const authHeaders = () => {
  const token = sessionStorage.getItem('token') || ''
  return token ? { Authorization: `Bearer ${token}` } : {}
}

const fetchUserInfo = async () => {
  errorMessage.value = ''
  try {
    const res = await axios.get('/api/user', { headers: authHeaders() })
    userInfo.value = res.data?.data || {}
    user.value = userInfo.value
    sessionStorage.setItem('user', JSON.stringify(userInfo.value))
  } catch (err) {
    errorMessage.value = err.response?.data?.message || err.message || '获取用户信息失败'
  }
}

const logout = async () => {
  errorMessage.value = ''
  try {
    await axios.post('/api/logout', {}, { headers: authHeaders() })
  } catch (err) {
    errorMessage.value = err.response?.data?.message || err.message || '退出失败'
  }

  sessionStorage.removeItem('token')
  sessionStorage.removeItem('user')
  localStorage.removeItem('chat_selected_session_id')
  router.push('/login')
}

const goToChat = () => {
  router.push('/chat')
}

const formatDate = (date) => {
  if (!date) return '未知'
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

onMounted(() => {
  fetchUserInfo()
})
</script>

<style scoped>
.home-card {
  max-width: 480px;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 32px;
}

.header-buttons {
  display: flex;
  gap: 10px;
}

.welcome-section {
  text-align: left;
}

.welcome-section h2 {
  text-align: left;
  margin-bottom: 4px;
}

.subtitle {
  text-align: left;
  color: var(--color-primary);
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 0;
}

.logout-btn {
  width: auto;
  padding: 10px 20px;
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #fca5a5;
  font-size: 13px;
  opacity: 1;
  animation: none;
}

.chat-btn {
  width: auto;
  padding: 10px 20px;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-accent) 100%);
  border: none;
  color: var(--color-bg-dark);
  font-size: 13px;
  font-weight: 600;
}

.chat-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(99, 102, 241, 0.4);
}

.logout-btn:hover {
  background: rgba(239, 68, 68, 0.25);
  box-shadow: none;
  transform: none;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px;
  background: rgba(0, 0, 0, 0.2);
  border-radius: var(--radius-md);
  margin-bottom: 28px;
}

.avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-accent) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: 'Outfit', sans-serif;
  font-size: 28px;
  font-weight: 600;
  color: var(--color-bg-dark);
  flex-shrink: 0;
}

.user-details h3 {
  font-family: 'Outfit', sans-serif;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 4px;
  text-align: left;
}

.user-email {
  color: var(--color-text-muted);
  font-size: 14px;
  text-align: left;
  margin: 0;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 24px;
}

.info-card {
  background: rgba(0, 0, 0, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: var(--radius-sm);
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.info-card.full-width {
  grid-column: 1 / -1;
}

.info-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-muted);
}

.info-value {
  font-family: 'Outfit', sans-serif;
  font-size: 15px;
  font-weight: 500;
  color: var(--color-text);
}

.status-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  color: var(--color-text-muted);
  font-size: 12px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 8px rgba(34, 197, 94, 0.5);
}
</style>


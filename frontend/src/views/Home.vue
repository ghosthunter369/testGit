<template>
  <div class="home-page">
    <div class="header">
      <div class="user-info">
        <span class="greeting">{{ greeting }}</span>
        <span class="username">{{ username }}</span>
      </div>
      <button class="logout-btn" @click="logout">退出登录</button>
    </div>

    <div class="feature-grid">
      <div class="feature-card" @click="$router.push('/chat')">
        <span class="feature-icon">&#129302;</span>
        <span class="feature-name">AI 对话</span>
        <span class="feature-desc">智能助手随时聊</span>
      </div>
      <div class="feature-card" @click="$router.push('/bills')">
        <span class="feature-icon">&#128178;</span>
        <span class="feature-name">记账本</span>
        <span class="feature-desc">拍照识别自动记账</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const router = useRouter()
const headers = computed(() => {
  const token = sessionStorage.getItem('token')
  return token ? { Authorization: 'Bearer ' + token } : {}
})

const username = ref('用户')
const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '早上好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

function logout() {
  axios.post('/api/logout', {}, { headers: headers.value }).finally(() => {
    sessionStorage.removeItem('token')
    sessionStorage.removeItem('user')
    router.push('/login')
  })
}

onMounted(() => {
  try {
    const user = JSON.parse(sessionStorage.getItem('user') || '{}')
    if (user.username) username.value = user.username
  } catch (_) {}
})
</script>

<style scoped>
.home-page {
  max-width: 480px;
  margin: 0 auto;
  padding: 16px;
  min-height: 100vh;
  background: #f5f5f5;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}
.user-info {
  display: flex;
  flex-direction: column;
}
.greeting { font-size: 14px; color: #999; }
.username { font-size: 20px; font-weight: 700; color: #333; }
.logout-btn {
  background: none;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 6px 14px;
  color: #666;
  font-size: 13px;
  cursor: pointer;
}
.feature-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.feature-card {
  background: #fff;
  border-radius: 14px;
  padding: 24px 16px;
  text-align: center;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
  transition: transform 0.15s, box-shadow 0.15s;
}
.feature-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.12);
}
.feature-icon { font-size: 36px; display: block; margin-bottom: 8px; }
.feature-name { display: block; font-size: 15px; font-weight: 600; color: #333; margin-bottom: 4px; }
.feature-desc { display: block; font-size: 12px; color: #aaa; }
</style>

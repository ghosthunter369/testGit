<template>
  <div class="glass-card">
    <h2>创建账号</h2>
    <p class="subtitle">加入我们，开始使用</p>

    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="success" class="success">{{ success }}</div>

    <form @submit.prevent="handleRegister">
      <div class="form-group">
        <label>用户名</label>
        <input v-model="username" type="text" placeholder="请输入用户名" required />
      </div>
      <div class="form-group">
        <label>密码</label>
        <input v-model="password" type="password" placeholder="请输入密码" required />
      </div>
      <div class="form-group">
        <label>邮箱</label>
        <input v-model="email" type="email" placeholder="your@email.com" />
      </div>
      <button type="submit">注册</button>
    </form>

    <div class="link">
      <router-link to="/login">已有账号？立即登录</router-link>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const router = useRouter()
const username = ref('')
const password = ref('')
const email = ref('')
const error = ref('')
const success = ref('')

const handleRegister = async () => {
  error.value = ''
  success.value = ''

  try {
    const res = await axios.post('/api/register', {
      username: username.value,
      password: password.value,
      email: email.value
    })

    success.value = res.data?.message || '注册成功，正在跳转...'
    setTimeout(() => {
      router.push('/login')
    }, 1200)
  } catch (err) {
    error.value = err.response?.data?.message || err.message || '注册失败，请重试'
  }
}
</script>


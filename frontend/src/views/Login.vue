<template>
  <div class="glass-card">
    <h2>欢迎回来</h2>
    <p class="subtitle">登录到您的账户</p>

    <div v-if="error" class="error">{{ error }}</div>

    <form @submit.prevent="handleLogin">
      <div class="form-group">
        <label>用户名</label>
        <input v-model="username" type="text" placeholder="请输入用户名" required />
      </div>

      <div class="form-group">
        <label>密码</label>
        <input v-model="password" type="password" placeholder="请输入密码" required />
      </div>

      <div class="remember-row">
        <el-checkbox v-model="rememberMe" class="remember-checkbox">30天免登录</el-checkbox>
      </div>

      <button type="submit">登录</button>
    </form>

    <div class="link">
      <router-link to="/register">还没有账号？立即注册</router-link>
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
const rememberMe = ref(false)
const error = ref('')

const handleLogin = async () => {
  error.value = ''
  try {
    const res = await axios.post('/api/login', {
      username: username.value,
      password: password.value,
      remember_me: rememberMe.value
    })

    const payload = res.data?.data || {}
    if (!payload?.token) {
      throw new Error(res.data?.message || '登录失败，请重试')
    }

    sessionStorage.setItem('token', payload.token)
    sessionStorage.setItem('user', JSON.stringify(payload.user || {}))

    router.push('/home')
  } catch (err) {
    error.value = err.response?.data?.message || err.message || '登录失败，请重试'
  }
}
</script>

<style scoped>
.remember-row {
  margin-top: 6px;
  margin-bottom: 10px;
  min-height: 28px;
  display: flex;
  align-items: center;
}

:deep(.remember-checkbox.el-checkbox) {
  --el-checkbox-checked-input-border-color: var(--color-primary);
  --el-checkbox-checked-bg-color: var(--color-primary);
  --el-checkbox-input-border-color: rgba(255, 255, 255, 0.35);
  --el-checkbox-input-bg-color: rgba(255, 255, 255, 0.02);
  --el-checkbox-text-color: rgba(248, 249, 250, 0.88);
  --el-checkbox-checked-text-color: #f8f9fa;
}

:deep(.remember-checkbox .el-checkbox__label) {
  font-size: 14px;
  font-weight: 500;
  padding-left: 8px;
}
</style>


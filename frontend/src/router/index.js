import { createRouter, createWebHistory } from 'vue-router'
import axios from 'axios'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import Home from '../views/Home.vue'
import Chat from '../views/Chat.vue'

const routes = [
  { path: '/', redirect: '/login' },
  { path: '/login', component: Login },
  { path: '/register', component: Register },
  { path: '/home', component: Home, meta: { requiresAuth: true } },
  { path: '/chat', component: Chat, meta: { requiresAuth: true } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const checkAuth = async () => {
  const token = sessionStorage.getItem('token')
  const headers = token ? { Authorization: `Bearer ${token}` } : {}
  try {
    const res = await axios.get('/api/user', { headers })
    if (res.data?.data) {
      sessionStorage.setItem('user', JSON.stringify(res.data.data))
    }
    return true
  } catch (_) {
    sessionStorage.removeItem('token')
    sessionStorage.removeItem('user')
    return false
  }
}

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) return true
  const authed = await checkAuth()
  if (!authed) return '/login'
  return true
})

export default router


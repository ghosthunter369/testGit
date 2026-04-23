<template>
  <div class="stats-page">
    <div class="header">
      <button class="back-btn" @click="$router.push('/bills')">&#8592; 返回</button>
      <span class="title">月度统计</span>
      <div style="width:60px"></div>
    </div>

    <div class="month-selector">
      <button @click="prevMonth">&#8249;</button>
      <span>{{ year }}年{{ month }}月</span>
      <button @click="nextMonth">&#8250;</button>
    </div>

    <div class="overview">
      <div class="card">
        <div class="card-label">总支出</div>
        <div class="card-value expense">{{ stats.total_expense.toFixed(2) }}</div>
      </div>
      <div class="card">
        <div class="card-label">总收入</div>
        <div class="card-value income">{{ stats.total_income.toFixed(2) }}</div>
      </div>
      <div class="card">
        <div class="card-label">结余</div>
        <div class="card-value" :class="stats.balance >= 0 ? 'income' : 'expense'">
          {{ stats.balance.toFixed(2) }}
        </div>
      </div>
    </div>

    <div class="section">
      <h3>每日支出趋势</h3>
      <div class="bar-chart">
        <div v-for="d in dailyStats" :key="d.date" class="bar-row">
          <span class="bar-label">{{ d.date.slice(5) }}</span>
          <div class="bar-track">
            <div
              class="bar-fill expense"
              :style="{ width: barWidth(d.expense) + '%' }"
              :title="'支出 ' + d.expense.toFixed(2)"
            ></div>
            <div
              v-if="d.income > 0"
              class="bar-fill income"
              :style="{ width: barWidth(d.income) + '%' }"
              :title="'收入 ' + d.income.toFixed(2)"
            ></div>
          </div>
          <span class="bar-value">{{ d.expense.toFixed(0) }}</span>
        </div>
      </div>
    </div>

    <div class="section">
      <h3>分类支出排行</h3>
      <div v-for="c in categoryStats" :key="c.category_id" class="rank-item">
        <span class="rank-icon">{{ c.category_icon }}</span>
        <span class="rank-name">{{ c.category_name }}</span>
        <div class="rank-bar-track">
          <div class="rank-bar-fill" :style="{ width: rankWidth(c.amount) + '%' }"></div>
        </div>
        <span class="rank-amount">{{ c.amount.toFixed(2) }}</span>
        <span class="rank-count">{{ c.count }}笔</span>
      </div>
      <div v-if="categoryStats.length === 0" class="empty">暂无数据</div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'

const headers = computed(() => {
  const token = sessionStorage.getItem('token')
  return token ? { Authorization: 'Bearer ' + token } : {}
})

const now = new Date()
const year = ref(now.getFullYear())
const month = ref(now.getMonth() + 1)

const stats = ref({ total_expense: 0, total_income: 0, balance: 0 })
const dailyStats = ref([])
const categoryStats = ref([])

const maxExpense = computed(() => {
  if (dailyStats.value.length === 0) return 1
  return Math.max(...dailyStats.value.map(d => d.expense), 1)
})

const maxCategoryAmount = computed(() => {
  if (categoryStats.value.length === 0) return 1
  return Math.max(...categoryStats.value.map(c => c.amount), 1)
})

function barWidth(val) {
  return Math.max((val / maxExpense.value) * 100, val > 0 ? 3 : 0)
}

function rankWidth(val) {
  return Math.max((val / maxCategoryAmount.value) * 100, val > 0 ? 5 : 0)
}

function prevMonth() {
  if (month.value === 1) { month.value = 12; year.value-- }
  else { month.value-- }
  loadStats()
}

function nextMonth() {
  if (month.value === 12) { month.value = 1; year.value++ }
  else { month.value++ }
  loadStats()
}

async function loadStats() {
  try {
    const res = await axios.get('/api/bill/stats/monthly', {
      headers: headers.value,
      params: { year: year.value, month: month.value }
    })
    if (res.data?.code === 0) {
      const d = res.data.data
      stats.value = {
        total_expense: d.total_expense,
        total_income: d.total_income,
        balance: d.balance
      }
      dailyStats.value = d.daily_stats || []
      categoryStats.value = d.category_stats || []
    }
  } catch (_) {}
}

onMounted(() => { loadStats() })
</script>

<style scoped>
.stats-page {
  max-width: 480px;
  margin: 0 auto;
  padding: 16px;
  min-height: 100vh;
  background: #f5f5f5;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}
.header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px;
}
.back-btn {
  background: #fff; border: 1px solid #ddd; border-radius: 8px;
  padding: 6px 12px; cursor: pointer; font-size: 13px;
}
.title { font-size: 16px; font-weight: 600; }
.month-selector {
  display: flex; align-items: center; justify-content: center; gap: 16px;
  margin-bottom: 16px; font-size: 15px; font-weight: 600;
}
.month-selector button {
  background: #fff; border: 1px solid #ddd; border-radius: 8px;
  padding: 6px 12px; cursor: pointer; font-size: 18px;
}
.overview {
  display: flex; gap: 8px; margin-bottom: 20px;
}
.card {
  flex: 1; background: #fff; border-radius: 10px; padding: 12px 8px;
  text-align: center; box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.card-label { font-size: 11px; color: #999; margin-bottom: 4px; }
.card-value { font-size: 16px; font-weight: 700; }
.card-value.expense { color: #e74c3c; }
.card-value.income { color: #27ae60; }
.section {
  background: #fff; border-radius: 12px; padding: 16px;
  margin-bottom: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.section h3 { font-size: 14px; margin: 0 0 12px 0; color: #333; }
.bar-chart { display: flex; flex-direction: column; gap: 4px; }
.bar-row { display: flex; align-items: center; gap: 6px; }
.bar-label { font-size: 11px; color: #999; width: 32px; text-align: right; flex-shrink: 0; }
.bar-track {
  flex: 1; height: 14px; background: #f5f5f5; border-radius: 3px;
  display: flex; gap: 1px; overflow: hidden;
}
.bar-fill { height: 100%; border-radius: 3px; min-width: 0; transition: width 0.3s; }
.bar-fill.expense { background: #e74c3c; }
.bar-fill.income { background: #27ae60; }
.bar-value { font-size: 10px; color: #999; width: 36px; flex-shrink: 0; }
.rank-item {
  display: flex; align-items: center; gap: 8px; margin-bottom: 10px;
}
.rank-icon { font-size: 18px; width: 24px; text-align: center; }
.rank-name { font-size: 13px; color: #333; width: 60px; flex-shrink: 0; }
.rank-bar-track {
  flex: 1; height: 10px; background: #f0f0f0; border-radius: 3px; overflow: hidden;
}
.rank-bar-fill {
  height: 100%; background: #4a90d9; border-radius: 3px; transition: width 0.3s;
}
.rank-amount { font-size: 13px; font-weight: 600; color: #333; width: 56px; text-align: right; flex-shrink: 0; }
.rank-count { font-size: 11px; color: #999; width: 28px; text-align: right; flex-shrink: 0; }
.empty { padding: 20px; text-align: center; color: #ccc; font-size: 13px; }
</style>

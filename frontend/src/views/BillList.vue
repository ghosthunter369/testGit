<template>
  <div class="bill-page">
    <div class="header">
      <button class="back-btn" @click="$router.push('/home')">&#8592;</button>
      <div class="month-selector">
        <button @click="prevMonth">&#8249;</button>
        <span>{{ year }}年{{ month }}月</span>
        <button @click="nextMonth">&#8250;</button>
      </div>
      <button class="stats-btn" @click="$router.push('/bills/stats')">统计</button>
    </div>

    <div class="summary-bar">
      <div class="summary-item">
        <span class="label">总支出</span>
        <span class="amount expense">{{ summary.expense.toFixed(2) }}</span>
      </div>
      <div class="divider"></div>
      <div class="summary-item">
        <span class="label">总收入</span>
        <span class="amount income">{{ summary.income.toFixed(2) }}</span>
      </div>
    </div>

    <div class="filter-bar">
      <select v-model="filterType" @change="loadBills">
        <option value="-1">全部</option>
        <option value="0">支出</option>
        <option value="1">收入</option>
      </select>
      <select v-model="filterCategory" @change="loadBills">
        <option value="0">全部分类</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">
          {{ cat.icon }} {{ cat.name }}
        </option>
      </select>
    </div>

    <div class="add-btn-wrapper">
      <button class="add-btn" @click="$router.push('/bills/add')">+ 记一笔</button>
    </div>

    <div class="bill-list">
      <div v-for="group in groupedBills" :key="group.date" class="date-group">
        <div class="date-title">{{ group.date }}</div>
        <div v-for="bill in group.items" :key="bill.id" class="bill-item" @click="goEdit(bill.id)">
          <span class="bill-icon">{{ bill.category_icon || '?' }}</span>
          <div class="bill-info">
            <div class="bill-merchant">{{ bill.merchant || bill.description || '未命名' }}</div>
            <div class="bill-desc">{{ bill.category_name }}</div>
          </div>
          <span :class="['bill-amount', bill.type === 1 ? 'income' : 'expense']">
            {{ bill.type === 1 ? '+' : '-' }}{{ bill.amount.toFixed(2) }}
          </span>
        </div>
      </div>
      <div v-if="bills.length === 0 && !loading" class="empty">
        暂无记录，点击上方按钮记一笔
      </div>
      <div v-if="loading" class="loading">加载中...</div>
    </div>

    <div v-if="page * size < total" class="load-more">
      <button @click="loadMore">加载更多</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const router = useRouter()
const headers = computed(() => {
  const token = sessionStorage.getItem('token')
  return token ? { Authorization: 'Bearer ' + token } : {}
})

const now = new Date()
const year = ref(now.getFullYear())
const month = ref(now.getMonth() + 1)
const page = ref(1)
const size = 20
const total = ref(0)
const bills = ref([])
const loading = ref(false)
const categories = ref([])
const filterType = ref('-1')
const filterCategory = ref('0')

const summary = ref({ expense: 0, income: 0 })

const groupedBills = computed(() => {
  const map = {}
  bills.value.forEach(b => {
    if (!map[b.bill_date]) map[b.bill_date] = []
    map[b.bill_date].push(b)
  })
  return Object.keys(map).sort((a, b) => b.localeCompare(a)).map(date => ({ date, items: map[date] }))
})

function prevMonth() {
  if (month.value === 1) { month.value = 12; year.value-- }
  else { month.value-- }
  page.value = 1
  bills.value = []
  loadBills()
}

function nextMonth() {
  if (month.value === 12) { month.value = 1; year.value++ }
  else { month.value++ }
  page.value = 1
  bills.value = []
  loadBills()
}

function goEdit(id) {
  router.push('/bills/edit/' + id)
}

async function loadCategories() {
  try {
    const res = await axios.get('/api/bill/categories', { headers: headers.value })
    if (res.data?.code === 0) categories.value = res.data.data
  } catch (_) {}
}

async function loadBills() {
  loading.value = true
  try {
    const params = {
      page: page.value,
      size,
      year: year.value,
      month: month.value
    }
    if (filterType.value !== '-1') params.type = filterType.value
    if (filterCategory.value !== '0') params.category_id = filterCategory.value
    const res = await axios.get('/api/bill/list', { headers: headers.value, params })
    if (res.data?.code === 0) {
      if (page.value === 1) {
        bills.value = res.data.data.list
      } else {
        bills.value.push(...res.data.data.list)
      }
      total.value = res.data.data.total
      summary.value = res.data.data.summary
    }
  } catch (_) {}
  loading.value = false
}

function loadMore() {
  page.value++
  loadBills()
}

onMounted(() => {
  loadCategories()
  loadBills()
})
</script>

<style scoped>
.bill-page {
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
  margin-bottom: 12px;
}
.back-btn, .stats-btn {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 14px;
}
.month-selector {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 16px;
  font-weight: 600;
}
.month-selector button {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #333;
}
.summary-bar {
  display: flex;
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.summary-item {
  flex: 1;
  text-align: center;
}
.summary-item .label {
  display: block;
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}
.summary-item .amount {
  font-size: 18px;
  font-weight: 700;
}
.summary-item .amount.expense { color: #e74c3c; }
.summary-item .amount.income { color: #27ae60; }
.divider {
  width: 1px;
  background: #eee;
}
.filter-bar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.filter-bar select {
  flex: 1;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: #fff;
  font-size: 13px;
}
.add-btn-wrapper {
  margin-bottom: 12px;
}
.add-btn {
  width: 100%;
  padding: 12px;
  background: #4a90d9;
  color: #fff;
  border: none;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
}
.add-btn:hover { background: #357abd; }
.bill-list {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.date-group {}
.date-title {
  padding: 8px 16px;
  font-size: 12px;
  color: #999;
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
}
.bill-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
}
.bill-item:last-child { border-bottom: none; }
.bill-item:hover { background: #f9f9f9; }
.bill-icon {
  font-size: 22px;
  margin-right: 12px;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f4f8;
  border-radius: 50%;
}
.bill-info {
  flex: 1;
  min-width: 0;
}
.bill-merchant {
  font-size: 14px;
  color: #333;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.bill-desc {
  font-size: 12px;
  color: #aaa;
  margin-top: 2px;
}
.bill-amount {
  font-size: 15px;
  font-weight: 600;
  white-space: nowrap;
}
.bill-amount.expense { color: #e74c3c; }
.bill-amount.income { color: #27ae60; }
.empty {
  padding: 40px 16px;
  text-align: center;
  color: #ccc;
  font-size: 14px;
}
.loading {
  padding: 20px;
  text-align: center;
  color: #999;
  font-size: 13px;
}
.load-more {
  text-align: center;
  padding: 16px;
}
.load-more button {
  background: none;
  border: 1px solid #ddd;
  border-radius: 20px;
  padding: 8px 24px;
  color: #666;
  cursor: pointer;
  font-size: 13px;
}
</style>

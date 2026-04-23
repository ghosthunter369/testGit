<template>
  <div class="bill-edit-page">
    <div class="header">
      <button class="back-btn" @click="$router.back()">&#8592; 返回</button>
      <span class="title">编辑账单</span>
      <button class="del-btn" @click="remove">删除</button>
    </div>

    <div class="type-tabs">
      <button :class="['tab', form.type === 0 ? 'active expense' : '']" @click="form.type = 0; loadCategories()">支出</button>
      <button :class="['tab', form.type === 1 ? 'active income' : '']" @click="form.type = 1; loadCategories()">收入</button>
    </div>

    <div class="form-group">
      <label>金额</label>
      <input v-model="amountDisplay" type="text" placeholder="0.00" class="amount-input" @input="onAmountInput">
    </div>

    <div class="form-group">
      <label>分类</label>
      <div class="category-grid">
        <div
          v-for="cat in filteredCategories"
          :key="cat.id"
          :class="['category-item', form.category_id === cat.id ? 'selected' : '']"
          @click="form.category_id = cat.id"
        >
          <span class="cat-icon">{{ cat.icon }}</span>
          <span class="cat-name">{{ cat.name }}</span>
        </div>
      </div>
    </div>

    <div class="form-group">
      <label>商户</label>
      <input v-model="form.merchant" placeholder="商户名称">
    </div>

    <div class="form-group">
      <label>日期</label>
      <input v-model="form.bill_date" type="date">
    </div>

    <div class="form-group">
      <label>备注</label>
      <input v-model="form.description" placeholder="消费描述">
    </div>

    <button class="save-btn" @click="save" :disabled="saving">
      {{ saving ? '保存中...' : '保存修改' }}
    </button>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from 'axios'

const router = useRouter()
const route = useRoute()
const headers = computed(() => {
  const token = sessionStorage.getItem('token')
  return token ? { Authorization: 'Bearer ' + token } : {}
})

const saving = ref(false)
const categories = ref([])
const amountDisplay = ref('')

const form = ref({
  category_id: 0,
  amount: 0,
  type: 0,
  merchant: '',
  description: '',
  bill_date: '',
  image_url: ''
})

const filteredCategories = computed(() => {
  return categories.value.filter(c => c.type === form.value.type)
})

function onAmountInput(e) {
  let v = e.target.value.replace(/[^0-9.]/g, '')
  if (v.split('.').length > 2) v = v.slice(0, v.lastIndexOf('.'))
  amountDisplay.value = v
  form.value.amount = parseFloat(v) || 0
}

async function loadCategories() {
  try {
    const params = {}
    if (form.value.type === 0) params.type = 0
    else params.type = 1
    const res = await axios.get('/api/bill/categories', { headers: headers.value, params })
    if (res.data?.code === 0) categories.value = res.data.data
  } catch (_) {}
}

async function loadBill() {
  try {
    const res = await axios.get('/api/bill/detail/' + route.params.id, { headers: headers.value })
    if (res.data?.code === 0) {
      const d = res.data.data
      form.value = {
        category_id: d.category_id,
        amount: d.amount,
        type: d.type,
        merchant: d.merchant || '',
        description: d.description || '',
        bill_date: d.bill_date,
        image_url: d.image_url || ''
      }
      amountDisplay.value = String(d.amount)
    }
  } catch (err) {
    alert('加载失败')
    router.back()
  }
}

async function save() {
  if (form.value.amount <= 0) { alert('请输入金额'); return }
  if (!form.value.bill_date) { alert('请选择日期'); return }
  saving.value = true
  try {
    const res = await axios.put('/api/bill/update/' + route.params.id, form.value, { headers: headers.value })
    if (res.data?.code === 0) {
      router.back()
    } else {
      alert(res.data?.message || '保存失败')
    }
  } catch (err) {
    alert('保存失败')
  }
  saving.value = false
}

async function remove() {
  if (!confirm('确定要删除这条账单吗？')) return
  try {
    const res = await axios.delete('/api/bill/delete/' + route.params.id, { headers: headers.value })
    if (res.data?.code === 0) {
      router.back()
    } else {
      alert(res.data?.message || '删除失败')
    }
  } catch (err) {
    alert('删除失败')
  }
}

onMounted(() => {
  loadBill()
  loadCategories()
})
</script>

<style scoped>
.bill-edit-page {
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
  margin-bottom: 16px;
}
.back-btn {
  background: #fff;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 13px;
}
.title { font-size: 16px; font-weight: 600; }
.del-btn {
  background: none;
  border: none;
  color: #e74c3c;
  font-size: 13px;
  cursor: pointer;
}
.type-tabs { display: flex; gap: 8px; margin-bottom: 16px; }
.tab {
  flex: 1; padding: 10px; border: 2px solid #eee; border-radius: 8px;
  background: #fff; font-size: 14px; font-weight: 600; cursor: pointer;
}
.tab.active.expense { border-color: #e74c3c; color: #e74c3c; background: #fef0ef; }
.tab.active.income { border-color: #27ae60; color: #27ae60; background: #eefaf2; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: 13px; color: #666; margin-bottom: 6px; }
.form-group input {
  width: 100%; padding: 10px 12px; border: 1px solid #ddd;
  border-radius: 8px; font-size: 14px; box-sizing: border-box; background: #fff;
}
.amount-input { font-size: 24px !important; font-weight: 700; text-align: center; }
.category-grid { display: flex; flex-wrap: wrap; gap: 8px; }
.category-item {
  display: flex; flex-direction: column; align-items: center; width: 60px;
  padding: 8px 4px; border: 2px solid #eee; border-radius: 10px;
  cursor: pointer; background: #fff;
}
.category-item.selected { border-color: #4a90d9; background: #eef3fa; }
.cat-icon { font-size: 20px; }
.cat-name { font-size: 10px; color: #666; margin-top: 2px; }
.save-btn {
  width: 100%; padding: 14px; background: #4a90d9; color: #fff;
  border: none; border-radius: 10px; font-size: 16px; font-weight: 600;
  cursor: pointer; margin-top: 8px;
}
.save-btn:hover { background: #357abd; }
.save-btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>

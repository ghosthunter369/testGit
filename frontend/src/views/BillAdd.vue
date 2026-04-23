<template>
  <div class="bill-add-page">
    <div class="header">
      <button class="back-btn" @click="$router.back()">&#8592; 返回</button>
      <span class="title">记一笔</span>
      <div style="width:60px"></div>
    </div>

    <div class="upload-area" @click="triggerUpload">
      <input ref="fileInput" type="file" accept="image/*" style="display:none" @change="onFileChange">
      <div v-if="previewUrl" class="preview">
        <img :src="previewUrl" alt="preview">
        <span class="recognizing" v-if="recognizing">AI 识别中...</span>
      </div>
      <div v-else class="upload-placeholder">
        <span class="upload-icon">&#128247;</span>
        <span>点击上传小票 / 账单截图</span>
      </div>
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
      <input v-model="form.merchant" placeholder="商户名称（选填）">
    </div>

    <div class="form-group">
      <label>日期</label>
      <input v-model="form.bill_date" type="date">
    </div>

    <div class="form-group">
      <label>备注</label>
      <input v-model="form.description" placeholder="消费描述（选填）">
    </div>

    <button class="save-btn" @click="save" :disabled="saving">
      {{ saving ? '保存中...' : '保存' }}
    </button>
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

const fileInput = ref(null)
const previewUrl = ref('')
const recognizing = ref(false)
const saving = ref(false)
const categories = ref([])
const amountDisplay = ref('')

const form = ref({
  category_id: 0,
  amount: 0,
  type: 0,
  merchant: '',
  description: '',
  bill_date: new Date().toISOString().slice(0, 10),
  image_url: ''
})

const filteredCategories = computed(() => {
  return categories.value.filter(c => c.type === form.value.type)
})

function triggerUpload() {
  fileInput.value.click()
}

async function onFileChange(e) {
  const file = e.target.files[0]
  if (!file) return

  previewUrl.value = URL.createObjectURL(file)
  recognizing.value = true

  try {
    const fd = new FormData()
    fd.append('image', file)
    const uploadRes = await axios.post('/api/bill/upload', fd, { headers: headers.value })
    if (uploadRes.data?.code !== 0) {
      alert(uploadRes.data?.message || '上传失败')
      recognizing.value = false
      return
    }

    form.value.image_url = uploadRes.data.data.image_url

    const recogRes = await axios.post('/api/bill/recognize', { image_url: form.value.image_url }, { headers: headers.value })
    if (recogRes.data?.code === 0) {
      const r = recogRes.data.data
      form.value.amount = r.amount || 0
      amountDisplay.value = r.amount ? String(r.amount) : ''
      form.value.merchant = r.merchant || ''
      form.value.description = r.description || ''
      form.value.bill_date = r.date || new Date().toISOString().slice(0, 10)
      const match = categories.value.find(c => c.name === r.category)
      if (match) form.value.category_id = match.id
    } else {
      alert(recogRes.data?.message || '识别失败，请手动填写')
    }
  } catch (err) {
    alert('上传或识别失败: ' + (err.response?.data?.message || err.message))
  }
  recognizing.value = false
}

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
    if (res.data?.code === 0) {
      categories.value = res.data.data
      if (!form.value.category_id && categories.value.length > 0) {
        form.value.category_id = categories.value[0].id
      }
    }
  } catch (_) {}
}

async function save() {
  if (form.value.amount <= 0) { alert('请输入金额'); return }
  if (!form.value.category_id) { alert('请选择分类'); return }
  if (!form.value.bill_date) { alert('请选择日期'); return }

  saving.value = true
  try {
    const res = await axios.post('/api/bill/create', form.value, { headers: headers.value })
    if (res.data?.code === 0) {
      router.push('/bills')
    } else {
      alert(res.data?.message || '保存失败')
    }
  } catch (err) {
    alert('保存失败: ' + (err.response?.data?.message || err.message))
  }
  saving.value = false
}

onMounted(() => {
  loadCategories()
})
</script>

<style scoped>
.bill-add-page {
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
.title {
  font-size: 16px;
  font-weight: 600;
}
.upload-area {
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 16px;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.upload-placeholder {
  padding: 24px;
  text-align: center;
  color: #999;
  font-size: 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
.upload-icon {
  font-size: 32px;
}
.preview {
  position: relative;
}
.preview img {
  width: 100%;
  max-height: 200px;
  object-fit: contain;
}
.recognizing {
  position: absolute;
  bottom: 8px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0,0,0,0.7);
  color: #fff;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
}
.type-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.tab {
  flex: 1;
  padding: 10px;
  border: 2px solid #eee;
  border-radius: 8px;
  background: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.tab.active.expense { border-color: #e74c3c; color: #e74c3c; background: #fef0ef; }
.tab.active.income { border-color: #27ae60; color: #27ae60; background: #eefaf2; }
.form-group {
  margin-bottom: 16px;
}
.form-group label {
  display: block;
  font-size: 13px;
  color: #666;
  margin-bottom: 6px;
}
.form-group input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 14px;
  box-sizing: border-box;
  background: #fff;
}
.amount-input {
  font-size: 24px !important;
  font-weight: 700;
  text-align: center;
}
.category-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.category-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 60px;
  padding: 8px 4px;
  border: 2px solid #eee;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}
.category-item.selected {
  border-color: #4a90d9;
  background: #eef3fa;
}
.cat-icon {
  font-size: 20px;
}
.cat-name {
  font-size: 10px;
  color: #666;
  margin-top: 2px;
  text-align: center;
}
.save-btn {
  width: 100%;
  padding: 14px;
  background: #4a90d9;
  color: #fff;
  border: none;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  margin-top: 8px;
}
.save-btn:hover { background: #357abd; }
.save-btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>

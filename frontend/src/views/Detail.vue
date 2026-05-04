<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  User, OfficeBuilding, Ship, Tickets, Edit, ArrowLeft,
  Phone, Message
} from '@element-plus/icons-vue'
import { supplierApi } from '../api/supplier'

const router = useRouter()
const route = useRoute()

const isNew = computed(() => !route.params.id)
const isReadOnly = computed(() => route.query.mode === 'view')
const loading = ref(false)

const form = reactive({
  source: '',
  company_name: '',
  company_name_en: '',
  contact: '',
  position: '',
  phone: '',
  wechat: '',
  email: '',
  qq: '',
  address: '',
  website: '',
  nvocc_no: '',
  staff_size: '',
  transport_type: '',
  routes: '',
  shipping_line: '',
  products: '',
  remark: '',
  status: '',
  card_image: '',
  card_image_back: ''
})

const transportTypeArray = ref([])

// 同步 transportTypeArray <-> form.transport_type
watch(transportTypeArray, (val) => {
  form.transport_type = val.join(',')
})

const pageTitle = computed(() => {
  if (isNew.value) return '新增供应商'
  if (isReadOnly.value) return '查看供应商'
  return '编辑供应商'
})

const fetchSupplier = async () => {
  if (isNew.value) return

  loading.value = true
  try {
    const response = await supplierApi.getOne(route.params.id)
    Object.assign(form, response.data)
    transportTypeArray.value = form.transport_type ? form.transport_type.split(',') : []
  } catch (error) {
    ElMessage.error('获取详情失败: ' + (error.response?.data?.error || error.message))
    router.push('/list')
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  if (!form.company_name && !form.contact) {
    ElMessage.warning('请至少填写公司名称或联系人信息')
    return
  }
  try {
    if (isNew.value) {
      await supplierApi.create(form)
    } else {
      await supplierApi.update(route.params.id, form)
    }
    ElMessage.success('保存成功')
    router.push('/list')
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.response?.data?.error || error.message))
  }
}

const goBack = () => router.push('/list')

const switchToEdit = () => router.push(`/detail/${route.params.id}?mode=edit`)

// 状态样式
const getStatusClass = (status) => {
  const map = {
    '合作中': 'status-success',
    '待开发': 'status-warning',
    '已暂停': 'status-inactive'
  }
  return map[status] || 'status-default'
}

onMounted(() => {
  fetchSupplier()
})
</script>

<template>
  <div class="detail-container">
    <!-- 顶部操作栏 -->
    <header class="detail-header glass-card">
      <div class="header-left">
        <button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          <span>返回列表</span>
        </button>
        <h1 class="page-title">{{ pageTitle }}</h1>
        <span v-if="form.status" :class="['status-badge', getStatusClass(form.status)]">
          {{ form.status }}
        </span>
      </div>
      <div class="header-right">
        <template v-if="isReadOnly">
          <button class="btn btn-ghost" @click="goBack">关闭</button>
          <button class="btn btn-primary" @click="switchToEdit">
            <el-icon><Edit /></el-icon>
            <span>编辑</span>
          </button>
        </template>
        <template v-else>
          <button class="btn btn-ghost" @click="goBack">取消</button>
          <button
            class="btn btn-primary"
            @click="handleSave"
            :disabled="!form.company_name && !form.contact"
          >
            <span>保存</span>
          </button>
        </template>
      </div>
    </header>

    <div class="detail-body" v-loading="loading">
      <!-- 名片图片 -->
      <aside v-if="form.card_image || form.card_image_back" class="cards-panel glass-card">
        <div class="panel-title">
          <el-icon><OfficeBuilding /></el-icon>
          <span>名片图片</span>
        </div>

        <div class="cards-grid">
          <div v-if="form.card_image" class="card-item">
            <div class="card-label">正面</div>
            <div class="card-image">
              <el-image
                :src="`/images/${encodeURIComponent(form.card_image)}`"
                fit="contain"
                :preview-src-list="[`/images/${encodeURIComponent(form.card_image)}`]"
              />
            </div>
          </div>
          <div v-if="form.card_image_back" class="card-item">
            <div class="card-label">反面</div>
            <div class="card-image">
              <el-image
                :src="`/images/${encodeURIComponent(form.card_image_back)}`"
                fit="contain"
                :preview-src-list="[`/images/${encodeURIComponent(form.card_image_back)}`]"
              />
            </div>
          </div>
        </div>
      </aside>

      <!-- 表单区域 -->
      <main class="form-panel glass-card" :class="{ 'full-width': !form.card_image && !form.card_image_back }">
        <div class="form-scroll">
          <!-- 公司信息 -->
          <section class="form-section">
            <div class="section-header">
              <el-icon class="section-icon"><OfficeBuilding /></el-icon>
              <h2 class="section-title">公司信息</h2>
            </div>
            <div class="field-grid">
              <div class="field-item full">
                <label>公司名称</label>
                <div class="field-value company-name" v-if="isReadOnly">{{ form.company_name || '-' }}</div>
                <input v-else v-model="form.company_name" type="text" placeholder="请输入公司中文名称" />
              </div>
              <div class="field-item full">
                <label>英文名称</label>
                <div class="field-value" v-if="isReadOnly">{{ form.company_name_en || '-' }}</div>
                <input v-else v-model="form.company_name_en" type="text" placeholder="请输入公司英文名称" />
              </div>
              <div class="field-item">
                <label>NVOCC编号</label>
                <div class="field-value" v-if="isReadOnly">{{ form.nvocc_no || '-' }}</div>
                <input v-else v-model="form.nvocc_no" type="text" placeholder="NVOCC资质编号" />
              </div>
              <div class="field-item">
                <label>公司网站</label>
                <div class="field-value" v-if="isReadOnly">
                  <a v-if="form.website" :href="`https://${form.website}`" target="_blank" class="link">{{ form.website }}</a>
                  <span v-else>-</span>
                </div>
                <input v-else v-model="form.website" type="text" placeholder="www.example.com" />
              </div>
              <div class="field-item full">
                <label>公司地址</label>
                <div class="field-value" v-if="isReadOnly">{{ form.address || '-' }}</div>
                <input v-else v-model="form.address" type="text" placeholder="请输入公司地址" />
              </div>
              <div class="field-item">
                <label>来源</label>
                <div class="field-value" v-if="isReadOnly">{{ form.source || '-' }}</div>
                <input v-else v-model="form.source" type="text" placeholder="供应商来源渠道" />
              </div>
              <div class="field-item">
                <label>人员规模</label>
                <div class="field-value" v-if="isReadOnly">{{ form.staff_size || '-' }}</div>
                <input v-else v-model="form.staff_size" type="text" placeholder="公司人员规模" />
              </div>
            </div>
          </section>

          <!-- 联系人信息 -->
          <section class="form-section">
            <div class="section-header">
              <el-icon class="section-icon"><User /></el-icon>
              <h2 class="section-title">联系人信息</h2>
            </div>
            <div class="field-grid">
              <div class="field-item">
                <label>联系人</label>
                <div class="field-value contact" v-if="isReadOnly">
                  <span class="name">{{ form.contact || '-' }}</span>
                  <span v-if="form.position" class="position">{{ form.position }}</span>
                </div>
                <input v-else v-model="form.contact" type="text" placeholder="姓名" />
              </div>
              <div class="field-item">
                <label>职位</label>
                <div class="field-value" v-if="isReadOnly">{{ form.position || '-' }}</div>
                <input v-else v-model="form.position" type="text" placeholder="职位" />
              </div>
              <div class="field-item">
                <label>手机号码</label>
                <div class="field-value phone" v-if="isReadOnly">
                  <a v-if="form.phone" :href="`tel:${form.phone}`" class="link">
                    <el-icon><Phone /></el-icon>
                    {{ form.phone }}
                  </a>
                  <span v-else>-</span>
                </div>
                <input v-else v-model="form.phone" type="text" placeholder="手机号码" />
              </div>
              <div class="field-item">
                <label>电子邮箱</label>
                <div class="field-value email" v-if="isReadOnly">
                  <a v-if="form.email" :href="`mailto:${form.email}`" class="link">
                    <el-icon><Message /></el-icon>
                    {{ form.email }}
                  </a>
                  <span v-else>-</span>
                </div>
                <input v-else v-model="form.email" type="email" placeholder="email@example.com" />
              </div>
              <div class="field-item">
                <label>微信号</label>
                <div class="field-value" v-if="isReadOnly">{{ form.wechat || '-' }}</div>
                <input v-else v-model="form.wechat" type="text" placeholder="微信号" />
              </div>
              <div class="field-item">
                <label>QQ号</label>
                <div class="field-value" v-if="isReadOnly">{{ form.qq || '-' }}</div>
                <input v-else v-model="form.qq" type="text" placeholder="QQ号" />
              </div>
            </div>
          </section>

          <!-- 业务信息 -->
          <section class="form-section">
            <div class="section-header">
              <el-icon class="section-icon"><Ship /></el-icon>
              <h2 class="section-title">业务信息</h2>
            </div>
            <div class="field-grid">
              <div class="field-item">
                <label>运输方式</label>
                <div class="field-value" v-if="isReadOnly">
                  <div class="transport-tags" v-if="form.transport_type">
                    <span class="transport-tag" v-for="t in form.transport_type.split(',')" :key="t">{{ t }}</span>
                  </div>
                  <span v-else>-</span>
                </div>
                <el-select v-else v-model="transportTypeArray" multiple placeholder="请选择" style="width: 100%">
                  <el-option label="空运" value="空运" />
                  <el-option label="海运" value="海运" />
                  <el-option label="卡车" value="卡车" />
                  <el-option label="铁路" value="铁路" />
                  <el-option label="多式联运" value="多式联运" />
                </el-select>
              </div>
              <div class="field-item">
                <label>合作状态</label>
                <div class="field-value" v-if="isReadOnly">
                  <span v-if="form.status" :class="['status-tag', getStatusClass(form.status)]">{{ form.status }}</span>
                  <span v-else>-</span>
                </div>
                <el-select v-else v-model="form.status" placeholder="请选择" style="width: 100%">
                  <el-option label="合作中" value="合作中" />
                  <el-option label="待开发" value="待开发" />
                  <el-option label="已暂停" value="已暂停" />
                </el-select>
              </div>
              <div class="field-item full">
                <label>优势航线</label>
                <div class="field-value" v-if="isReadOnly">{{ form.routes || '-' }}</div>
                <input v-else v-model="form.routes" type="text" placeholder="如：东南亚、欧洲、美加等" />
              </div>
              <div class="field-item full">
                <label>船司关系</label>
                <div class="field-value" v-if="isReadOnly">{{ form.shipping_line || '-' }}</div>
                <input v-else v-model="form.shipping_line" type="text" placeholder="合作船公司" />
              </div>
              <div class="field-item full">
                <label>特色产品</label>
                <div class="field-value multiline" v-if="isReadOnly">{{ form.products || '-' }}</div>
                <textarea v-else v-model="form.products" rows="2" placeholder="特色服务或产品"></textarea>
              </div>
            </div>
          </section>

          <!-- 其他信息 -->
          <section class="form-section">
            <div class="section-header">
              <el-icon class="section-icon"><Tickets /></el-icon>
              <h2 class="section-title">其他信息</h2>
            </div>
            <div class="field-grid">
              <div class="field-item full">
                <label>备注</label>
                <div class="field-value multiline" v-if="isReadOnly">{{ form.remark || '-' }}</div>
                <textarea v-else v-model="form.remark" rows="3" placeholder="其他备注信息"></textarea>
              </div>
            </div>
          </section>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
/* 玻璃态卡片 */
.glass-card {
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-lg, 24px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}

/* 头部 */
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 28px;
  margin-bottom: 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm, 8px);
  color: var(--text-mist);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-pearl);
  transform: translateX(-3px);
}

.page-title {
  font-family: 'Playfair Display', serif;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-pearl);
  margin: 0;
}

.status-badge {
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.status-success {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-warning {
  background: rgba(251, 191, 36, 0.2);
  color: #fcd34d;
  border: 1px solid rgba(251, 191, 36, 0.3);
}

.status-inactive {
  background: rgba(107, 114, 128, 0.2);
  color: #9ca3af;
  border: 1px solid rgba(107, 114, 128, 0.3);
}

.header-right {
  display: flex;
  gap: 12px;
}

/* 按钮 */
.btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: var(--radius-sm, 8px);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-ghost {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-mist);
}

.btn-ghost:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-pearl);
}

.btn-primary {
  background: linear-gradient(135deg, var(--teal), var(--surface));
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(20, 184, 166, 0.3);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 主体布局 */
.detail-body {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 24px;
}

/* 名片面板 */
.cards-panel {
  padding: 24px;
  position: sticky;
  top: 100px;
  height: fit-content;
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-pearl);
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.panel-title .el-icon {
  color: var(--teal);
}

.cards-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card-item {
  border-radius: var(--radius-md, 16px);
  overflow: hidden;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.card-label {
  padding: 10px 16px;
  background: rgba(255, 255, 255, 0.03);
  font-size: 12px;
  font-weight: 600;
  color: var(--text-ghost);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.card-image {
  padding: 16px;
  min-height: 160px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-image :deep(.el-image) {
  width: 100%;
  max-height: 200px;
}

.card-image :deep(.el-image__inner) {
  object-fit: contain;
  width: auto;
  height: auto;
  max-width: 100%;
  max-height: 200px;
}

/* 表单面板 */
.form-panel {
  display: flex;
  flex-direction: column;
}

.form-panel.full-width {
  grid-column: 1 / -1;
}

.form-scroll {
  padding: 28px;
  overflow-y: auto;
  max-height: calc(100vh - 220px);
}

/* 表单区域 */
.form-section {
  margin-bottom: 32px;
}

.form-section:last-child {
  margin-bottom: 0;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.section-icon {
  font-size: 20px;
  color: var(--teal);
}

.section-title {
  font-family: 'Playfair Display', serif;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-pearl);
  margin: 0;
}

.field-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.field-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-item.full {
  grid-column: 1 / -1;
}

.field-item label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-ghost);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.field-value {
  padding: 12px 0;
  color: var(--text-pearl);
  font-size: 15px;
  min-height: 44px;
  display: flex;
  align-items: center;
}

.field-value.company-name {
  font-weight: 600;
  font-size: 18px;
}

.field-value.multiline {
  white-space: pre-wrap;
  line-height: 1.6;
  align-items: flex-start;
  padding-top: 12px;
}

.field-value.contact {
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}

.field-value.contact .name {
  font-weight: 600;
}

.field-value.contact .position {
  font-size: 13px;
  color: var(--text-ghost);
}

.field-value.phone,
.field-value.email {
  gap: 8px;
}

.link {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--teal);
  text-decoration: none;
  transition: color 0.2s ease;
}

.link:hover {
  color: var(--surface);
}

.transport-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.transport-tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  background: rgba(20, 184, 166, 0.15);
  color: var(--teal);
  border: 1px solid rgba(20, 184, 166, 0.25);
}

/* 状态标签 */
.status-tag {
  display: inline-flex;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

/* 输入框样式 */
.field-item input,
.field-item select,
.field-item textarea {
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm, 8px);
  color: var(--text-pearl);
  font-size: 14px;
  transition: all 0.3s ease;
}

.field-item input::placeholder,
.field-item textarea::placeholder {
  color: var(--text-ghost);
}

.field-item input:focus,
.field-item select:focus,
.field-item textarea:focus {
  outline: none;
  border-color: var(--teal);
  background: rgba(255, 255, 255, 0.05);
  box-shadow: 0 0 0 4px rgba(20, 184, 166, 0.1);
}

.field-item select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%239ca3af' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
  background-position: right 12px center;
  background-repeat: no-repeat;
  background-size: 16px;
  padding-right: 40px;
}

.field-item textarea {
  resize: vertical;
  min-height: 80px;
  line-height: 1.5;
}

/* Element Plus Select 样式 */
.field-item :deep(.el-select) {
  width: 100%;
}

.field-item :deep(.el-select .el-select__wrapper) {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm, 8px);
  padding: 8px 16px;
  box-shadow: none;
  transition: all 0.3s ease;
}

.field-item :deep(.el-select .el-select__wrapper:hover) {
  border-color: rgba(255, 255, 255, 0.15);
}

.field-item :deep(.el-select .el-select__wrapper.is-focused) {
  border-color: var(--teal);
  background: rgba(255, 255, 255, 0.05);
  box-shadow: 0 0 0 4px rgba(20, 184, 166, 0.1);
}

.field-item :deep(.el-select .el-select__selection) {
  color: var(--text-pearl);
  font-size: 14px;
}

.field-item :deep(.el-select .el-select__placeholder) {
  color: var(--text-ghost);
}

/* 响应式 */
@media (max-width: 1200px) {
  .detail-body {
    grid-template-columns: 1fr;
  }

  .cards-panel {
    position: static;
  }

  .cards-grid {
    flex-direction: row;
    flex-wrap: wrap;
  }

  .card-item {
    flex: 1;
    min-width: 280px;
  }
}

@media (max-width: 768px) {
  .detail-header {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
    padding: 16px 20px;
  }

  .header-left {
    flex-wrap: wrap;
    gap: 12px;
    width: 100%;
  }

  .back-btn {
    padding: 12px 16px;
  }

  .back-btn span {
    display: inline;
  }

  .page-title {
    font-size: 20px;
  }

  .header-right {
    width: 100%;
    justify-content: flex-end;
    gap: 10px;
  }

  .btn {
    padding: 12px 20px;
    flex: 1;
    justify-content: center;
  }

  .field-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .form-scroll {
    max-height: none;
    padding: 20px;
  }

  .section-title {
    font-size: 16px;
  }

  .field-item input,
  .field-item select,
  .field-item textarea {
    padding: 14px 16px;
    font-size: 16px;
  }

  .field-value {
    font-size: 16px;
    min-height: 48px;
  }

  .cards-grid {
    flex-direction: column;
  }

  .card-item {
    min-width: auto;
  }
}

@media (max-width: 480px) {
  .detail-container {
    padding: 12px;
  }

  .detail-header {
    padding: 12px 16px;
    margin-bottom: 16px;
  }

  .back-btn {
    width: 100%;
    justify-content: center;
  }

  .page-title {
    font-size: 18px;
    width: 100%;
  }

  .status-badge {
    font-size: 11px;
    padding: 5px 12px;
  }

  .header-right {
    flex-direction: column;
  }

  .btn {
    width: 100%;
  }

  .form-scroll {
    padding: 16px;
  }

  .form-section {
    margin-bottom: 24px;
  }

  .section-header {
    margin-bottom: 16px;
    padding-bottom: 10px;
  }

  .section-title {
    font-size: 15px;
  }

  .cards-panel {
    padding: 16px;
  }

  .card-label {
    padding: 8px 12px;
    font-size: 11px;
  }

  .card-image {
    padding: 12px;
    min-height: 120px;
  }
}
</style>

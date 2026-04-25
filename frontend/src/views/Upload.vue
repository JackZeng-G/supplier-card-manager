<script setup>
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Upload as UploadIcon, User, OfficeBuilding, Ship, Tickets,
  DocumentChecked, Check, Loading
} from '@element-plus/icons-vue'
import { supplierApi } from '../api/supplier'

const router = useRouter()

const uploadRef = ref()
const uploadBackRef = ref()
const loading = ref(false)
const loadingBack = ref(false)
const ocrResult = ref(null)
const ocrResultBack = ref(null)
const cardImage = ref('')
const cardImageBack = ref('')
const activeStep = ref(0)
const showRawText = ref(false)

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
  card_image_back: ''
})

const transportTypeArray = ref([])

watch(transportTypeArray, (val) => {
  form.transport_type = val.join(',')
})

const handleBeforeUpload = (file) => {
  const isImage = ['image/jpeg', 'image/png', 'image/jpg'].includes(file.type)
  const isLt5M = file.size / 1024 / 1024 < 5

  if (!isImage) {
    ElMessage.error('只能上传 JPG/PNG 格式的图片!')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过 5MB!')
    return false
  }
  return true
}

const handleUploadRequest = async (options) => {
  loading.value = true
  try {
    const response = await supplierApi.uploadCard(options.file)
    if (response.data.card_image) {
      cardImage.value = response.data.card_image
    }
    if (response.data.ocr_result) {
      ocrResult.value = response.data.ocr_result
      mergeOcrResult(response.data.ocr_result)
      activeStep.value = 1
    }
    ElMessage.success('名片正面识别成功')
  } catch (error) {
    ElMessage.error('名片正面识别失败: ' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

const handleUploadBackRequest = async (options) => {
  loadingBack.value = true
  try {
    const response = await supplierApi.uploadCardBack(options.file)
    if (response.data.card_image_back) {
      cardImageBack.value = response.data.card_image_back
      form.card_image_back = response.data.card_image_back
    }
    if (response.data.ocr_result) {
      ocrResultBack.value = response.data.ocr_result
      mergeOcrResult(response.data.ocr_result, true)
    }
    ElMessage.success('名片反面识别成功')
  } catch (error) {
    ElMessage.error('名片反面识别失败: ' + (error.response?.data?.error || error.message))
  } finally {
    loadingBack.value = false
  }
}

const mergeOcrResult = (result, isBack = false) => {
  if (result.company_name) form.company_name = result.company_name
  if (result.company_name_en) form.company_name_en = result.company_name_en
  if (result.contact) form.contact = result.contact
  if (result.position) form.position = result.position
  if (result.phone) form.phone = result.phone
  if (result.wechat) form.wechat = result.wechat
  if (result.email) form.email = result.email
  if (result.qq) form.qq = result.qq
  if (result.address) form.address = result.address
  if (result.website) form.website = result.website
  if (result.nvocc_no) form.nvocc_no = result.nvocc_no
  if (result.routes) form.routes = result.routes
  if (result.shipping_line) form.shipping_line = result.shipping_line
  if (result.products) form.products = result.products

  if (result.unmatched_text) {
    const prefix = isBack ? '【反面未识别信息】\n' : '【正面未识别信息】\n'
    if (form.remark) {
      form.remark += '\n\n' + prefix + result.unmatched_text
    } else {
      form.remark = prefix + result.unmatched_text
    }
  }
}

const saveSupplier = async () => {
  if (!form.company_name && !form.contact) {
    ElMessage.warning('请至少填写公司名称或联系人信息')
    return
  }
  try {
    const data = { ...form, card_image: cardImage.value }
    await supplierApi.create(data)
    ElMessage.success('保存成功')
    router.push('/list')
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.response?.data?.error || error.message))
  }
}

const goToList = () => router.push('/list')

const resetForm = () => {
  Object.keys(form).forEach(key => { form[key] = '' })
  cardImage.value = ''
  cardImageBack.value = ''
  ocrResult.value = null
  ocrResultBack.value = null
  activeStep.value = 0
  transportTypeArray.value = []
}
</script>

<template>
  <div class="upload-container">
    <!-- 步骤指示器 -->
    <section class="steps-panel glass-card">
      <div class="steps-inner">
        <div
          v-for="(step, index) in [
            { title: '上传名片', icon: UploadIcon },
            { title: '确认信息', icon: DocumentChecked },
            { title: '保存完成', icon: Check }
          ]"
          :key="index"
          :class="['step-item', { active: activeStep >= index, completed: activeStep > index }]"
        >
          <div class="step-number">
            <el-icon v-if="activeStep > index"><Check /></el-icon>
            <span v-else>{{ index + 1 }}</span>
          </div>
          <div class="step-info">
            <span class="step-title">{{ step.title }}</span>
          </div>
          <div v-if="index < 2" class="step-connector" :class="{ active: activeStep > index }"></div>
        </div>
      </div>
    </section>

    <div class="content-grid">
      <!-- 左侧：上传区 -->
      <section class="upload-panel glass-card">
        <div class="panel-header">
          <h2 class="panel-title">
            <el-icon><UploadIcon /></el-icon>
            <span>名片上传</span>
          </h2>
          <p class="panel-desc">上传名片图片，系统将自动识别并提取信息</p>
        </div>

        <div class="upload-grid">
          <!-- 正面 -->
          <div class="upload-box">
            <div class="upload-label">
              <span class="label-tag primary">正面</span>
              <span class="label-text">名片正面（必填）</span>
            </div>
            <div
              class="upload-area"
              :class="{ 'has-image': cardImage, loading: loading }"
              @click="uploadRef?.$el?.querySelector('input')?.click()"
            >
              <template v-if="!cardImage">
                <div class="upload-placeholder">
                  <div class="upload-icon-wrapper primary">
                    <el-icon class="upload-icon"><UploadIcon /></el-icon>
                  </div>
                  <div class="upload-text">
                    <p class="main-text">点击或拖拽上传</p>
                    <p class="sub-text">支持 JPG/PNG，最大 5MB</p>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="image-preview">
                  <img :src="`/uploads/${encodeURIComponent(cardImage)}`" alt="名片正面" />
                  <div class="preview-overlay">
                    <span>点击更换</span>
                  </div>
                </div>
              </template>
              <div v-if="loading" class="loading-overlay">
                <div class="spinner"></div>
                <span>正在识别...</span>
              </div>
            </div>
            <el-upload
              ref="uploadRef"
              class="hidden-upload"
              :auto-upload="true"
              :show-file-list="false"
              :before-upload="handleBeforeUpload"
              :http-request="handleUploadRequest"
            />
          </div>

          <!-- 反面 -->
          <div class="upload-box">
            <div class="upload-label">
              <span class="label-tag secondary">反面</span>
              <span class="label-text">名片反面（选填）</span>
            </div>
            <div
              class="upload-area"
              :class="{ 'has-image': cardImageBack, loading: loadingBack }"
              @click="uploadBackRef?.$el?.querySelector('input')?.click()"
            >
              <template v-if="!cardImageBack">
                <div class="upload-placeholder">
                  <div class="upload-icon-wrapper secondary">
                    <el-icon class="upload-icon"><UploadIcon /></el-icon>
                  </div>
                  <div class="upload-text">
                    <p class="main-text">点击或拖拽上传</p>
                    <p class="sub-text">支持 JPG/PNG，最大 5MB</p>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="image-preview">
                  <img :src="`/uploads/${encodeURIComponent(cardImageBack)}`" alt="名片反面" />
                  <div class="preview-overlay">
                    <span>点击更换</span>
                  </div>
                </div>
              </template>
              <div v-if="loadingBack" class="loading-overlay">
                <div class="spinner"></div>
                <span>正在识别...</span>
              </div>
            </div>
            <el-upload
              ref="uploadBackRef"
              class="hidden-upload"
              :auto-upload="true"
              :show-file-list="false"
              :before-upload="handleBeforeUpload"
              :http-request="handleUploadBackRequest"
            />
          </div>
        </div>

        <!-- OCR结果 -->
        <div v-if="ocrResult || ocrResultBack" class="ocr-result-box">
          <div class="result-header" @click="showRawText = !showRawText">
            <el-icon class="result-icon"><DocumentChecked /></el-icon>
            <span class="result-title">OCR识别结果</span>
            <span class="result-toggle">{{ showRawText ? '收起' : '展开' }}</span>
          </div>
          <div v-if="showRawText" class="result-content">
            <div v-if="ocrResult" class="raw-block">
              <span class="block-label">正面原始文本</span>
              <pre>{{ ocrResult.raw_text }}</pre>
            </div>
            <div v-if="ocrResultBack" class="raw-block">
              <span class="block-label">反面原始文本</span>
              <pre>{{ ocrResultBack.raw_text }}</pre>
            </div>
          </div>
        </div>
      </section>

      <!-- 右侧：表单区 -->
      <section class="form-panel glass-card">
        <div class="panel-header with-actions">
          <h2 class="panel-title">
            <el-icon><OfficeBuilding /></el-icon>
            <span>供应商信息</span>
          </h2>
          <div class="header-actions">
            <button class="btn btn-ghost" @click="resetForm">重置</button>
            <button class="btn btn-ghost" @click="goToList">返回</button>
            <button
              class="btn btn-primary"
              @click="saveSupplier"
              :disabled="!form.company_name && !form.contact"
            >
              <el-icon><Check /></el-icon>
              <span>保存</span>
            </button>
          </div>
        </div>

        <div class="form-body">
          <!-- 公司信息 -->
          <div class="form-section">
            <div class="section-title">
              <el-icon><OfficeBuilding /></el-icon>
              <span>公司信息</span>
            </div>
            <div class="form-grid">
              <div class="form-item full">
                <label>公司名称 <span class="required">*</span></label>
                <input v-model="form.company_name" type="text" placeholder="请输入公司中文名称" />
              </div>
              <div class="form-item full">
                <label>英文名称</label>
                <input v-model="form.company_name_en" type="text" placeholder="请输入公司英文名称" />
              </div>
              <div class="form-item">
                <label>NVOCC编号</label>
                <input v-model="form.nvocc_no" type="text" placeholder="NVOCC资质编号" />
              </div>
              <div class="form-item">
                <label>公司网站</label>
                <input v-model="form.website" type="text" placeholder="www.example.com" />
              </div>
              <div class="form-item full">
                <label>公司地址</label>
                <input v-model="form.address" type="text" placeholder="请输入公司地址" />
              </div>
            </div>
          </div>

          <!-- 联系人信息 -->
          <div class="form-section">
            <div class="section-title">
              <el-icon><User /></el-icon>
              <span>联系人信息</span>
            </div>
            <div class="form-grid">
              <div class="form-item">
                <label>联系人</label>
                <input v-model="form.contact" type="text" placeholder="姓名" />
              </div>
              <div class="form-item">
                <label>职位</label>
                <input v-model="form.position" type="text" placeholder="职位" />
              </div>
              <div class="form-item">
                <label>手机号码</label>
                <input v-model="form.phone" type="text" placeholder="手机号码" />
              </div>
              <div class="form-item">
                <label>电子邮箱</label>
                <input v-model="form.email" type="email" placeholder="email@example.com" />
              </div>
              <div class="form-item">
                <label>微信号</label>
                <input v-model="form.wechat" type="text" placeholder="微信号" />
              </div>
              <div class="form-item">
                <label>QQ号</label>
                <input v-model="form.qq" type="text" placeholder="QQ号" />
              </div>
            </div>
          </div>

          <!-- 业务信息 -->
          <div class="form-section">
            <div class="section-title">
              <el-icon><Ship /></el-icon>
              <span>业务信息</span>
            </div>
            <div class="form-grid">
              <div class="form-item">
                <label>运输方式</label>
                <el-select v-model="transportTypeArray" multiple placeholder="请选择" style="width: 100%">
                  <el-option label="空运" value="空运" />
                  <el-option label="海运" value="海运" />
                  <el-option label="卡车" value="卡车" />
                  <el-option label="铁路" value="铁路" />
                  <el-option label="多式联运" value="多式联运" />
                </el-select>
              </div>
              <div class="form-item">
                <label>合作状态</label>
                <select v-model="form.status">
                  <option value="">请选择</option>
                  <option value="合作中">合作中</option>
                  <option value="待开发">待开发</option>
                  <option value="已暂停">已暂停</option>
                </select>
              </div>
              <div class="form-item full">
                <label>优势航线</label>
                <input v-model="form.routes" type="text" placeholder="如：东南亚、欧洲、美加等" />
              </div>
              <div class="form-item full">
                <label>船司关系</label>
                <input v-model="form.shipping_line" type="text" placeholder="合作船公司" />
              </div>
              <div class="form-item full">
                <label>特色产品</label>
                <textarea v-model="form.products" rows="2" placeholder="特色服务或产品"></textarea>
              </div>
            </div>
          </div>

          <!-- 其他信息 -->
          <div class="form-section">
            <div class="section-title">
              <el-icon><Tickets /></el-icon>
              <span>其他信息</span>
            </div>
            <div class="form-grid">
              <div class="form-item">
                <label>来源</label>
                <input v-model="form.source" type="text" placeholder="供应商来源渠道" />
              </div>
              <div class="form-item">
                <label>人员规模</label>
                <input v-model="form.staff_size" type="text" placeholder="公司人员规模" />
              </div>
              <div class="form-item full">
                <label>备注</label>
                <textarea v-model="form.remark" rows="3" placeholder="其他备注信息"></textarea>
              </div>
            </div>
          </div>
        </div>
      </section>
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

/* 步骤指示器 */
.steps-panel {
  padding: 32px;
  margin-bottom: 24px;
}

.steps-inner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0;
}

.step-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.step-number {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.05);
  border: 2px solid rgba(255, 255, 255, 0.15);
  color: var(--text-ghost);
  font-size: 16px;
  font-weight: 600;
  transition: all 0.3s ease;
}

.step-item.active .step-number {
  background: linear-gradient(135deg, var(--teal), var(--surface));
  border-color: transparent;
  color: white;
  box-shadow: 0 4px 20px rgba(20, 184, 166, 0.4);
}

.step-item.completed .step-number {
  background: linear-gradient(135deg, var(--emerald), #059669);
  border-color: transparent;
  color: white;
}

.step-info {
  display: flex;
  flex-direction: column;
}

.step-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-ghost);
  transition: color 0.3s ease;
}

.step-item.active .step-title {
  color: var(--text-pearl);
}

.step-connector {
  width: 80px;
  height: 2px;
  background: rgba(255, 255, 255, 0.1);
  margin: 0 20px;
  border-radius: 1px;
  transition: background 0.3s ease;
}

.step-connector.active {
  background: linear-gradient(90deg, var(--teal), var(--surface));
}

/* 内容网格 */
.content-grid {
  display: grid;
  grid-template-columns: 380px 1fr;
  gap: 24px;
}

/* 面板通用 */
.panel-header {
  padding: 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.panel-header.with-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: 'Playfair Display', serif;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-pearl);
  margin: 0;
}

.panel-title .el-icon {
  color: var(--teal);
}

.panel-desc {
  font-size: 13px;
  color: var(--text-mist);
  margin: 8px 0 0;
}

/* 上传区 */
.upload-grid {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.upload-label {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.label-tag {
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.label-tag.primary {
  background: linear-gradient(135deg, var(--teal), var(--surface));
  color: white;
}

.label-tag.secondary {
  background: linear-gradient(135deg, var(--surface), var(--ocean));
  color: white;
}

.label-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-pearl);
}

.upload-area {
  border: 2px dashed rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-md, 16px);
  height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.02);
}

.upload-area:hover {
  border-color: var(--teal);
  background: rgba(20, 184, 166, 0.05);
}

.upload-area.has-image {
  border-style: solid;
  border-color: var(--teal);
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  text-align: center;
}

.upload-icon-wrapper {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.upload-icon-wrapper.primary {
  background: linear-gradient(135deg, var(--teal), var(--surface));
  color: white;
}

.upload-icon-wrapper.secondary {
  background: linear-gradient(135deg, var(--surface), var(--ocean));
  color: white;
}

.upload-icon {
  font-size: 24px;
}

.upload-text .main-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-pearl);
  margin: 0 0 4px;
}

.upload-text .sub-text {
  font-size: 12px;
  color: var(--text-ghost);
  margin: 0;
}

.image-preview {
  width: 100%;
  height: 100%;
  position: relative;
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  padding: 12px;
}

.preview-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.image-preview:hover .preview-overlay {
  opacity: 1;
}

.preview-overlay span {
  padding: 10px 20px;
  background: var(--teal);
  border-radius: var(--radius-sm, 8px);
  font-size: 13px;
  font-weight: 500;
  color: white;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(10, 25, 47, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--teal);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-overlay span {
  font-size: 13px;
  color: var(--text-mist);
}

.hidden-upload {
  display: none;
}

/* OCR结果 */
.ocr-result-box {
  margin: 0 24px 24px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-md, 16px);
  overflow: hidden;
}

.result-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  background: rgba(255, 255, 255, 0.03);
  cursor: pointer;
  transition: background 0.3s ease;
}

.result-header:hover {
  background: rgba(255, 255, 255, 0.05);
}

.result-icon {
  color: var(--emerald);
  font-size: 18px;
}

.result-title {
  flex: 1;
  font-weight: 500;
  color: var(--text-pearl);
}

.result-toggle {
  font-size: 12px;
  color: var(--teal);
  font-weight: 500;
}

.result-content {
  padding: 16px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.raw-block {
  margin-bottom: 16px;
}

.raw-block:last-child {
  margin-bottom: 0;
}

.block-label {
  display: block;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-ghost);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
}

.raw-block pre {
  background: rgba(0, 0, 0, 0.3);
  padding: 12px;
  border-radius: var(--radius-sm, 8px);
  font-size: 12px;
  color: var(--text-mist);
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  max-height: 120px;
  overflow-y: auto;
}

/* 表单面板 */
.form-panel {
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 200px);
}

.form-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.header-actions {
  display: flex;
  gap: 10px;
}

/* 按钮 */
.btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border: none;
  border-radius: var(--radius-sm, 8px);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-ghost {
  background: rgba(255, 255, 255, 0.05);
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

/* 表单区域 */
.form-section {
  margin-bottom: 28px;
}

.form-section:last-child {
  margin-bottom: 0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-pearl);
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.section-title .el-icon {
  color: var(--teal);
  font-size: 18px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-item.full {
  grid-column: 1 / -1;
}

.form-item label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-mist);
}

.required {
  color: var(--coral);
}

.form-item input,
.form-item select,
.form-item textarea {
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm, 8px);
  color: var(--text-pearl);
  font-size: 14px;
  transition: all 0.3s ease;
}

.form-item input::placeholder,
.form-item textarea::placeholder {
  color: var(--text-ghost);
}

.form-item input:focus,
.form-item select:focus,
.form-item textarea:focus {
  outline: none;
  border-color: var(--teal);
  background: rgba(255, 255, 255, 0.05);
  box-shadow: 0 0 0 4px rgba(20, 184, 166, 0.1);
}

.form-item select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%239ca3af' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
  background-position: right 12px center;
  background-repeat: no-repeat;
  background-size: 16px;
  padding-right: 40px;
}

.form-item textarea {
  resize: vertical;
  min-height: 80px;
  line-height: 1.5;
}

/* 响应式 */
@media (max-width: 1200px) {
  .content-grid {
    grid-template-columns: 1fr;
  }

  .form-panel {
    max-height: none;
  }
}

@media (max-width: 768px) {
  .steps-panel {
    padding: 20px;
    margin-bottom: 16px;
  }

  .steps-inner {
    flex-wrap: wrap;
    gap: 12px;
    justify-content: center;
  }

  .step-connector {
    display: none;
  }

  .step-number {
    width: 40px;
    height: 40px;
    font-size: 14px;
  }

  .step-title {
    font-size: 13px;
  }

  .upload-grid {
    padding: 16px;
    gap: 16px;
  }

  .upload-area {
    height: 150px;
  }

  .upload-icon-wrapper {
    width: 48px;
    height: 48px;
  }

  .upload-icon {
    font-size: 20px;
  }

  .form-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }

  .panel-header.with-actions {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap;
    gap: 8px;
  }

  .header-actions .btn {
    flex: 1;
    justify-content: center;
    min-width: 80px;
  }

  .btn {
    padding: 12px 16px;
  }

  .form-body {
    padding: 16px;
  }

  .form-section {
    margin-bottom: 20px;
  }

  .section-title {
    font-size: 14px;
    margin-bottom: 14px;
  }

  .form-item input,
  .form-item select,
  .form-item textarea {
    padding: 14px 16px;
    font-size: 16px;
  }

  .form-item label {
    font-size: 12px;
  }

  .ocr-result-box {
    margin: 0 16px 16px;
  }

  .result-header {
    padding: 12px 16px;
  }

  .result-content {
    padding: 12px 16px;
  }
}

@media (max-width: 480px) {
  .upload-container {
    padding: 12px;
  }

  .steps-panel {
    padding: 16px;
  }

  .step-item {
    flex-direction: column;
    text-align: center;
    gap: 8px;
  }

  .step-info {
    align-items: center;
  }

  .panel-title {
    font-size: 18px;
  }

  .panel-desc {
    font-size: 12px;
  }

  .upload-label {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }

  .upload-area {
    height: 130px;
  }

  .upload-text .main-text {
    font-size: 13px;
  }

  .upload-text .sub-text {
    font-size: 11px;
  }

  .header-actions {
    gap: 6px;
  }

  .btn {
    padding: 10px 12px;
    font-size: 13px;
  }

  .btn span {
    display: none;
  }

  .btn-primary span {
    display: inline;
  }
}
</style>

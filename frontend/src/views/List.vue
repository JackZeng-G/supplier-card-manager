<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search, View, Edit, Delete, Download,
  OfficeBuilding, TrendCharts, Clock, CircleClose, Check
} from '@element-plus/icons-vue'
import { supplierApi } from '../api/supplier'

const router = useRouter()

const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const searchKeyword = ref('')
const filterStatus = ref('')
const filterTransportType = ref([])

// 统计数据
const stats = ref({
  total: 0,
  cooperating: 0,
  pending: 0,
  paused: 0
})

// 获取供应商列表
const fetchSuppliers = async () => {
  loading.value = true
  try {
    const response = await supplierApi.getList({
      page: currentPage.value,
      page_size: pageSize.value,
      search: searchKeyword.value,
      status: filterStatus.value,
      transport_type: filterTransportType.value.join(',')
    })
    tableData.value = response.data.list || []
    total.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('获取列表失败: ' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

// 更新统计数据
const fetchStats = async () => {
  try {
    const response = await supplierApi.getStats()
    stats.value = response.data
  } catch (error) {
    // 静默失败，统计数据不影响主要功能
  }
}

// 搜索防抖
let searchTimer = null
const handleSearchInput = () => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    fetchSuppliers()
  }, 500)  // 500ms 防抖延迟
}

// 立即搜索（点击搜索按钮或按回车）
const handleSearch = () => {
  if (searchTimer) {
    clearTimeout(searchTimer)
    searchTimer = null
  }
  currentPage.value = 1
  fetchSuppliers()
}

// 分页
const handlePageChange = (page) => {
  currentPage.value = page
  fetchSuppliers()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  fetchSuppliers()
}

// 操作
const handleView = (row) => router.push(`/detail/${row.id}?mode=view`)
const handleEdit = (row) => router.push(`/detail/${row.id}?mode=edit`)

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该供应商吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await supplierApi.delete(row.id)
    ElMessage.success('删除成功')
    fetchSuppliers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.response?.data?.error || error.message))
    }
  }
}

const handleExport = async () => {
  try {
    const response = await supplierApi.exportExcel()
    const blob = new Blob([response.data], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `suppliers_${new Date().toISOString().split('T')[0]}.xlsx`
    link.click()
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

const handleAdd = () => router.push('/upload')

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
  fetchSuppliers()
  fetchStats()
})
</script>

<template>
  <div class="list-container">
    <!-- 统计面板 -->
    <section class="stats-panel">
      <div class="stat-card" v-for="(stat, index) in [
        { icon: OfficeBuilding, value: stats.total, label: '供应商总数', class: 'total' },
        { icon: Check, value: stats.cooperating, label: '合作中', class: 'success' },
        { icon: Clock, value: stats.pending, label: '待开发', class: 'warning' },
        { icon: CircleClose, value: stats.paused, label: '已暂停', class: 'inactive' }
      ]" :key="index" :class="stat.class" :style="{ animationDelay: `${index * 0.1}s` }">
        <div class="stat-icon-wrapper">
          <el-icon class="stat-icon"><component :is="stat.icon" /></el-icon>
        </div>
        <div class="stat-body">
          <span class="stat-value">{{ stat.value }}</span>
          <span class="stat-label">{{ stat.label }}</span>
        </div>
        <div class="stat-glow"></div>
      </div>
    </section>

    <!-- 搜索区域 -->
    <section class="search-panel glass-card">
      <div class="search-inner">
        <div class="search-row">
          <div class="search-field">
            <el-icon class="search-icon"><Search /></el-icon>
            <input
              v-model="searchKeyword"
              type="text"
              class="search-input"
              placeholder="搜索公司名称、联系人、电话、邮箱..."
              @input="handleSearchInput"
              @keyup.enter="handleSearch"
            />
          </div>
          <div class="filter-group">
            <el-select v-model="filterStatus" placeholder="合作状态" clearable @change="handleSearch">
              <el-option label="合作中" value="合作中" />
              <el-option label="待开发" value="待开发" />
              <el-option label="已暂停" value="已暂停" />
            </el-select>
            <el-select v-model="filterTransportType" placeholder="运输方式" clearable multiple @change="handleSearch">
              <el-option label="空运" value="空运" />
              <el-option label="海运" value="海运" />
              <el-option label="卡车" value="卡车" />
              <el-option label="铁路" value="铁路" />
              <el-option label="多式联运" value="多式联运" />
            </el-select>
          </div>
        </div>
        <div class="action-group">
          <button class="btn btn-search" @click="handleSearch">
            <el-icon><Search /></el-icon>
            <span>搜索</span>
          </button>
          <button class="btn btn-export" @click="handleExport">
            <el-icon><Download /></el-icon>
            <span>导出Excel</span>
          </button>
        </div>
      </div>
    </section>

    <!-- 数据表格 -->
    <section class="table-panel glass-card">
      <div class="table-header">
        <h2 class="table-title">
          <el-icon><TrendCharts /></el-icon>
          <span>供应商列表</span>
        </h2>
        <span class="record-count">共 {{ total }} 条记录</span>
      </div>

      <div class="table-wrapper" v-loading="loading">
        <table class="data-table">
          <thead>
            <tr>
              <th class="col-index">#</th>
              <th class="col-company">公司名称</th>
              <th class="col-contact">联系人</th>
              <th class="col-phone">电话</th>
              <th class="col-products">特色产品</th>
              <th class="col-transport">运输方式</th>
              <th class="col-routes">优势航线</th>
              <th class="col-status" style="text-align: center;">状态</th>
              <th class="col-actions" style="text-align: center;">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="!loading && tableData.length === 0">
              <td colspan="9" class="empty-state">
                <div class="empty-content">
                  <el-icon class="empty-icon"><OfficeBuilding /></el-icon>
                  <p>暂无供应商数据</p>
                  <button class="btn btn-add" @click="handleAdd">添加第一个供应商</button>
                </div>
              </td>
            </tr>
            <tr
              v-for="(row, index) in tableData"
              :key="row.id"
              class="data-row"
              :style="{ animationDelay: `${index * 0.05}s` }"
            >
              <td class="col-index" data-label="#"> {{ (currentPage - 1) * pageSize + index + 1 }}</td>
              <td class="col-company" data-label="公司">
                <div class="company-cell">
                  <span class="company-name">{{ row.company_name || '未填写' }}</span>
                  <span class="company-en" v-if="row.company_name_en">{{ row.company_name_en }}</span>
                </div>
              </td>
              <td class="col-contact" data-label="联系人">
                <div class="contact-cell">
                  <span class="contact-name">{{ row.contact || '-' }}</span>
                  <span class="contact-position" v-if="row.position">{{ row.position }}</span>
                </div>
              </td>
              <td class="col-phone" data-label="电话">{{ row.phone || '-' }}</td>
              <td class="col-products" data-label="特色产品" :title="row.products">{{ row.products || '-' }}</td>
              <td class="col-transport" data-label="运输方式">
                <div class="transport-tags" v-if="row.transport_type">
                  <span class="transport-tag" v-for="t in row.transport_type.split(',')" :key="t">{{ t }}</span>
                </div>
                <span v-else>-</span>
              </td>
              <td class="col-routes" data-label="航线" :title="row.routes">{{ row.routes || '-' }}</td>
              <td class="col-status" data-label="状态">
                <span :class="['status-tag', getStatusClass(row.status)]">
                  {{ row.status || '待定' }}
                </span>
              </td>
              <td class="col-actions" data-label="操作">
                <div class="action-btns">
                  <button class="action-btn view" @click="handleView(row)" title="查看">
                    <el-icon><View /></el-icon>
                  </button>
                  <button class="action-btn edit" @click="handleEdit(row)" title="编辑">
                    <el-icon><Edit /></el-icon>
                  </button>
                  <button class="action-btn delete" @click="handleDelete(row)" title="删除">
                    <el-icon><Delete /></el-icon>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div class="pagination-bar" v-if="total > 0">
        <div class="page-info">
          显示 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, total) }} / {{ total }}
        </div>
        <div class="page-controls">
          <select class="page-select" v-model="pageSize" @change="handleSizeChange(pageSize)">
            <option :value="10">10条/页</option>
            <option :value="20">20条/页</option>
            <option :value="50">50条/页</option>
          </select>
          <button
            class="page-btn"
            :disabled="currentPage === 1"
            @click="handlePageChange(currentPage - 1)"
          >上一页</button>
          <span class="page-current">第 {{ currentPage }} 页</span>
          <button
            class="page-btn"
            :disabled="currentPage >= Math.ceil(total / pageSize)"
            @click="handlePageChange(currentPage + 1)"
          >下一页</button>
        </div>
      </div>
    </section>
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

/* 统计面板 */
.stats-panel {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  position: relative;
  padding: 24px;
  border-radius: var(--radius-lg, 24px);
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  gap: 16px;
  overflow: hidden;
  animation: fadeSlideUp 0.6s ease forwards;
  opacity: 0;
  transform: translateY(20px);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.3);
}

@keyframes fadeSlideUp {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.stat-glow {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--teal), transparent);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.stat-card:hover .stat-glow {
  opacity: 1;
}

.stat-icon-wrapper {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md, 16px);
  background: rgba(255, 255, 255, 0.05);
}

.stat-card.total .stat-icon-wrapper {
  background: linear-gradient(135deg, rgba(20, 184, 166, 0.3), rgba(30, 58, 95, 0.3));
  color: var(--teal);
}

.stat-card.success .stat-icon-wrapper {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.3), rgba(5, 150, 105, 0.3));
  color: var(--emerald);
}

.stat-card.warning .stat-icon-wrapper {
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.3), rgba(245, 158, 11, 0.3));
  color: var(--amber);
}

.stat-card.inactive .stat-icon-wrapper {
  background: linear-gradient(135deg, rgba(107, 114, 128, 0.3), rgba(75, 85, 99, 0.3));
  color: var(--text-ghost);
}

.stat-icon {
  font-size: 28px;
}

.stat-body {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-family: 'Playfair Display', serif;
  font-size: 32px;
  font-weight: 700;
  color: var(--text-pearl);
  line-height: 1;
}

.stat-label {
  font-size: 13px;
  color: var(--text-mist);
  margin-top: 4px;
  font-weight: 500;
}

/* 搜索面板 */
.search-panel {
  padding: 20px 24px;
  margin-bottom: 24px;
}

.search-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  flex-wrap: wrap;
}

.search-row {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  gap: 12px;
}

.filter-group :deep(.el-select) {
  width: 160px;
}

.filter-group :deep(.el-select .el-select__wrapper) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm, 8px);
  box-shadow: none;
}

.filter-group :deep(.el-select .el-select__wrapper:hover) {
  border-color: rgba(255, 255, 255, 0.15);
}

.filter-group :deep(.el-select .el-select__wrapper.is-focused) {
  border-color: var(--teal);
}

.filter-group :deep(.el-select .el-select__selection),
.filter-group :deep(.el-select .el-select__placeholder) {
  color: var(--text-pearl);
  font-size: 13px;
}

.search-field {
  flex: 1;
  min-width: 300px;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 16px;
  color: var(--text-ghost);
  font-size: 18px;
}

.search-input {
  width: 100%;
  padding: 14px 16px 14px 48px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md, 16px);
  color: var(--text-pearl);
  font-size: 14px;
  transition: all 0.3s ease;
}

.search-input::placeholder {
  color: var(--text-ghost);
}

.search-input:focus {
  outline: none;
  border-color: var(--teal);
  background: rgba(255, 255, 255, 0.08);
  box-shadow: 0 0 0 4px rgba(20, 184, 166, 0.1);
}

.action-group {
  display: flex;
  gap: 12px;
}

/* 按钮样式 */
.btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border: none;
  border-radius: var(--radius-md, 16px);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.btn-search {
  background: linear-gradient(135deg, var(--teal), var(--surface));
  color: white;
}

.btn-search:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(20, 184, 166, 0.3);
}

.btn-add {
  background: linear-gradient(135deg, var(--coral), #ea580c);
  color: white;
}

.btn-add:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(249, 115, 22, 0.3);
}

.btn-export {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: var(--text-pearl);
}

.btn-export:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: translateY(-2px);
}

/* 表格面板 */
.table-panel {
  overflow: hidden;
}

.table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.table-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-family: 'Playfair Display', serif;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-pearl);
  margin: 0;
}

.table-title .el-icon {
  color: var(--teal);
}

.record-count {
  font-size: 13px;
  color: var(--text-mist);
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: var(--radius-sm, 8px);
}

.table-wrapper {
  overflow-x: auto;
}

/* 表格样式 */
.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: 18px 20px;
  text-align: left;
  font-size: 13px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-pearl);
  background: linear-gradient(180deg, rgba(20, 184, 166, 0.15) 0%, rgba(30, 58, 95, 0.2) 100%);
  border-bottom: 2px solid var(--teal);
  white-space: nowrap;
}

.data-table th:first-child {
  border-radius: var(--radius-sm, 8px) 0 0 0;
}

.data-table th:last-child {
  border-radius: 0 var(--radius-sm, 8px) 0 0;
}

.data-table td {
  padding: 18px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  vertical-align: middle;
}

.data-row {
  animation: fadeIn 0.5s ease forwards;
  opacity: 0;
}

@keyframes fadeIn {
  to { opacity: 1; }
}

.data-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

/* 单元格样式 */
.company-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.company-name {
  font-weight: 600;
  color: var(--text-pearl);
}

.company-en {
  font-size: 12px;
  color: var(--text-ghost);
  font-style: italic;
}

.contact-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.contact-name {
  font-weight: 500;
  color: var(--text-pearl);
}

.contact-position {
  font-size: 12px;
  color: var(--text-ghost);
}

/* 列宽设置 - 同时应用于表头和表体 */
.col-index,
.col-company,
.col-contact,
.col-phone,
.col-products,
.col-transport,
.col-routes,
.col-status,
.col-actions {
  vertical-align: middle;
}

.col-index {
  width: 60px;
  text-align: center !important;
}

.col-company {
  min-width: 200px;
}

.col-contact {
  width: 100px;
}

.col-phone {
  width: 130px;
  font-family: 'DM Sans', monospace;
  color: var(--text-mist);
}

.col-products {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-mist);
}

.col-transport {
  min-width: 100px;
}

.transport-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.transport-tag {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  background: rgba(20, 184, 166, 0.15);
  color: var(--teal);
  border: 1px solid rgba(20, 184, 166, 0.25);
  white-space: nowrap;
}

.col-routes {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.col-routes {
  min-width: 120px;
}

.col-status {
  width: 100px;
}

.col-actions {
  width: 130px;
}

/* 状态标签 */
.status-tag {
  display: inline-flex;
  align-items: center;
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

.status-default {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-mist);
  border: 1px solid rgba(255, 255, 255, 0.15);
}

/* 操作按钮 */
.action-btns {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: var(--radius-sm, 8px);
  cursor: pointer;
  transition: all 0.2s ease;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-mist);
}

.action-btn:hover {
  transform: scale(1.1);
}

.action-btn.view:hover {
  background: rgba(56, 189, 248, 0.2);
  color: #38bdf8;
}

.action-btn.edit:hover {
  background: rgba(251, 191, 36, 0.2);
  color: #fbbf24;
}

.action-btn.delete:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}

/* 空状态 */
.empty-state {
  padding: 80px 20px !important;
  text-align: center;
}

.data-table tbody tr:only-child td {
  text-align: center;
}

.empty-state td {
  display: table-cell !important;
}

.empty-state td::before {
  display: none;
}

.empty-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.empty-icon {
  font-size: 64px;
  color: var(--text-ghost);
  opacity: 0.5;
}

.empty-content p {
  font-size: 16px;
  color: var(--text-mist);
}

/* 分页 */
.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  flex-wrap: wrap;
  gap: 16px;
}

.page-info {
  font-size: 13px;
  color: var(--text-mist);
}

.page-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-select {
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm, 8px);
  color: var(--text-pearl);
  font-size: 13px;
  cursor: pointer;
}

.page-select:focus {
  outline: none;
  border-color: var(--teal);
}

.page-btn {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm, 8px);
  color: var(--text-pearl);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.page-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  border-color: var(--teal);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-current {
  font-size: 13px;
  color: var(--text-mist);
  padding: 0 8px;
}

/* 响应式 */
@media (max-width: 1200px) {
  .stats-panel {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-panel {
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
  }

  .stat-card {
    padding: 12px 10px;
    flex-direction: row;
    align-items: center;
    text-align: left;
    gap: 10px;
  }

  .stat-icon-wrapper {
    width: 36px;
    height: 36px;
    flex-shrink: 0;
  }

  .stat-icon {
    font-size: 18px;
  }

  .stat-body {
    flex: 1;
    min-width: 0;
  }

  .stat-value {
    font-size: 20px;
    line-height: 1.2;
  }

  .stat-label {
    font-size: 10px;
    margin-top: 2px;
  }

  .stat-glow {
    display: none;
  }

  .search-inner {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }

  .search-field {
    min-width: auto;
  }

  .filter-group {
    width: 100%;
  }

  .filter-group :deep(.el-select) {
    flex: 1;
    min-width: 100px;
  }

  .action-group {
    width: 100%;
    justify-content: space-between;
  }

  .btn {
    flex: 1;
    justify-content: center;
  }

  .table-wrapper {
    padding: 0 12px;
  }

  .data-table thead {
    display: none;
  }

  .data-table td {
    display: flex;
    padding: 12px 0;
    border: none;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .data-table td::before {
    content: attr(data-label);
    font-weight: 700;
    color: var(--text-ghost);
    text-transform: uppercase;
    font-size: 11px;
    min-width: 80px;
    margin-right: 12px;
  }

  .data-row {
    display: block;
    padding: 16px 0;
    border-bottom: 2px solid rgba(255, 255, 255, 0.08);
  }

  .col-index,
  .col-company,
  .col-contact,
  .col-phone,
  .col-products,
  .col-transport,
  .col-routes,
  .col-status,
  .col-actions {
    width: 100% !important;
    text-align: left !important;
  }

  .col-index {
    display: none !important;
  }

  .company-cell,
  .contact-cell {
    width: 100%;
  }

  .action-btns {
    width: 100%;
    justify-content: flex-start;
    gap: 16px;
  }

  .action-btn {
    width: 44px;
    height: 44px;
  }

  .pagination-bar {
    flex-direction: column;
    text-align: center;
    gap: 12px;
    padding: 16px 20px;
  }

  .page-controls {
    width: 100%;
    justify-content: space-between;
  }

  .page-select {
    min-width: 100px;
  }

  .page-btn {
    padding: 10px 20px;
  }

  /* 确保空状态在移动端居中 */
  .data-table tbody tr:only-child td.empty-state {
    display: flex !important;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center !important;
    padding: 80px 20px !important;
  }

  .empty-state td::before {
    display: none !important;
  }
}

@media (max-width: 480px) {
  .stats-panel {
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }

  .stat-card {
    padding: 10px 8px;
    gap: 8px;
  }

  .stat-icon-wrapper {
    width: 32px;
    height: 32px;
  }

  .stat-icon {
    font-size: 16px;
  }

  .stat-value {
    font-size: 18px;
  }

  .stat-label {
    font-size: 9px;
  }

  .search-row {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-group {
    flex-direction: column;
  }

  .filter-group :deep(.el-select) {
    width: 100%;
  }

  .table-header {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }

  .page-controls {
    flex-wrap: wrap;
    gap: 8px;
    justify-content: center;
  }

  .page-select {
    order: 1;
    width: 100%;
  }

  /* 确保空状态在小屏移动端居中 */
  .data-table tbody tr:only-child td.empty-state {
    display: flex !important;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center !important;
    padding: 60px 16px !important;
  }

  .empty-state td::before {
    display: none !important;
  }
}
</style>

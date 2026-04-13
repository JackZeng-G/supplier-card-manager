import axios from 'axios'

// 前后端合并后，使用相对路径即可
const api = axios.create({
  timeout: 30000
})

// 供应商相关API
export const supplierApi = {
  // 上传名片正面并OCR识别
  uploadCard(file) {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/api/suppliers/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  // 上传名片反面并OCR识别
  uploadCardBack(file) {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/api/suppliers/upload-back', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },

  // 获取供应商列表
  getList(params = {}) {
    return api.get('/api/suppliers', { params })
  },

  // 获取单个供应商
  getOne(id) {
    return api.get(`/api/suppliers/${id}`)
  },

  // 创建供应商
  create(data) {
    return api.post('/api/suppliers', data)
  },

  // 更新供应商
  update(id, data) {
    return api.put(`/api/suppliers/${id}`, data)
  },

  // 删除供应商
  delete(id) {
    return api.delete(`/api/suppliers/${id}`)
  },

  // 导出Excel
  exportExcel() {
    return api.get('/api/suppliers/export', { responseType: 'blob' })
  },

  // 获取统计数据
  getStats() {
    return api.get('/api/suppliers/stats')
  }
}

export default api

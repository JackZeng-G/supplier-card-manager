import { createApp } from 'vue'
// 按需引入 Element Plus 组件
import {
  ElButton,
  ElForm,
  ElFormItem,
  ElInput,
  ElSelect,
  ElOption,
  ElTable,
  ElTableColumn,
  ElPagination,
  ElDialog,
  ElImage,
  ElUpload,
  ElTag,
  ElEmpty,
  ElIcon,
  ElMessage,
  ElMessageBox,
  ElLoading
} from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'
import './style.css'

const app = createApp(App)

// 注册组件
const components = [
  ElButton,
  ElForm,
  ElFormItem,
  ElInput,
  ElSelect,
  ElOption,
  ElTable,
  ElTableColumn,
  ElPagination,
  ElDialog,
  ElImage,
  ElUpload,
  ElTag,
  ElEmpty,
  ElIcon
]

components.forEach(component => {
  app.component(component.name, component)
})

// 注册插件
app.use(ElLoading)

app.use(router)
app.mount('#app')
import { createRouter, createWebHistory } from 'vue-router'

// 路由懒加载 - 按需加载组件，减少首屏加载时间
const Upload = () => import('../views/Upload.vue')
const List = () => import('../views/List.vue')
const Detail = () => import('../views/Detail.vue')

const routes = [
  {
    path: '/',
    redirect: '/list'
  },
  {
    path: '/upload',
    name: 'Upload',
    component: Upload,
    meta: { title: '上传名片' }
  },
  {
    path: '/list',
    name: 'List',
    component: List,
    meta: { title: '供应商列表' }
  },
  {
    path: '/detail/:id?',
    name: 'Detail',
    component: Detail,
    meta: { title: '供应商详情' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  document.title = to.meta.title ? `${to.meta.title} - 供应商名片管理系统` : '供应商名片管理系统'
  next()
})

export default router

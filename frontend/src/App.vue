<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Document, Upload } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const menuItems = [
  { path: '/list', icon: Document, title: '供应商列表' },
  { path: '/upload', icon: Upload, title: '上传名片' }
]

const activeMenu = computed(() => route.path)

const navigate = (path) => {
  router.push(path)
}
</script>

<template>
  <div class="ocean-system">
    <!-- 动态海洋背景 -->
    <div class="ocean-canvas">
      <div class="deep-layer"></div>
      <div class="wave-layer">
        <svg class="wave wave-1" viewBox="0 0 1440 320" preserveAspectRatio="none">
            <path fill="rgba(30, 58, 95, 0.4)" d="M0,192L48,197.3C96,203,192,213,288,229.3C384,245,480,267,576,261.3C672,256,768,224,864,213.3C960,203,1056,213,1152,229.3C1248,245,1344,267,1392,277.3L1440,288L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z"></path>
          </svg>
        <svg class="wave wave-2" viewBox="0 0 1440 320" preserveAspectRatio="none">
          <path fill="rgba(20, 184, 166, 0.15)" d="M0,256L48,250.7C96,245,192,235,288,224C384,213,480,203,576,213.3C672,224,768,256,864,266.7C960,277,1056,267,1152,245.3C1248,224,1344,192,1392,176L1440,160L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z"></path>
          </svg>
      </div>
      <!-- 浮动光斑 -->
      <div class="light-orbs">
        <div class="orb orb-1"></div>
        <div class="orb orb-2"></div>
        <div class="orb orb-3"></div>
      </div>
    </div>

    <!-- 玻璃态导航栏 -->
    <header class="glass-nav">
      <div class="nav-container">
        <div class="brand">
          <div class="brand-icon">
            <svg viewBox="0 0 24 24" fill="none">
              <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </div>
          <div class="brand-text">
            <h1 class="brand-title">供应商名片管理</h1>
            <span class="brand-subtitle">Supplier Management System</span>
          </div>
        </div>

        <nav class="nav-menu">
          <button
            v-for="item in menuItems"
            :key="item.path"
            :class="['nav-link', { active: activeMenu === item.path }]"
            @click="navigate(item.path)"
          >
            <el-icon class="nav-icon"><component :is="item.icon" /></el-icon>
            <span class="nav-label">{{ item.title }}</span>
          </button>
        </nav>
      </div>
    </header>

    <!-- 主内容区 -->
    <main class="main-stage">
      <router-view />
    </main>

    <!-- 底部 -->
    <footer class="glass-footer">
      <div class="footer-line"></div>
      <span class="footer-text">© 2026 Supplier Card Management System · by JackZeng</span>
    </footer>
  </div>
</template>

<style>
/* 全局重置 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* CSS变量 - 海洋奢华主题 */
:root {
  /* 深海色系 */
  --abyss: #0a192f;
  --deep: #0d2137;
  --ocean: #1e3a5f;
  --teal: #14b8a6;
  --surface: #2d5a87;

  /* 玻璃态 */
  --glass-bg: rgba(255, 255, 255, 0.05);
  --glass-border: rgba(255, 255, 255, 0.1);
  --glass-highlight: rgba(255, 255, 255, 0.15);

  /* 文字 */
  --text-pearl: #f0f4f8;
  --text-mist: rgba(240, 244, 248, 0.7);
  --text-ghost: rgba(240, 244, 248, 0.4);

  /* 强调色 */
  --coral: #f97316;
  --amber: #fbbf24;
  --emerald: #10b981;

  /* 阴影 */
  --shadow-soft: 0 8px 32px rgba(0, 0, 0, 0.3);
  --shadow-glow: 0 0 60px rgba(20, 184, 166, 0.2);

  /* 圆角 */
  --radius-sm: 8px;
  --radius-md: 16px;
  --radius-lg: 24px;
  --radius-xl: 32px;
}

body {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
  background: var(--abyss);
  color: var(--text-pearl);
  -webkit-font-smoothing: antialiased;
}

/* 修复Element Plus下拉框在深色主题下的显示问题 */
.el-select-dropdown {
  background-color: #1e3a5f !important;
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
}

.el-select-dropdown__item {
  color: #f0f4f8 !important;
  background-color: transparent !important;
}

.el-select-dropdown__item:hover {
  background-color: #2d5a87 !important;
}

.el-select-dropdown__item.selected {
  color: #14b8a6 !important;
  font-weight: 600;
}

/* 主容器 */
.ocean-system {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
}

/* 海洋背景画布 */
.ocean-canvas {
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
}

.deep-layer {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 80% 50% at 20% 20%, rgba(30, 58, 95, 0.5) 0%, transparent 50%),
    radial-gradient(ellipse 60% 40% at 80% 80%, rgba(20, 184, 166, 0.15) 0%, transparent 50%),
    linear-gradient(180deg, var(--abyss) 0%, var(--deep) 50%, var(--ocean) 100%);
}

/* 波浪层 */
.wave-layer {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 40%;
}

.wave {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 200%;
  height: 100%;
  animation: wave-flow 20s linear infinite;
}

.wave-2 {
  animation-delay: -10s;
  animation-duration: 25s;
}

@keyframes wave-flow {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}

/* 光斑效果 */
.light-orbs {
  position: absolute;
  inset: 0;
}

.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  animation: float 20s ease-in-out infinite;
}

.orb-1 {
  width: 400px;
  height: 400px;
  background: rgba(20, 184, 166, 0.15);
  top: 10%;
  left: 10%;
  animation-delay: 0s;
}

.orb-2 {
  width: 300px;
  height: 300px;
  background: rgba(30, 58, 95, 0.3);
  top: 50%;
  right: 15%;
  animation-delay: -7s;
}

.orb-3 {
  width: 250px;
  height: 250px;
  background: rgba(249, 115, 22, 0.1);
  bottom: 20%;
  left: 30%;
  animation-delay: -14s;
}

@keyframes float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(30px, -30px) scale(1.1); }
  66% { transform: translate(-20px, 20px) scale(0.9); }
}

/* 玻璃态导航 */
.glass-nav {
  position: sticky;
  top: 0;
  z-index: 100;
  background: rgba(10, 25, 47, 0.7);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--glass-border);
}

.nav-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 32px;
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.brand {
  display: flex;
  align-items: center;
  gap: 16px;
}

.brand-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--teal), var(--surface));
  border-radius: var(--radius-md);
  color: white;
  box-shadow: 0 4px 20px rgba(20, 184, 166, 0.3);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.brand-icon:hover {
  transform: scale(1.05);
  box-shadow: 0 6px 30px rgba(20, 184, 166, 0.4);
}

.brand-icon svg {
  width: 28px;
  height: 28px;
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-title {
  font-family: 'Playfair Display', serif;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-pearl);
  letter-spacing: 0.5px;
}

.brand-subtitle {
  font-size: 10px;
  font-weight: 500;
  color: var(--text-ghost);
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-top: 2px;
}

/* 导航菜单 */
.nav-menu {
  display: flex;
  gap: 8px;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 24px;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  color: var(--text-mist);
  font-family: 'DM Sans', sans-serif;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.nav-link::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(255,255,255,0.1), transparent);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.nav-link:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
  color: var(--text-pearl);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
}

.nav-link:hover::before {
  opacity: 1;
}

.nav-link.active {
  background: linear-gradient(135deg, var(--teal), var(--surface));
  border-color: transparent;
  color: white;
  box-shadow: 0 4px 20px rgba(20, 184, 166, 0.4);
}

.nav-link.active:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 30px rgba(20, 184, 166, 0.5);
}

.nav-icon {
  font-size: 18px;
}

/* 主内容 */
.main-stage {
  flex: 1;
  padding: 32px;
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  position: relative;
  z-index: 1;
}

/* 玻璃态底部 */
.glass-footer {
  position: relative;
  z-index: 10;
  background: rgba(10, 25, 47, 0.6);
  backdrop-filter: blur(15px);
  -webkit-backdrop-filter: blur(15px);
  border-top: 1px solid var(--glass-border);
  padding: 20px 32px;
  text-align: center;
}

.footer-line {
  width: 60px;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--teal), transparent);
  margin: 0 auto 12px;
}

.footer-text {
  font-size: 12px;
  color: var(--text-ghost);
  letter-spacing: 1px;
}

/* 响应式 */
@media (max-width: 768px) {
  .nav-container {
    padding: 0 16px;
    height: 64px;
  }

  .brand-icon {
    width: 40px;
    height: 40px;
  }

  .brand-icon svg {
    width: 24px;
    height: 24px;
  }

  .brand-title {
    font-size: 18px;
  }

  .brand-subtitle {
    display: none;
  }

  .nav-menu {
    gap: 6px;
  }

  .nav-link {
    padding: 10px 16px;
    font-size: 13px;
  }

  .nav-label {
    display: none;
  }

  .main-stage {
    padding: 16px;
  }

  .glass-footer {
    padding: 16px;
  }

  .footer-text {
    font-size: 11px;
  }

  /* 禁用移动端背景动画以提升性能 */
  .wave,
  .orb {
    animation: none;
  }
}

@media (max-width: 480px) {
  .nav-container {
    padding: 0 12px;
    height: 56px;
  }

  .brand {
    gap: 10px;
  }

  .brand-icon {
    width: 36px;
    height: 36px;
    border-radius: 10px;
  }

  .brand-icon svg {
    width: 20px;
    height: 20px;
  }

  .brand-title {
    font-size: 16px;
  }

  .nav-link {
    padding: 8px 12px;
  }

  .nav-icon {
    font-size: 16px;
  }

  .main-stage {
    padding: 12px;
  }

  .glass-footer {
    padding: 12px;
  }

  .footer-line {
    width: 40px;
  }

  .footer-text {
    font-size: 10px;
    letter-spacing: 0.5px;
  }

  /* 简化移动端背景 */
  .orb {
    display: none;
  }

  .light-orbs {
    display: none;
  }
}

/* 触摸设备优化 */
@media (hover: none) and (pointer: coarse) {
  .nav-link {
    min-height: 44px;
  }

  .brand-icon {
    min-width: 44px;
    min-height: 44px;
  }
}

/* 横屏模式优化 */
@media (max-height: 500px) and (orientation: landscape) {
  .glass-nav {
    position: relative;
  }

  .nav-container {
    height: 56px;
  }
}
</style>

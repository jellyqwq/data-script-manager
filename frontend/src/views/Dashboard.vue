<template>
  <div class="layout-wrapper">
    <!-- 顶部栏 -->
    <el-header class="header">
      <div class="header-left">
        <el-button
          type="primary"
          icon="Menu"
          size="small"
          circle
          @click="toggleCollapse"
        />
        <span class="header-title">数据采集脚本管理系统</span>
      </div>
      <div class="header-right">
        <el-button type="text" @click="logout">退出登录</el-button>
      </div>
    </el-header>

    <!-- 主区域 -->
    <el-container style="flex: 1">
      <!-- 侧边栏 -->
      <el-aside :width="isCollapsed ? '64px' : '200px'" class="aside">
        <el-menu
  :default-active="active"
  :collapse="isCollapsed"
  class="el-menu-vertical-demo"
  @select="handleSelect"
>
  <el-menu-item index="script">
    <el-icon><Document /></el-icon>
    <span>脚本管理</span>
  </el-menu-item>
  <el-menu-item index="schedule">
    <el-icon><Timer /></el-icon>
    <span>任务调度</span>
  </el-menu-item>
  <el-menu-item index="nodes">
    <el-icon><Cpu /></el-icon>
    <span>节点管理</span>
  </el-menu-item>
  <el-menu-item index="logs">
    <el-icon><Memo /></el-icon>
    <span>日志监控</span>
  </el-menu-item>
  <el-menu-item index="settings">
    <el-icon><Setting /></el-icon>
    <span>系统设置</span>
  </el-menu-item>
</el-menu>

      </el-aside>

      <!-- 内容区 -->
      <el-main>
        <component :is="currentView" />
      </el-main>
    </el-container>

    <!-- 页脚 -->
    <el-footer class="footer">
      © 2025 数据采集脚本管理系统 v1.0.0
    </el-footer>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import ScriptManager from '../components/ScriptManager.vue'
import ScheduleManager from '../components/ScheduleManager.vue'
import NodeManager from '../components/NodeManager.vue'
import LogMonitor from '../components/LogMonitor.vue'
import SystemConfig from '../components/SystemConfig.vue'
import { Document, Timer, Cpu, Memo, Setting } from '@element-plus/icons-vue'

const components = {
  Document,
  Timer,
  Cpu,
  Memo,
  Setting
}

const router = useRouter()
const currentView = ref(ScriptManager)
const active = ref('script')
const isCollapsed = ref(false)

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}

const handleSelect = (key) => {
  active.value = key
  switch (key) {
    case 'script':
      currentView.value = ScriptManager
      break
    case 'schedule':
      currentView.value = ScheduleManager
      break
    case 'nodes':
      currentView.value = NodeManager
      break
    case 'logs':
      currentView.value = LogMonitor
      break
    case 'settings':
      currentView.value = SystemConfig
      break
  }
}

const logout = () => {
  localStorage.removeItem('token')
  router.push('/login')
}
</script>

<style scoped>
.layout-wrapper {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  height: 60px;
  background-color: #f5f7fa;
  border-bottom: 1px solid #ebeef5;
}

.header-title {
  margin-left: 12px;
  font-weight: bold;
  font-size: 16px;
  color: #333;
}

.footer {
  height: 40px;
  text-align: center;
  line-height: 40px;
  font-size: 13px;
  color: #999;
  background-color: #f9f9f9;
  border-top: 1px solid #ebeef5;
}

.el-menu-vertical-demo {
  border-right: none;
}

.aside {
  transition: width 0.2s;
}
</style>

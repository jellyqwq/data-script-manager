<template>
  <div>
    <div style="margin-bottom: 16px; display: flex; gap: 12px; align-items: center;">
      <el-select
        v-model="selectedScript"
        placeholder="选择脚本"
        clearable
        style="width: 200px"
        @change="handleFilterChange"
      >
        <el-option
          v-for="s in scripts"
          :key="s.id"
          :label="s.script_name"
          :value="s.id"
        />
      </el-select>

      <el-select
        v-model="selectedLevel"
        placeholder="日志级别"
        clearable
        style="width: 160px"
        @change="handleFilterChange"
      >
        <el-option label="全部" value="" />
        <el-option label="INFO" value="INFO" />
        <el-option label="ERROR" value="ERROR" />
        <el-option label="DEBUG" value="DEBUG" />
      </el-select>

      <el-button type="primary" @click="loadLogs">刷新</el-button>
      <el-button type="danger" @click="clearLogs">清空日志</el-button>
    </div>

    <el-table :data="logs" height="500" border stripe>
      <el-table-column prop="timestamp" label="时间" width="200">
        <template #default="{ row }">
          {{ formatDate(row.timestamp) }}
        </template>
      </el-table-column>

      <el-table-column prop="level" label="等级" width="100">
        <template #default="{ row }">
          <el-tag :type="levelColor(row.level)">
            {{ row.level }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="message" label="内容" />

      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button type="danger" text @click="deleteLog(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top: 12px; text-align: center"
      background
      layout="prev, pager, next"
      :current-page="page"
      :page-size="pageSize"
      :total="total"
      @current-change="handlePageChange"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { clearLogs as clearLogList, deleteLog as removeLog, getLogs } from '../api/logs'
import { getScripts } from '../api/scripts'
import { usePolling } from '../composables/usePolling'
import type { LogEntry, LogQuery } from '../types/log'
import type { Script } from '../types/script'

const scripts = ref<Script[]>([])
const logs = ref<LogEntry[]>([])
const selectedScript = ref('')
const selectedLevel = ref('')
const page = ref(1)
const pageSize = 20
const total = ref(0)

const loadScripts = async () => {
  try {
    const data = await getScripts()
    scripts.value = data.items
  } catch (err) {
    ElMessage.error('脚本列表获取失败')
    scripts.value = []
  }
}

const loadLogs = async () => {
  try {
    const params: LogQuery = {
      page: page.value,
      page_size: pageSize
    }
    if (selectedScript.value) params.script_id = selectedScript.value
    if (selectedLevel.value) params.level = selectedLevel.value

    const data = await getLogs(params)
    logs.value = data.data
    total.value = data.total
  } catch (err) {
    ElMessage.error('日志获取失败')
  }
}

const formatDate = (ts: string) => {
  return new Date(ts).toLocaleString()
}

const levelColor = (level: string) => {
  switch (level) {
    case 'INFO': return 'success'
    case 'ERROR': return 'danger'
    case 'DEBUG': return 'info'
    default: return ''
  }
}

const handleFilterChange = () => {
  page.value = 1
  void loadLogs()
}

const handlePageChange = (newPage: number) => {
  page.value = newPage
  void loadLogs()
}

const deleteLog = async (id: string) => {
  try {
    await removeLog(id)
    ElMessage.success('日志已删除')
    void loadLogs()
  } catch (err) {
    ElMessage.error('删除失败')
  }
}

const clearLogs = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有日志吗？', '警告', { type: 'warning' })
    await clearLogList()
    ElMessage.success('日志已清空')
    void loadLogs()
  } catch (err) {
    ElMessage.error('清空失败')
  }
}

const { start: startLogPolling, stop: stopLogPolling } = usePolling(loadLogs, 10000)

onMounted(async () => {
  await loadScripts()
  await loadLogs()
  startLogPolling()
})

defineExpose({
  stopLogTimer: stopLogPolling
})
</script>

<style scoped>
.el-tag {
  text-transform: uppercase;
}
</style>

<template>
  <div>
    <div style="margin-bottom: 16px; display: flex; gap: 12px; align-items: center;">
      <el-select v-model="selectedScript" placeholder="选择脚本" clearable style="width: 200px">
        <el-option
          v-for="s in scripts"
          :key="s.id"
          :label="s.script_name"
          :value="s.id"
        />
      </el-select>

      <el-select v-model="selectedLevel" placeholder="日志级别" clearable style="width: 160px">
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

<script setup>
import { ref, onMounted, onBeforeUnmount, defineExpose } from 'vue'
import axios from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const scripts = ref([])
const logs = ref([])
const selectedScript = ref('')
const selectedLevel = ref('')
const page = ref(1)
const pageSize = 20
const total = ref(0)
const logTimer = ref(null)

const loadScripts = async () => {
  try {
    const res = await axios.get('/auth/scripts');
    console.log("Response from /auth/scripts:", res.data); // 打印响应

    if (res.data && res.data.items && Array.isArray(res.data.items)) {
      scripts.value = res.data.items;
    } else if (res.data && Array.isArray(res.data)) {
      scripts.value = res.data; // 假设直接返回的是数组
    } else {
      console.error("Error: /auth/scripts returned unexpected format:", res.data);
      scripts.value = [];
    }
  } catch (err) {
    ElMessage.error('脚本列表获取失败');
    console.error('脚本列表获取失败:', err);
    scripts.value = [];
  }
};

const loadLogs = async () => {
  try {
    const params = {
      page: page.value,
      page_size: pageSize
    }
    if (selectedScript.value) params.script_id = selectedScript.value
    if (selectedLevel.value) params.level = selectedLevel.value

    const res = await axios.get('/auth/logs', { params })
    logs.value = res.data.data
    total.value = res.data.total
  } catch (err) {
    ElMessage.error('日志获取失败')
  }
}

const formatDate = (ts) => {
  return new Date(ts).toLocaleString()
}

const levelColor = (level) => {
  switch (level) {
    case 'INFO': return 'success'
    case 'ERROR': return 'danger'
    case 'DEBUG': return 'info'
    default: return ''
  }
}

const handlePageChange = (newPage) => {
  page.value = newPage
  loadLogs()
}

const deleteLog = async (id) => {
  try {
    await axios.delete(`/auth/logs/${id}`)
    ElMessage.success('日志已删除')
    loadLogs()
  } catch (err) {
    ElMessage.error('删除失败')
  }
}

const clearLogs = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有日志吗？', '警告', { type: 'warning' })
    await axios.delete('/auth/logs')
    ElMessage.success('日志已清空')
    loadLogs()
  } catch (err) {
    ElMessage.error('清空失败')
  }
}

onMounted(() => {
  loadScripts()
  loadLogs()
  logTimer.value = setInterval(loadLogs, 10000) // 每 10 秒自动刷新
})

onBeforeUnmount(() => {
  if (logTimer.value) {
    clearInterval(logTimer.value)
    logTimer.value = null
  }
})

function stopLogTimer() {
  if (logTimer.value) {
    clearInterval(logTimer.value)
    logTimer.value = null
  }
}
defineExpose({
  stopLogTimer
})
</script>

<style scoped>
.el-tag {
  text-transform: uppercase;
}
</style>

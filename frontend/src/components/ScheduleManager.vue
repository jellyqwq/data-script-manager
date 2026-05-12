<template>
  <div>
    <el-button type="primary" @click="openDialog">新增任务</el-button>

    <el-table :data="schedules" style="margin-top: 20px;">
      <el-table-column prop="script_name" label="脚本名称" width="200" />
      <el-table-column prop="node_name" label="执行节点" width="180" />
      <el-table-column prop="cron" label="Cron表达式" />
      <el-table-column prop="enabled" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="320">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" @click="toggleEnable(row)">{{ row.enabled ? '停用' : '启用' }}</el-button>
          <el-button size="small" type="danger" @click="deleteSchedule(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :title="form.id ? '编辑任务' : '新建任务'" v-model="dialogVisible">
      <el-form :model="form" label-width="90px">
        <el-form-item label="脚本">
          <el-select v-model="form.script_id" placeholder="请选择脚本" @change="handleScriptChange">
            <el-option v-for="item in scripts" :key="item.id" :label="item.script_name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="执行节点">
          <el-select v-model="form.node_id" placeholder="请选择节点">
            <el-option v-for="item in nodes" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
          <div v-if="!form.node_id" style="color: #f56c6c; font-size: 12px; margin-top: 4px;">未选择节点，将随机分配节点</div>
        </el-form-item>

        <el-form-item label="Cron表达式">
          <cron-element-plus v-model="value" :button-props="{ type: 'primary' }" @error="cronError = $event" />
          <p class="text-lightest pt-2">cron expression: {{ value }}</p>
        </el-form-item>

        <el-form-item label="变量组运行">
          <el-switch v-model="form.use_env_groups" />
        </el-form-item>

        <el-form-item v-if="form.use_env_groups" label="变量组">
          <el-select v-model="form.env_group_ids" multiple placeholder="请选择变量组" style="width: 100%">
            <el-option
              v-for="item in envGroups"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
          <div style="color: #909399; font-size: 12px; margin-top: 4px;">
            开启后，同一脚本会按选中的变量组分别执行。
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitSchedule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'
import '@vue-js-cron/light/dist/light.css'
import { CronElementPlus } from '@vue-js-cron/element-plus'
import { getEnvGroups } from '../api/envGroups'

const schedules = ref([])
const scripts = ref([])
const nodes = ref([])
const envGroups = ref([])

const dialogVisible = ref(false)
const form = ref({
  id: '',
  script_id: '',
  cron: '* * * * *',
  node_id: '',
  use_env_groups: false,
  env_group_ids: []
})
const value = ref('* * * * *')
const cronError = ref(null)

const formatDate = (str) => new Date(str).toLocaleString()

const loadScripts = async () => {
  try {
    const res = await axios.get('/auth/scripts')
    scripts.value = res.data?.items ?? []
  } catch (error) {
    console.error('Error loading scripts:', error)
    scripts.value = []
  }
}

const loadNodes = async () => {
  try {
    const res = await axios.get('/auth/nodes')
    nodes.value = res.data ?? []
  } catch (error) {
    console.error('Error loading nodes:', error)
    nodes.value = []
  }
}

const loadSchedules = async () => {
  const res = await axios.get('/auth/schedules')
  const scriptMap = {}
  scripts.value.forEach(s => (scriptMap[s.id] = s.script_name))
  const nodeMap = {}
  nodes.value.forEach(n => (nodeMap[n.id] = n.name))
  schedules.value = res.data.map(item => ({
    ...item,
    script_name: scriptMap[item.script_id] || '未知脚本',
    node_name: nodeMap[item.node_id] || '未指定节点'
  }))
}

const openDialog = (row = null) => {
  if (row) {
    form.value = {
      id: row.id,
      script_id: row.script_id,
      cron: row.cron || '* * * * *',
      node_id: row.node_id || '',
      use_env_groups: row.use_env_groups || false,
      env_group_ids: row.env_group_ids || []
    }
    value.value = row.cron || '* * * * *'
    loadEnvGroups(row.script_id)
  } else {
    form.value = {
      id: '',
      script_id: '',
      cron: '* * * * *',
      node_id: '',
      use_env_groups: false,
      env_group_ids: []
    }
    value.value = '* * * * *'
    envGroups.value = []
  }
  dialogVisible.value = true
}

const loadEnvGroups = async (scriptId) => {
  if (!scriptId) {
    envGroups.value = []
    return
  }
  try {
    const data = await getEnvGroups(scriptId)
    envGroups.value = data.items
  } catch (error) {
    envGroups.value = []
  }
}

const handleScriptChange = async (scriptId) => {
  form.value.env_group_ids = []
  await loadEnvGroups(scriptId)
}

const submitSchedule = async () => {
  try {
    form.value.cron = value.value
    const payload = { ...form.value }
    if (!payload.node_id) delete payload.node_id
    if (!payload.use_env_groups) payload.env_group_ids = []
    if (form.value.id) {
      await axios.put(`/auth/schedules/${form.value.id}`, payload)
      ElMessage.success('修改成功')
    } else {
      await axios.post('/auth/schedules', payload)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    await loadScripts()
    await loadNodes()
    await loadSchedules()
  } catch (err) {
    ElMessage.error(err.response?.data?.error || '操作失败')
  }
}

const deleteSchedule = async (id) => {
  await ElMessageBox.confirm('确认删除该任务？', '提示', { type: 'warning' })
  await axios.delete(`/auth/schedules/${id}`)
  ElMessage.success('删除成功')
  await loadSchedules()
}

const toggleEnable = async (row) => {
  await axios.put(`/auth/schedules/${row.id}`, { enabled: !row.enabled })
  ElMessage.success(row.enabled ? '已停用' : '已启用')
  await loadSchedules()
}

onMounted(async () => {
  await loadScripts()
  await loadNodes()
  await loadSchedules()
})
</script>

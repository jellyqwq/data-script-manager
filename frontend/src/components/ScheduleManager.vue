<template>
  <div>
    <el-button type="primary" @click="openDialog">新增任务</el-button>

    <el-table :data="schedules" style="margin-top: 20px;">
      <el-table-column prop="script_name" label="脚本名称" width="200" />
      <el-table-column prop="cron" label="Cron表达式" />
      <el-table-column prop="enabled" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '启用' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="next_run" label="下次运行" width="180">
        <template #default="{ row }">
          {{ formatDate(row.next_run) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="320">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" @click="toggleEnable(row)">
            {{ row.enabled ? '停用' : '启用' }}
          </el-button>
          <el-button size="small" type="danger" @click="deleteSchedule(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :title="form.id ? '编辑任务' : '新建任务'" v-model="dialogVisible">
      <el-form :model="form" label-width="90px">
        <el-form-item label="脚本">
          <el-select v-model="form.script_id" placeholder="请选择脚本">
            <el-option
              v-for="item in scripts"
              :key="item.id"
              :label="item.script_name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="Cron表达式">
          <cron-element-plus
            v-model="value"
            :button-props="{ type: 'primary' }"
            @error="error=$event" />

          <p class="text-lightest pt-2">cron expression: {{value}}</p>
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

const schedules = ref([])
const scripts = ref([])

const dialogVisible = ref(false)
const form = ref({
  id: '',
  script_id: '',
  cron: '* * * * *'  // 初始默认值很重要
})

const formatDate = (str) => {
  return new Date(str).toLocaleString()
}

const loadScripts = async () => {
  const res = await axios.get('/auth/scripts')
  scripts.value = res.data
}

const loadSchedules = async () => {
  const res = await axios.get('/auth/schedules')
  const scriptMap = {}
  scripts.value.forEach(s => (scriptMap[s.id] = s.script_name))
  schedules.value = res.data.map(item => ({
    ...item,
    script_name: scriptMap[item.script_id] || '未知脚本'
  }))
}

const openDialog = (row = null) => {
  if (row) {
    form.value = {
      id: row.id,
      script_id: row.script_id,
      cron: row.cron || '* * * * *'
    }
  } else {
    form.value = {
      id: '',
      script_id: '',
      cron: '* * * * *'
    }
  }
  dialogVisible.value = true
}



const submitSchedule = async () => {
  try {
    if (form.value.id) {
      console.log(form)
      console.log(form.cron)
      await axios.put(`/auth/schedules/${form.value.id}`, {
        cron: form.value.cron
      })
      ElMessage.success('修改成功')
    } else {
      await axios.post('/auth/schedules', form.value)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
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
  await loadSchedules()
})
</script>

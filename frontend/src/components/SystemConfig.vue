<template>
  <div>
    <el-button type="primary" @click="openDialog">新增变量</el-button>

    <el-table
      :data="paginatedEnvVars"
      style="margin-top: 20px"
      @sort-change="handleSortChange"
      :default-sort="{ prop: 'key', order: 'ascending' }"
    >
      <el-table-column prop="key" label="变量名" sortable />
      <el-table-column prop="value" label="变量值" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteEnvVar(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :page-sizes="[10, 20, 50, 100]"
      :total="totalEnvVars"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
      style="margin-top: 20px; display: flex; justify-content: center;"
    >
    </el-pagination>

    <el-dialog :title="form._id ? '编辑变量' : '新增变量'" v-model="dialogVisible">
      <el-form :model="form" label-width="80px">
        <el-form-item label="变量名">
          <el-input v-model="form.key" />
        </el-form-item>
        <el-form-item label="变量值">
          <el-input v-model="form.value" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEnvVar">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const envVars = ref([])
const dialogVisible = ref(false)
const form = ref({ _id: '', key: '', value: '' })

const currentPage = ref(1)
const pageSize = ref(10)
const totalEnvVars = ref(0)
const sortBy = ref('key')
const sortOrder = ref('asc')

const loadEnvVars = async (page = currentPage.value, size = pageSize.value, sortField = sortBy.value, sortDirection = sortOrder.value) => {
  try {
    const res = await axios.get(`/auth/env-vars?page=${page}&pageSize=${size}&sortBy=${sortField}&sortOrder=${sortDirection}`)
    envVars.value = res.data.items // 假设你的后端返回的数据结构是 { items: [], total: number }
    totalEnvVars.value = res.data.total
  } catch (err) {
    ElMessage.error('加载环境变量失败')
    console.error('加载环境变量失败:', err)
  }
}

const paginatedEnvVars = computed(() => {
  // 如果后端已经做了分页和排序，则不需要前端再处理
  return envVars.value
})

const openDialog = (row = null) => {
  if (row) {
    form.value = { ...row }
  } else {
    form.value = { _id: '', key: '', value: '' }
  }
  dialogVisible.value = true
}

const saveEnvVar = async () => {
  try {
    if (form.value.id) {
      await axios.put(`/auth/env-vars/${form.value.id}`, form.value)
    } else {
      await axios.post('/auth/env-vars', form.value)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await loadEnvVars()
  } catch (err) {
    ElMessage.error(err.response?.data?.error || '保存失败')
  }
}

const deleteEnvVar = async (id) => {
  await ElMessageBox.confirm('确认删除该变量？', '提示', { type: 'warning' })
  await axios.delete(`/auth/env-vars/${id}`)
  ElMessage.success('删除成功')
  await loadEnvVars()
}

const handleSizeChange = (newSize) => {
  pageSize.value = newSize
  currentPage.value = 1
  loadEnvVars()
}

const handleCurrentChange = (newPage) => {
  currentPage.value = newPage
  loadEnvVars()
}

const handleSortChange = ({ prop, order }) => {
  sortBy.value = prop
  sortOrder.value = order === 'ascending' ? 'asc' : 'desc'
  loadEnvVars()
}

onMounted(loadEnvVars)
</script>
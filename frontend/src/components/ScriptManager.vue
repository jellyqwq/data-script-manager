<template>
  <div>
    <el-button type="primary" @click="openDialog()">新增脚本</el-button>

    <el-table :data="paginatedScripts" style="width: 100%; margin-top: 20px;">
      <el-table-column prop="script_name" label="名称" width="200" />
      <el-table-column prop="description" label="说明" />
      <el-table-column prop="created_at" label="创建时间" width="200">
        <template #default="scope">
          {{ formatDate(scope.row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column prop="last_modified" label="修改时间" width="200">
        <template #default="scope">
          {{ formatDate(scope.row.last_modified) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="scope">
          <el-button size="small" type="primary" @click="openDialog(scope.row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteScript(scope.row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="currentPage"
      v-model:page-size="pageSize"
      :page-sizes="[10, 20, 50, 100]"
      :total="totalScripts"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
      style="margin-top: 20px; display: flex; justify-content: center;"
    >
    </el-pagination>

    <el-dialog :title="form.id ? '编辑脚本' : '新建脚本'" v-model="dialogVisible" width="600px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="form.script_name" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.description" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="10"
            placeholder="请输入脚本内容"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitScript">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const scripts = ref([])
const dialogVisible = ref(false)
const form = ref({
  id: '',
  script_name: '',
  description: '',
  content: ''
})

const currentPage = ref(1)
const pageSize = ref(10)
const totalScripts = ref(0)

const loadScripts = async (page = currentPage.value, size = pageSize.value) => {
  try {
    const res = await axios.get(`/auth/scripts?page=${page}&pageSize=${size}`)
    scripts.value = res.data.items // 假设你的后端返回的数据结构是 { items: [], total: number }
    totalScripts.value = res.data.total
  } catch (err) {
    ElMessage.error('加载脚本列表失败')
    console.error('加载脚本列表失败:', err)
  }
}

const paginatedScripts = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return scripts.value.slice(start, end)
})

const formatDate = (ms) => {
  const date = new Date(ms)
  return date.toLocaleString()
}

const openDialog = (row = null) => {
  if (row) {
    form.value = { ...row }
  } else {
    form.value = {
      id: '',
      script_name: '',
      description: '',
      content: ''
    }
  }
  dialogVisible.value = true
}

const submitScript = async () => {
  try {
    if (form.value.id) {
      await axios.put(`/auth/scripts/${form.value.id}`, form.value)
      ElMessage.success('脚本更新成功')
    } else {
      await axios.post('/auth/scripts', form.value)
      ElMessage.success('脚本创建成功')
    }
    dialogVisible.value = false
    await loadScripts() // 重新加载数据，会回到第一页
  } catch (err) {
    ElMessage.error(err.response?.data?.error || '保存失败')
  }
}

const deleteScript = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该脚本吗？', '提示', {
      type: 'warning'
    })
    await axios.delete(`/auth/scripts/${id}`)
    ElMessage.success('删除成功')
    await loadScripts() // 重新加载数据，会回到第一页
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(err.response?.data?.error || '删除失败')
    }
  }
}

const handleSizeChange = (newSize) => {
  pageSize.value = newSize
  currentPage.value = 1 // 切换每页条数时，回到第一页
  loadScripts()
}

const handleCurrentChange = (newPage) => {
  currentPage.value = newPage
  loadScripts()
}

onMounted(() => {
  loadScripts()
})
</script>

<style scoped>
.el-button + .el-button {
  margin-left: 8px;
}
</style>
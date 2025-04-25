<template>
  <div>
    <el-button type="primary" @click="openEditDialog(null)">新增节点</el-button>

    <el-table :data="nodes" style="margin-top: 20px;">
      <el-table-column prop="name" label="节点名称" width="180" />
      <el-table-column prop="address" label="IP 地址" width="160" />
      <el-table-column prop="cpu_usage" label="CPU使用率" width="120">
        <template #default="{ row }">{{ row.cpu_usage?.toFixed(2) }}%</template>
      </el-table-column>
      <el-table-column prop="mem_usage" label="内存使用率" width="120">
        <template #default="{ row }">{{ row.mem_usage?.toFixed(2) }}%</template>
      </el-table-column>
      <el-table-column prop="disk_usage" label="磁盘使用率" width="120">
        <template #default="{ row }">{{ row.disk_usage?.toFixed(2) }}%</template>
      </el-table-column>
      <el-table-column prop="online" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.online ? 'success' : 'info'">
            {{ row.online ? '在线' : '离线' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="180">
        <template #default="{ row }">
          {{ new Date(row.updated_at).toLocaleString() }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteNode(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 编辑弹窗 -->
    <el-dialog :title="editingNode.id ? '编辑节点' : '新增节点'" v-model="dialogVisible">
      <el-form :model="editingNode" label-width="80px">
        <el-form-item label="节点名称">
          <el-input v-model="editingNode.name" />
        </el-form-item>
        <el-form-item label="IP 地址">
          <el-input v-model="editingNode.address" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveNode">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '../api'
import { ElMessage, ElMessageBox } from 'element-plus'

const nodes = ref([])
const dialogVisible = ref(false)
const editingNode = ref({
  id: '',
  name: '',
  address: ''
})

const loadNodes = async () => {
  const res = await axios.get('/auth/nodes')
  nodes.value = res.data
}

const openEditDialog = (node) => {
  if (node) {
    editingNode.value = { id: node.id, name: node.name, address: node.address }
  } else {
    editingNode.value = { id: '', name: '', address: '' }
  }
  dialogVisible.value = true
}

const saveNode = async () => {
  try {
    if (editingNode.value.id) {
      await axios.put(`/auth/nodes/${editingNode.value.id}`, {
        name: editingNode.value.name,
        address: editingNode.value.address
      })
      ElMessage.success('更新成功')
    } else {
      await axios.post('/auth/nodes', {
        name: editingNode.value.name,
        address: editingNode.value.address
      })
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    await loadNodes()
  } catch (err) {
    ElMessage.error(err.response?.data?.error || '操作失败')
  }
}

const deleteNode = async (id) => {
  try {
    await ElMessageBox.confirm('确认删除该节点？', '提示', { type: 'warning' })
    await axios.delete(`/auth/nodes/${id}`)
    ElMessage.success('删除成功')
    await loadNodes()
  } catch (err) {
    ElMessage.error(err.response?.data?.error || '删除失败')
  }
}

onMounted(() => {
  loadNodes()
  setInterval(loadNodes, 5000)
})
</script>

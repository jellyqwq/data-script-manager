<template>
  <div>
    <el-button type="primary" @click="openDialog()">新增脚本</el-button>

    <el-table :data="scripts" style="width: 100%; margin-top: 20px;">
      <el-table-column prop="script_name" label="名称" width="200" />
      <el-table-column prop="original_filename" label="文件名" width="220" />
      <el-table-column prop="language" label="类型" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.language">{{ row.language }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="sha1" label="SHA1" min-width="260">
        <template #default="{ row }">
          <el-text v-if="row.sha1" class="sha1-text" truncated>{{ row.sha1 }}</el-text>
        </template>
      </el-table-column>
      <el-table-column prop="size" label="大小" width="110">
        <template #default="{ row }">
          {{ formatSize(row.size) }}
        </template>
      </el-table-column>
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

    <el-dialog :title="form.id ? '编辑脚本' : '上传脚本'" v-model="dialogVisible" width="680px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="名称">
          <el-input v-model="form.script_name" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.description" />
        </el-form-item>
        <el-form-item label="脚本文件" required>
          <el-upload
            drag
            action="#"
            :auto-upload="false"
            :limit="1"
            :file-list="fileList"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">拖拽脚本到这里，或 <em>点击选择文件</em></div>
            <template #tip>
              <div class="el-upload__tip">支持 .py、.sh、.js，上传后后端会计算 SHA1 并记录元数据。</div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item v-if="form.sha1" label="当前 SHA1">
          <el-text class="sha1-text" truncated>{{ form.sha1 }}</el-text>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitScript">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadFile, UploadUserFile } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { createScript, deleteScript as removeScript, getScripts, updateScript } from '../api/scripts'
import type { Script } from '../types/script'

const scripts = ref<Script[]>([])
const dialogVisible = ref(false)
const form = ref({
  id: '',
  script_name: '',
  description: '',
  sha1: ''
})
const selectedFile = ref<File | null>(null)
const fileList = ref<UploadUserFile[]>([])

const currentPage = ref(1)
const pageSize = ref(10)
const totalScripts = ref(0)

const loadScripts = async (page = currentPage.value, size = pageSize.value) => {
  try {
    const data = await getScripts({ page, pageSize: size })
    scripts.value = data.items
    totalScripts.value = data.total
  } catch (err) {
    ElMessage.error('加载脚本列表失败')
  }
}

const formatDate = (ms?: string) => {
  if (!ms) return '-'
  const date = new Date(ms)
  return date.toLocaleString()
}

const formatSize = (size?: number) => {
  if (!size) return '-'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

const openDialog = (row: Script | null = null) => {
  selectedFile.value = null
  fileList.value = []
  if (row) {
    form.value = {
      id: row.id,
      script_name: row.script_name,
      description: row.description ?? '',
      sha1: row.sha1 ?? ''
    }
  } else {
    form.value = {
      id: '',
      script_name: '',
      description: '',
      sha1: ''
    }
  }
  dialogVisible.value = true
}

const handleFileChange = (uploadFile: UploadFile) => {
  selectedFile.value = uploadFile.raw ?? null
  if (!form.value.script_name && uploadFile.name) {
    form.value.script_name = uploadFile.name.replace(/\.[^.]+$/, '')
  }
}

const handleFileRemove = () => {
  selectedFile.value = null
}

const submitScript = async () => {
  try {
    if (!form.value.id && !selectedFile.value) {
      ElMessage.warning('请先选择脚本文件')
      return
    }

    const payload = {
      scriptName: form.value.script_name,
      description: form.value.description,
      file: selectedFile.value
    }

    if (form.value.id) {
      await updateScript(form.value.id, payload)
      ElMessage.success('脚本更新成功')
    } else {
      await createScript(payload)
      ElMessage.success('脚本创建成功')
    }
    dialogVisible.value = false
    await loadScripts()
  } catch (err) {
    const message = err instanceof Error ? err.message : '保存失败'
    ElMessage.error(message)
  }
}

const deleteScript = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除该脚本吗？', '提示', {
      type: 'warning'
    })
    await removeScript(id)
    ElMessage.success('删除成功')
    await loadScripts()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSizeChange = (newSize: number) => {
  pageSize.value = newSize
  currentPage.value = 1
  void loadScripts()
}

const handleCurrentChange = (newPage: number) => {
  currentPage.value = newPage
  void loadScripts()
}

onMounted(() => {
  loadScripts()
})
</script>

<style scoped>
.el-button + .el-button {
  margin-left: 8px;
}

.sha1-text {
  max-width: 100%;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
}
</style>

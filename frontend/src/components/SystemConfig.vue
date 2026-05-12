<template>
  <div>
    <div class="toolbar">
      <el-select
        v-model="selectedScriptId"
        placeholder="选择脚本"
        clearable
        filterable
        style="width: 260px"
        @change="handleScriptChange"
      >
        <el-option
          v-for="script in scripts"
          :key="script.id"
          :label="script.script_name"
          :value="script.id"
        />
      </el-select>

      <el-button type="primary" :disabled="!selectedScriptId" @click="openDialog()">新增变量组</el-button>
      <el-button @click="loadGroups">刷新</el-button>
    </div>

    <el-table :data="groups" style="margin-top: 20px">
      <el-table-column prop="name" label="变量组" width="180" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'info'">
            {{ row.enabled ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="变量数量" width="100">
        <template #default="{ row }">
          {{ row.vars?.length ?? 0 }}
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" width="190">
        <template #default="{ row }">
          {{ formatDate(row.updated_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteGroup(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="selectedScriptId && groups.length === 0" description="当前脚本还没有变量组" />
    <el-empty v-if="!selectedScriptId" description="请先选择脚本，再配置变量组" />

    <el-dialog :title="form.id ? '编辑变量组' : '新增变量组'" v-model="dialogVisible" width="760px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="所属脚本">
          <el-select v-model="form.script_id" placeholder="选择脚本" filterable style="width: 100%">
            <el-option
              v-for="script in scripts"
              :key="script.id"
              :label="script.script_name"
              :value="script.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="组名称">
          <el-input v-model="form.name" placeholder="例如：我的账号、朋友A、测试账号" />
        </el-form-item>
        <el-form-item label="是否启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="变量">
          <div class="env-pairs">
            <div v-for="(item, index) in form.vars" :key="index" class="env-row">
              <el-input v-model="item.key" placeholder="变量名，如 COOKIE" />
              <el-input v-model="item.value" placeholder="变量值" show-password />
              <el-button type="danger" text @click="removePair(index)">删除</el-button>
            </div>
            <el-button @click="addPair">新增变量</el-button>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveGroup">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createEnvGroup, deleteEnvGroup, getEnvGroups, updateEnvGroup } from '../api/envGroups'
import { getScripts } from '../api/scripts'
import type { EnvGroup, EnvGroupPayload, EnvPair } from '../types/envGroup'
import type { Script } from '../types/script'

const scripts = ref<Script[]>([])
const groups = ref<EnvGroup[]>([])
const selectedScriptId = ref('')
const dialogVisible = ref(false)
const form = ref<EnvGroupPayload & { id: string }>({
  id: '',
  script_id: '',
  name: '',
  enabled: true,
  vars: [{ key: '', value: '' }],
})

const loadScripts = async () => {
  const data = await getScripts({ page: 1, pageSize: 100 })
  scripts.value = data.items
  if (!selectedScriptId.value && scripts.value.length > 0) {
    selectedScriptId.value = scripts.value[0].id
  }
}

const loadGroups = async () => {
  if (!selectedScriptId.value) {
    groups.value = []
    return
  }

  try {
    const data = await getEnvGroups(selectedScriptId.value)
    groups.value = data.items
  } catch (err) {
    ElMessage.error('加载变量组失败')
  }
}

const handleScriptChange = async () => {
  await loadGroups()
}

const openDialog = (row: EnvGroup | null = null) => {
  if (row) {
    form.value = {
      id: row.id,
      script_id: row.script_id,
      name: row.name,
      enabled: row.enabled,
      vars: clonePairs(row.vars),
    }
  } else {
    form.value = {
      id: '',
      script_id: selectedScriptId.value,
      name: '',
      enabled: true,
      vars: [{ key: '', value: '' }],
    }
  }
  dialogVisible.value = true
}

const addPair = () => {
  form.value.vars.push({ key: '', value: '' })
}

const removePair = (index: number) => {
  form.value.vars.splice(index, 1)
  if (form.value.vars.length === 0) {
    addPair()
  }
}

const saveGroup = async () => {
  if (!form.value.script_id) {
    ElMessage.warning('请先选择脚本')
    return
  }
  if (!form.value.name.trim()) {
    ElMessage.warning('请填写变量组名称')
    return
  }

  const payload: EnvGroupPayload = {
    script_id: form.value.script_id,
    name: form.value.name,
    enabled: form.value.enabled,
    vars: form.value.vars.filter((item) => item.key.trim()),
  }

  try {
    if (form.value.id) {
      await updateEnvGroup(form.value.id, payload)
      ElMessage.success('变量组已更新')
    } else {
      await createEnvGroup(payload)
      ElMessage.success('变量组已创建')
    }
    dialogVisible.value = false
    selectedScriptId.value = payload.script_id
    await loadGroups()
  } catch (err) {
    ElMessage.error('保存变量组失败')
  }
}

const deleteGroup = async (id: string) => {
  try {
    await ElMessageBox.confirm('确认删除该变量组？', '提示', { type: 'warning' })
    await deleteEnvGroup(id)
    ElMessage.success('删除成功')
    await loadGroups()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const formatDate = (value?: string) => {
  return value ? new Date(value).toLocaleString() : '-'
}

const clonePairs = (pairs: EnvPair[] = []) => {
  if (pairs.length === 0) return [{ key: '', value: '' }]
  return pairs.map((item) => ({ key: item.key, value: item.value }))
}

onMounted(async () => {
  try {
    await loadScripts()
    await loadGroups()
  } catch (err) {
    ElMessage.error('初始化环境变量组失败')
  }
})
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
}

.env-pairs {
  width: 100%;
}

.env-row {
  display: grid;
  grid-template-columns: 1fr 2fr auto;
  gap: 8px;
  margin-bottom: 8px;
}
</style>

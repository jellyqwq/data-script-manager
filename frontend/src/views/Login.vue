<template>
  <div class="auth-wrapper">
    <div class="auth-box">
      <h2 class="title">数据采集脚本管理系统</h2>
      <el-form :model="form" label-width="80px">
        <el-form-item label="账号">
          <el-input v-model="form.username" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input type="password" v-model="form.password" show-password placeholder="请输入密码" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleLogin" style="width: 100%">登录</el-button>
        </el-form-item>
        <el-form-item class="action-row">
          <el-button type="text" @click="goRegister">注册账号</el-button>
          <el-button type="text" @click="goForget">忘记密码？</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import axios from '../api'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()
const form = reactive({ username: '', password: '' })

const handleLogin = async () => {
  try {
    const res = await axios.post('/login', form)
    localStorage.setItem('token', res.data.token)
    ElMessage.success('登录成功')
    localStorage.setItem('token', res.data.token)
    router.push('/dashboard')
  } catch (err) {
    ElMessage.error(err.response?.data?.error || '登录失败')
  }
}

const goRegister = () => router.push('/register')
const goForget = () => router.push('/forgot')
</script>

<style scoped>
@import '../styles/auth.css';

.action-row {
  display: flex;
  justify-content: space-between;
  padding: 0 8px;
}
</style>

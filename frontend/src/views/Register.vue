<template>
    <div class="auth-wrapper">
      <div class="auth-box">
        <h2 class="title">注册账号</h2>
        <el-form :model="form" :rules="rules" ref="registerForm" label-width="80px">
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="form.email" placeholder="请输入邮箱" />
          </el-form-item>
          <el-form-item label="验证码" prop="code">
        <el-row style="width: 100%;" gutter="8">
            <el-col :span="16">
            <el-input v-model="form.code" placeholder="请输入验证码" />
            </el-col>
            <el-col :span="8">
            <el-button @click="sendCode" type="primary" style="width: 100%;">发送验证码</el-button>
            </el-col>
        </el-row>
        </el-form-item>
          <el-form-item label="密码" prop="password">
            <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirm">
            <el-input v-model="form.confirm" type="password" show-password placeholder="请再次输入密码" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="submitRegister" style="width: 100%">注册</el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </template>
  
  <script setup>
  import { reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import axios from '../api' 
  import { useRouter } from 'vue-router'
  const router = useRouter()
  
  const form = reactive({
    email: '',
    code: '',
    password: '',
    confirm: ''
  })
  const registerForm = ref()
  
  const rules = {
    email: [{ required: true, message: '请输入邮箱', trigger: 'blur' }],
    code: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
    password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
    confirm: [
      { required: true, message: '请确认密码', trigger: 'blur' },
      {
        validator: (_, value) => value === form.password,
        message: '两次输入密码不一致',
        trigger: 'blur'
      }
    ]
  }
  
  const sendCode = async () => {
    if (!form.email) {
      return ElMessage.warning('请先输入邮箱')
    }
    try {
        await axios.post('/send-code', {
          email: form.email,
          scene: 'register' // ✅ 注册验证码
        })
        ElMessage.success('验证码已发送，请查收邮箱')
    } catch (err) {
        ElMessage.error(err.response?.data?.error || '验证码发送失败')
    }
    }

    const submitRegister = async () => {
    const valid = await registerForm.value.validate()
    if (!valid) return

    try {
        await axios.post('/register', {
        email: form.email,
        code: form.code,
        password: form.password
        })

        ElMessage.success('注册成功，请前往登录')
        router.push('/login')  // ✅ 注册成功后跳转登录页

    } catch (err) {
        const msg = err.response?.data?.error || '注册失败'
        ElMessage.error(msg)
    }
    }
  
    
  
  </script>
  
  <style scoped>
  @import '../styles/auth.css';
  </style>
  
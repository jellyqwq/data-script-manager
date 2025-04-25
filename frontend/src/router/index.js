import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import Register from '../views/Register.vue'
import Forgot from '../views/Forgot.vue'

const routes = [
  { path: '/login', component: Login },
  { path: '/register', component: Register },
  { path: '/forgot', component: Forgot },
  {
    path: '/dashboard',
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  { path: '/', redirect: '/login' },
  // 🚨 放在最后，匹配所有未定义路径
  {
    path: '/:pathMatch(.*)*',
    redirect: '/login'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  console.log('[守卫] token =', token)

  if (to.meta.requiresAuth && !token) {
    console.warn('[守卫] 未登录，跳转 login')
    next('/login')
  } else {
    console.log('[守卫] 有 token，放行')
    next()
  }
})


export default router

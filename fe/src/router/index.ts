import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/@login',
    },
    {
      path: '/@login',
      name: 'Login',
      component: () => import('@/views/login/index.vue'),
      meta: {
        requiresAuth: false,
        title: '登录',
      },
    },
    {
      path: '/@dashboard',
      component: () => import('@/layouts/BaseLayout.vue'),
      meta: {
        requiresAuth: true,
      },
      children: [
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/index.vue'),
          meta: {
            title: '仪表盘',
            requiresAuth: true,
          },
        },
        {
          path: 'users',
          name: 'Users',
          component: () => import('@/views/dashboard/users/index.vue'),
          meta: {
            title: '用户管理',
            requiresAuth: true,
            requiresAdmin: true,
          },
        },
        {
          path: 'usergroups',
          name: 'UserGroups',
          component: () => import('@/views/dashboard/usergroups/index.vue'),
          meta: {
            title: '用户组管理',
            requiresAuth: true,
            requiresAdmin: true,
          },
        },
        {
          path: 'cloudtokens',
          name: 'CloudTokens',
          component: () => import('@/views/dashboard/cloudtokens/index.vue'),
          meta: {
            title: '令牌管理',
            requiresAuth: true,
            requiresAdmin: true,
          },
        },
        {
          path: 'storages',
          name: 'Storages',
          component: () => import('@/views/dashboard/storages/index.vue'),
          meta: {
            title: '存储管理',
            requiresAuth: true,
            requiresAdmin: true,
          },
        },
      ],
    },
  ],
})

// 路由守卫
router.beforeEach((to, _, next) => {
  const authStore = useAuthStore()

  // 检查是否需要认证
  if (to.meta.requiresAuth) {
    // 需要认证的路由
    if (!authStore.isLogin) {
      // 未登录，跳转到登录页
      next('/@login')
      return
    }

    // 检查是否需要管理员权限
    if (to.meta.requiresAdmin && !authStore.isAdmin) {
      // 需要管理员权限但用户不是管理员，跳转到仪表板首页
      next('/@dashboard')
      return
    }

    next()
  } else if (to.path === '/@login' && authStore.isLogin) {
    // 已登录用户访问登录页，跳转到仪表板
    next('/@dashboard')
  } else {
    // 不需要认证的路由，直接通过
    next()
  }
})

export default router

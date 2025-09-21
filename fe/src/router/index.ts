import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

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

// 用户状态初始化标记
let userStateInitialized = false

// 路由守卫
router.beforeEach((to, _, next) => {
  const userStore = useUserStore()

  // 首次访问时初始化用户状态
  const initPromise = !userStateInitialized
    ? userStore.initUserState().then(() => {
        userStateInitialized = true
      })
    : Promise.resolve()

  initPromise
    .then(() => {
      // 检查是否需要认证
      if (to.meta.requiresAuth) {
        // 需要认证的路由
        if (!userStore.isLoggedIn) {
          // 未登录，跳转到登录页
          next('/@login')
          return
        }

        // 验证登录状态
        return userStore
          .checkLoginStatus()
          .then((isValid) => {
            if (!isValid) {
              // token 无效，跳转到登录页
              next('/@login')
              return
            }

            // 检查是否需要管理员权限
            if (to.meta.requiresAdmin && !userStore.isAdmin) {
              // 需要管理员权限但用户不是管理员，跳转到仪表板首页
              next('/@dashboard')
              return
            }

            next()
          })
          .catch((error) => {
            // 验证失败，跳转到登录页
            console.error('路由守卫验证失败:', error)
            next('/@login')
          })
      } else if (to.path === '/@login' && userStore.isLoggedIn) {
        // 已登录用户访问登录页，验证 token 是否有效
        return userStore
          .checkLoginStatus()
          .then((isValid) => {
            if (isValid) {
              // token 有效，跳转到仪表板
              next('/@dashboard')
            } else {
              // token 无效，允许访问登录页
              next()
            }
          })
          .catch((error) => {
            // 验证失败，允许访问登录页
            console.error('登录页面验证失败:', error)
            next()
          })
      } else {
        // 不需要认证的路由，直接通过
        next()
      }
    })
    .catch((error) => {
      console.error('用户状态初始化失败:', error)
      // 初始化失败，如果是需要认证的路由，跳转到登录页
      if (to.meta.requiresAuth) {
        next('/@login')
      } else {
        next()
      }
    })
})

export default router

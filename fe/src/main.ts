import { createApp } from 'vue'
import naive from 'naive-ui'
import App from './App.vue'
import router from './router/index'
import { pinia, setupStores } from './stores'

import './style.css'

const app = createApp(App)

app.use(pinia)
app.use(router)
app.use(naive)

// 初始化stores连接
setupStores()

app.mount('#app')

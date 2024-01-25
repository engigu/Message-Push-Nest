import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// import './assets/styles/global.css'
import pinia from './store';
import echarts from "./echarts";


const app = createApp(App)
app.use(router)
app.use(ElementPlus)
app.use(pinia);

app.provide('$echarts', echarts);

app.mount('#app')



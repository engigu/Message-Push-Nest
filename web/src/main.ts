import { createApp } from 'vue'
import './index.css'
import App from './App.vue'
import pinia from './store';
//@ts-ignore
import router from './router';

const app = createApp(App)
app.use(router)
app.use(pinia)
app.mount('#app')

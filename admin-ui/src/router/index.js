import Vue from 'vue'
import VueRouter from 'vue-router'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'Home',
		redirect: {name: "Dashboard"}
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
		meta: {
			requireLogin: true
		}
  },
  {
    path: '/namespaces',
    name: 'Namespaces',
    component: () => import('../views/Namespaces.vue'),
		meta: {
			requireLogin: true
		}
  },
  {
    path: '/accounts',
    name: 'Accounts',
    component: () => import('../views/Accounts.vue'),
		meta: {
			requireLogin: true
		}
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import(/* webpackChunkName: "about" */ '../views/login.vue'),
		meta: {
			requireUnlogin: true
		}
  },

]

const router = new VueRouter({
  // mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router

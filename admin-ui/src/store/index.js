import Vue from 'vue'
import Vuex from 'vuex'

// modules
import auth from './auth'
import namespaces from './namespaces'
import accounts from './accounts'
import servicesProviders from './servicesProviders'
import dns from './dns'
import settings from './settings'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
  },
  mutations: {
  },
  actions: {
  },
  modules: {
		auth,
		namespaces,
		accounts,
		servicesProviders,
		dns,
		settings,
  }
})

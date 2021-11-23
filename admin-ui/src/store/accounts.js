import api from "@/api.js"

export default {
	namespaced: true,
  state: {
		accounts: []
  },
  mutations: {
		setAccounts(state, accounts){
			state.accounts = accounts;
		}
  },
  actions: {
		fetch({commit}){
			return new Promise((resolve, reject) => {
				api.accounts.list()
				.then(response => {
					commit('setAccounts', response.pool)
					resolve(response)
				})
				.catch(error => {
					reject(error);
				})
			})
		}
  },
	getters: {
		all(state){
			return state.accounts;
		}
	}
}

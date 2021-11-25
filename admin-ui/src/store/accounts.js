import api from "@/api.js"

export default {
	namespaced: true,
  state: {
		accounts: [],
		loading: false
  },
  mutations: {
		setAccounts(state, accounts){
			state.accounts = accounts;
		},
		setLoading(state, data){
			state.loading = data;
		}
  },
  actions: {
		fetch({commit}){
			commit("setLoading", true);
			return new Promise((resolve, reject) => {
				api.accounts.list()
				.then(response => {
					commit('setAccounts', response.pool)
					resolve(response)
				})
				.catch(error => {
					reject(error);
				})
				.finally(()=>{
					commit("setLoading", false);
				})
			})
		}
  },
	getters: {
		all(state){
			return state.accounts;
		},
		isLoading(state){
			return state.loading
		}
	}
}

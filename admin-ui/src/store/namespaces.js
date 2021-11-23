import api from "@/api.js"

export default {
	namespaced: true,
  state: {
		namespaces: []
  },
  mutations: {
		setNamespaces(state, namespaces){
			state.namespaces = namespaces;
		}
  },
  actions: {
		fetch({commit}){
			return new Promise((resolve, reject) => {
				api.namespaces.list()
				.then(response => {
					commit('setNamespaces', response.pool)
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
			return state.namespaces;
		}
	}
}

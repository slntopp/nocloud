import api from "@/api.js"

export default {
	namespaced: true,
	store: {
		services: [],
		loading: false
	},
	mutations: {
		setServices(state, services){
			state.services = services;
		},
		setLoading(state, data){
			state.loading = data;
		}
	},
	actions: {
		fetch({commit}){
			commit("setLoading", true);
			return new Promise((resolve, reject) => {
				api.services.list()
				.then(response => {
					console.log(response);
					commit('setServices', response.pool)
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
			return state.services;
		},
		isLoading(state){
			return state.loading
		}
	}
}
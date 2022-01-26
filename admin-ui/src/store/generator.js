import api from "@/api.js"

export function generator(moduleName, respWrapper = 'pool'){
	return {
	namespaced: true,
  state: {
		value: [],
		loading: false
  },
  mutations: {
		setValue(state, value){
			state.value = value;
		},
		setLoading(state, data){
			state.loading = data;
		}
  },
  actions: {
		fetch({commit}){
			commit("setLoading", true);
			return new Promise((resolve, reject) => {
				api[moduleName].list()
				.then(response => {
					console.log(response);
					commit('setValue', response[respWrapper])
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
			return state.value;
		},
		isLoading(state){
			return state.loading
		}
	}
}
}
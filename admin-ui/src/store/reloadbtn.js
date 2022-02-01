export default {
	namespaced: true,
	state: {
		loading: false,
		onclick: {
			func: null,
			params: []
		},
		btnStates: {
			disabled: false,
			visible: true,
		}
	},
	mutations: {
		setCallback(state, {func, params}){
			console.log(func, params);
			state.onclick.func = func
			state.onclick.params = params
		},
		clear(state){
			state.onclick = {func: null, params: []};
		},
		setLoading(state, data){
			state.loading = data
		},
		setState(state, {stateName, value}){
			this.state.btnStates[stateName] = value
		}
	},
	actions: {
		async onclick({ state, commit }){
			if(state.onclick.func){
				commit('setLoading', true)
				await state.onclick.func.apply(null, state.onclick.params)
				commit('setLoading', false);
			}
		}
	},
	getters: {
		isLoading(state){
			return state.loading
		},
		states(state){
			return state.btnStates
		}
	}
}
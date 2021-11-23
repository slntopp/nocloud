import api from "@/api.js"
import Cookies from 'js-cookie'
import router from '@/router'

const COOKIES_NAME = 'noCloud-token';

export default {
	namespaced: true,
  state: {
		token: ''
  },
  mutations: {
		setToken(state, token){
			state.token = token;
		}
  },
  actions: {
		login({commit}, {login, password}){
			return new Promise((resolve, reject) => {
				api.auth(login, password)
				.then(response => {
					Cookies.set(COOKIES_NAME, response.token)
					commit('setToken', response.token);
					resolve(response);
				})
				.catch(error => {
					reject(error)
				})
			})
		},

		logout({commit}){
			commit('setToken', '');
			Cookies.remove(COOKIES_NAME);
			router.push({name: "Login"});
		},

		load({commit}){
			const token = Cookies.get(COOKIES_NAME);
			if(token){
				api.axios.defaults.headers.common['Authorization'] = "Bearer " + token;
				commit('setToken', token);
			}
		}
  },
	getters: {
		isLoggedIn(state){
			return state.token.length > 0;
		}
	}
}

<template>
  <v-app :style="{background: $vuetify.theme.themes.dark.background}">
    <v-navigation-drawer
      app
      permanent
      expand-on-hover
    >
			<v-list>
				<v-list-item v-if="isLoggedIn" @click="logoutHandler">
          <v-list-item-icon>
            <v-icon v-text="'mdi-account'"></v-icon>
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>logout</v-list-item-title>
          </v-list-item-content>
				</v-list-item>
			</v-list>
    </v-navigation-drawer>
    <v-app-bar
      app
			color="background"
			elevation=0
    >
      <div class="d-flex align-center">
        <v-img
          alt="Vuetify Logo"
          class="shrink mr-2"
          contain
          src="https://cdn.vuetifyjs.com/images/logos/vuetify-logo-dark.png"
          transition="scale-transition"
          width="40"
        />

        <v-img
          alt="Vuetify Name"
          class="shrink mt-1 hidden-sm-and-down"
          contain
          min-width="100"
          src="https://cdn.vuetifyjs.com/images/logos/vuetify-name-dark.png"
          width="100"
        />
      </div>

      <v-spacer></v-spacer>

    </v-app-bar>

    <v-main>
      <router-view/>
    </v-main>
  </v-app>
</template>

<script>

export default {
  name: 'App',

  data: () => ({
    //
  }),
	methods:{
		logoutHandler(){
			this.$store.dispatch('auth/logout')
		}
	},
	created(){
		this.$store.dispatch('auth/load')

		this.$router.onReady(()=>{
			const route = this.$route;
			console.log(route, this.isLoggedIn);
			if(route.matched.some(el => el.meta.requireLogin) && !this.isLoggedIn){
				this.$router.replace({name: "Login"});
			}

			if(route.matched.some(el => el.meta.requireUnlogin) && this.isLoggedIn){
				this.$router.replace({name: "Home"});
			}
		})

		this.$router.beforeEach((to, from, next)=>{
			if(to.matched.some(el => el.meta.requireLogin) && !this.isLoggedIn){
				next({name: "Login"});
			} else if(to.matched.some(el => el.meta.requireUnlogin) && this.isLoggedIn) {
				next(from);
			} else {
				next();
			}
		})
	},
	computed: {
		isLoggedIn(){
			const result = this.$store.getters['auth/isLoggedIn']
			console.log(result);
			return result
		}
	}
};
</script>


<style scoped lang="scss">
</style>
<template>
  <v-app :style="{background: $vuetify.theme.themes.dark.background}">
    <v-navigation-drawer
      app
      permanent
      color="background-light"
      :mini-variant="miniNav"
    >
    
      <router-link
        to="/"
      >
        <div class="d-flex gg-15px align-center justify-center" :class="[miniNav ? 'pa-3' : 'pa-5']">
          <v-img
            alt=""
            src="@/assets/logo.svg"
            max-height="42px"
            max-width="48px"
            contain
          ></v-img>
					
          <v-img
						v-if="!miniNav"
						transition="fade-transition"
            alt=""
            src="@/assets/logoTitle.svg"
            max-height="24px"
            max-width="122px"
            contain
          ></v-img>
        </div>
      </router-link>


      <v-list
        v-if="isLoggedIn"
        dense
      >

				<v-subheader>MAIN</v-subheader>

        <v-list-item :to="{name: 'Dashboard'}">
          <v-list-item-icon>
            <v-icon>mdi-view-dashboard-variant</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Dashboard</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item :to="{name: 'Namespaces'}">
          <v-list-item-icon>
            <v-icon>mdi-form-textbox</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Namespaces</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item :to="{name: 'Accounts'}">
          <v-list-item-icon>
            <v-icon>mdi-account-multiple</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Accounts</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item :to="{name: 'ServicesProviders'}">
          <v-list-item-icon>
            <v-icon>mdi-database-marker</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Services Providers</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item :to="{name: 'DNS manager'}">
          <v-list-item-icon>
            <v-icon>mdi-dns</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>DNS manager</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

				<v-subheader>SYSTEM</v-subheader>

        <v-list-item :to="{name: 'Settings'}">
          <v-list-item-icon>
            <v-icon>mdi-cogs</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Settings</v-list-item-title>
          </v-list-item-content>
        </v-list-item>


        
      </v-list>
    </v-navigation-drawer>


    <v-app-bar
      v-if="isLoggedIn"
      app
      color="background"
      elevation=0
    >

      <v-text-field
        hide-details
        prepend-inner-icon="mdi-magnify"
        placeholder="Search..."
        single-line
        :background-color="bgc"
        dence
        rounded
      ></v-text-field>

      <v-spacer></v-spacer>
      
      <div class="text-center">
        <v-menu
          offset-y
          transition="slide-y-transition"
        >
          <template v-slot:activator="{ on, attrs }">
            <v-btn
              class="mx-2"
              fab
              dark
              :color="bgc"
              v-bind="attrs"
              v-on="on"
            >
              <v-icon dark>
                mdi-account
              </v-icon>
            </v-btn>
          </template>
          <v-list
            dence
						min-width="250px"
          >
            <v-list-item>
							<v-list-item-content>
								<v-list-item-title class="text-h6">
									{{userdata.title}}
								</v-list-item-title>
								<v-list-item-subtitle>#{{userdata.uuid}}</v-list-item-subtitle>
							</v-list-item-content>
						</v-list-item>
						<v-divider></v-divider>
            <v-list-item @click="logoutHandler">
              <v-list-item-title>Logout</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </div>

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
    bgc: "#0c0c3c",
		miniNav: false
  }),
  methods:{
    logoutHandler(){
      this.$store.dispatch('auth/logout')
    },
		setTitle(value){
			document.title = `NoCloud | ${value}`;
		}
  },
  created(){
    this.$store.dispatch('auth/load')

    this.$router.onReady(()=>{
      const route = this.$route;
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

		this.$router.afterEach(to => {
			this.setTitle(to.name)
		})

		if(this.isLoggedIn){
			this.$store.dispatch('auth/fetchUserData')
		}
  },
  computed: {
    isLoggedIn(){
      const result = this.$store.getters['auth/isLoggedIn']
      return result
    },
		userdata(){
			return this.$store.getters['auth/userdata'];
		}
  }
};
</script>


<style scoped lang="scss">
@import '@/styles/globalStyles.scss';
</style>
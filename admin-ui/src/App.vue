<template>
  <v-app v-if="!isVNC" :style="{ background: $vuetify.theme.themes.dark.background }">
    <v-navigation-drawer
      app
      permanent
      :color="asideColor"
      :mini-variant="miniNav"
    >
      <router-link to="/">
        <!-- <div class="d-flex gg-15px align-center justify-center" :class="[miniNav ? 'pa-3' : 'pa-5']"> -->
        <div class="d-flex gg-15px align-center justify-center pa-5">
          <template v-if="!config.logoSrc">
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
          </template>
          <template v-else>
            <v-img
              v-if="!miniNav"
              transition="fade-transition"
              alt=""
              :src="config.logoSrc"
              contain
            ></v-img>
          </template>
        </div>
      </router-link>

      <v-list v-if="isLoggedIn" dense :dark="asideDark">
        <v-subheader>MAIN</v-subheader>

        <v-list-item v-bind="listItemBind" :to="{ name: 'Dashboard' }">
          <v-list-item-icon>
            <v-icon>mdi-view-dashboard-variant</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ navTitle("Dashboard") }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item v-bind="listItemBind" :to="{ name: 'Accounts' }">
          <v-list-item-icon>
            <v-icon>mdi-account-multiple</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ navTitle("Accounts") }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item v-bind="listItemBind" :to="{ name: 'Namespaces' }">
          <v-list-item-icon>
            <v-icon>mdi-form-textbox</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ navTitle("Namespaces") }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item
          v-bind="listItemBind"
          :to="{ name: 'Services' }"
          @click.ctrl="() => (easterEgg = true)"
        >
          <v-list-item-icon>
            <v-icon :color="easterEgg ? 'green darker-2' : undefined"
              >mdi-alien</v-icon
            >
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ navTitle("Services") }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item v-bind="listItemBind" :to="{ name: 'ServicesProviders' }">
          <v-list-item-icon>
            <v-icon>mdi-database-marker</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{
              navTitle("Services Providers")
            }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-subheader>BILLING</v-subheader>

        <v-list-item v-bind="listItemBind" :to="{ name: 'Plans' }">
          <v-list-item-icon>
            <v-icon>mdi-script-text</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ navTitle("Plans") }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item v-bind="listItemBind" :to="{ name: 'Transactions' }">
          <v-list-item-icon>
            <v-icon>mdi-abacus</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{
              navTitle("Transactions")
            }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-subheader>SYSTEM</v-subheader>

        <v-list-item v-bind="listItemBind" :to="{ name: 'DNS manager' }">
          <v-list-item-icon>
            <v-icon>mdi-dns</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ navTitle("DNS manager") }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>

        <v-list-item v-bind="listItemBind" :to="{ name: 'Settings' }">
          <v-list-item-icon>
            <v-icon>mdi-cogs</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>{{ navTitle("Settings") }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-app-bar v-if="isLoggedIn" app color="background" elevation="0">
      <v-row style="width: 100%" justify="center" align="center">
        <v-col>
          <v-text-field
            hide-details
            prepend-inner-icon="mdi-magnify"
            placeholder="Search..."
            single-line
            background-color="background-light"
            dence
            v-model="searchParam"
            rounded
          ></v-text-field>
        </v-col>
        <v-col class="d-flex justify-center">
          <v-btn
            v-if="btnStates.visible"
            :disabled="btnStates.disabled"
            color="background-light"
            fab
            small
            :loading="btnLoading"
            @click="() => this.$store.dispatch('reloadBtn/onclick')"
          >
            <v-icon>mdi-reload</v-icon>
          </v-btn>
        </v-col>
        <v-col class="d-flex justify-end align-center">
          <balance title="Balance: " />
          <v-menu offset-y transition="slide-y-transition">
            <template v-slot:activator="{ on, attrs }">
              <v-btn
                class="mx-2"
                fab
                color="background-light"
                v-bind="attrs"
                v-on="on"
              >
                <v-icon dark> mdi-account </v-icon>
              </v-btn>
            </template>
            <v-list dence min-width="250px">
              <v-list-item>
                <v-list-item-content>
                  <v-list-item-title class="text-h6">
                    {{ userdata.title }}
                  </v-list-item-title>
                  <v-list-item-subtitle
                    >#{{ userdata.uuid }}</v-list-item-subtitle
                  >
                </v-list-item-content>
              </v-list-item>
              <v-divider></v-divider>
              <v-list-item @click="logoutHandler">
                <v-list-item-title>Logout</v-list-item-title>
              </v-list-item>
            </v-list>
          </v-menu>
        </v-col>
      </v-row>
      <v-spacer></v-spacer>
    </v-app-bar>

    <v-main>
      <router-view />
    </v-main>
  </v-app>
  <router-view v-else/>
</template>

<script>
import config from "@/config";
import balance from "@/components/balance.vue";

export default {
  components: { balance },
  name: "App",

  data: () => ({
    miniNav: true,
    easterEgg: false,
    config,
    navTitles: config.navTitles ?? {},
  }),
  methods: {
    logoutHandler() {
      this.$store.dispatch("auth/logout");
    },
    setTitle(value) {
      document.title = `${value} | NoCloud`;
    },
    onResize() {
      this.miniNav = window.innerWidth <= 768;
    },
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
  },
  computed: {
    isLoggedIn() {
      const result = this.$store.getters["auth/isLoggedIn"];
      return result;
    },
    userdata() {
      return this.$store.getters["auth/userdata"];
    },
    btnLoading() {
      return this.$store.getters["reloadBtn/isLoading"];
    },
    btnStates() {
      return this.$store.getters["reloadBtn/states"];
    },
    asideColor() {
      return config?.dye?.aside?.background ?? "background-light";
    },
    asideLinksColor() {
      return config?.dye?.aside?.links ?? undefined;
    },
    asideDark() {
      return config?.dye?.aside?.whiteText ?? undefined;
    },
    listItemBind() {
      if (this.asideLinksColor)
        return {
          color: this.asideLinksColor,
          dark: this.asideDark,
        };
      else return {};
    },
    searchParam: {
      get() {
        return this.$store.getters["appSearch/param"];
      },
      set(newValue) {
        this.$store.commit("appSearch/setSearchParam", newValue);
      },
    },
    isVNC(){
      return this.$route.path.includes('vnc')
    }
  },
  created() {
    this.$store.dispatch("auth/load");

    this.$router.onReady(() => {
      const route = this.$route;
      if (
        route.matched.some((el) => el.meta.requireLogin) &&
        !this.isLoggedIn
      ) {
        this.$router.replace({ name: "Login" });
      }

      if (
        route.matched.some((el) => el.meta.requireUnlogin) &&
        this.isLoggedIn
      ) {
        this.$router.replace({ name: "Home" });
      }
    });

    this.$router.beforeEach((to, from, next) => {
      this.$store.commit("reloadBtn/setLoading", false);

      if (to.matched.some((el) => el.meta.requireLogin) && !this.isLoggedIn) {
        next({ name: "Login" });
      } else if (
        to.matched.some((el) => el.meta.requireUnlogin) &&
        this.isLoggedIn
      ) {
        next(from);
      } else {
        next();
      }
    });

    this.$router.afterEach((to) => {
      this.setTitle(to.name);
    });

    if (this.isLoggedIn) {
      this.$store.dispatch("auth/fetchUserData");
    }
  },
  mounted() {
    this.onResize();
    window.addEventListener("resize", this.onResize, { passive: true });
  },
};
</script>

<style scoped lang="scss">
@import "@/styles/globalStyles.scss";
</style>

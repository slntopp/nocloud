<template>
  <div>
    <v-app
      :style="{
        background: $vuetify.theme.themes[theme]?.colors.background,
        width: isFullscreen && !isNotPlugin ? '0px' : undefined,
        height: isFullscreen && !isNotPlugin ? '0px' : undefined,
      }"
    >
      <div
        v-if="isFullscreenAvailable && isFullscreen"
        style="position: fixed; right: 5vw; z-index: 100; margin-top: 1vh"
      >
        <v-btn @click="toggleFullscreen" color="background-light" icon variant="elevated" size="small">
          <v-icon>mdi-fullscreen-exit</v-icon>
        </v-btn>
      </div>

      <template v-if="!isFullscreen">
        <v-navigation-drawer
          permanent
          :color="asideColor"
          :rail="isMenuMinimize"
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
                  
                ></v-img>

                <v-img
                  v-if="!isMenuMinimize"
                  transition="fade-transition"
                  alt=""
                  src="@/assets/logoTitle.svg"
                  max-height="24px"
                  max-width="122px"
                  
                ></v-img>
              </template>
              <template v-else>
                <v-img
                  v-if="!isMenuMinimize"
                  transition="fade-transition"
                  alt=""
                  :src="config.logoSrc"
                  
                ></v-img>
              </template>
            </div>
          </router-link>

          <v-list
            style="height: 100%"
            v-if="isLoggedIn"
            density="compact"
            :theme="asideDark ? 'dark' : undefined"
          >
            <div
              :class="{
                'd-flex': true,
                'align-center': true,
                'justify-space-between': !isMenuMinimize,
                'flex-column-reverse': isMenuMinimize,
              }"
            >
              <v-list-subheader>MAIN</v-list-subheader>
              <v-btn @click="isMenuMinimize = !isMenuMinimize" icon variant="text">
                <v-icon v-if="isMenuMinimize">mdi-arrow-right</v-icon>
                <v-icon v-else>mdi-arrow-left</v-icon>
              </v-btn>
            </div>

            <div style="height: 100%" id="drawer-menu-hover">
              <v-list-item v-bind="listItemBind" :to="{ name: 'Dashboard' }">
                <template #prepend>
                  <v-icon>mdi-view-dashboard-variant</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Dashboard")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Accounts' }">
                <template #prepend>
                  <v-icon>mdi-account-multiple</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Accounts")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item
                v-bind="listItemBind"
                :to="{ name: 'AccountGroups' }"
              >
                <template #prepend>
                  <v-icon>mdi-account-group-outline</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Account Groups")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Instances' }">
                <template #prepend>
                  <v-icon>mdi-server</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Instances")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Showcases' }">
                <template #prepend>
                  <v-icon>mdi-store-search</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Showcases")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-subheader>BILLING</v-list-subheader>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Plans' }">
                <template #prepend>
                  <v-icon>mdi-script-text</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Price Models")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Addons' }">
                <template #prepend>
                  <v-icon>mdi-tab-plus</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Addons")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Promocodes' }">
                <template #prepend>
                  <v-icon>mdi-sale</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Promocodes")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Transactions' }">
                <template #prepend>
                  <v-icon>mdi-abacus</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Transactions")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Invoices' }">
                <template #prepend>
                  <v-icon>mdi-invoice-list</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Invoices")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Currencies' }">
                <template #prepend>
                  <v-icon>mdi-cash</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Currencies")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Reports' }">
                <template #prepend>
                  <v-icon>mdi-chart-gantt</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Reports")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-subheader v-if="plugins?.length > 0">PLUGINS</v-list-subheader>

              <v-list-item
                v-bind="listItemBind"
                v-for="plugin of plugins"
                :key="plugin.url"
                :to="{
                  name: 'Plugin',
                  params: { title: plugin.title },
                  query: {
                    url: plugin.url,
                    fullscreen: viewport > 768 ? false : true,
                  },
                }"
              >
                <template #prepend>
                  <v-icon>mdi-{{ plugin.icon }}</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle(plugin.title)
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-subheader>SYSTEM</v-list-subheader>

              <v-list-item
                v-bind="listItemBind"
                @click.ctrl.prevent="openChatsInNewWindow"
                @click.exact="chatClick"
              >
                <template #prepend>
                  <v-icon
                    :color="
                      unreadChatsCount && isMenuMinimize ? 'red' : undefined
                    "
                    >mdi-chat</v-icon
                  >
                </template>

                
                  <v-list-item-title
                    >{{ navTitle("Chats") }}
                    <v-chip variant="flat"
                      color="red"
                      class="pa-2"
                      v-if="unreadChatsCount"
                      size="x-small"
                    >
                      {{ unreadChatsCount }}
                    </v-chip>
                  </v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'History' }">
                <template #prepend>
                  <v-icon>mdi-history</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Event log")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Statistics' }">
                <template #prepend>
                  <v-icon>mdi-chart-box-outline</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Statistics")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'DNS manager' }">
                <template #prepend>
                  <v-icon>mdi-dns</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("DNS manager")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Namespaces' }">
                <template #prepend>
                  <v-icon>mdi-form-textbox</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Groups (NameSpaces)")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item
                v-bind="listItemBind"
                :to="{ name: 'ServicesProviders' }"
              >
                <template #prepend>
                  <v-icon>mdi-database-marker</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Services Providers")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item
                v-bind="listItemBind"
                :to="{ name: 'Services' }"
                @click.ctrl="() => (easterEgg = true)"
              >
                <template #prepend>
                  <v-icon :color="easterEgg ? 'green darker-2' : undefined"
                    >mdi-alien</v-icon
                  >
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Services")
                  }}</v-list-item-title>
                
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Settings' }">
                <template #prepend>
                  <v-icon>mdi-cogs</v-icon>
                </template>

                
                  <v-list-item-title>{{
                    navTitle("Settings")
                  }}</v-list-item-title>
                
              </v-list-item>
            </div>
          </v-list>
        </v-navigation-drawer>

        <v-app-bar color="background" elevation="0">
          <v-row
            style="width: 100%"
            class="flex-nowrap justify-center align-center"
          >
            <template v-if="isLoggedIn">
              <v-col :style="viewport < 600 ? 'padding: 6px' : null">
                <app-search />
              </v-col>
              <v-col
                class="d-flex justify-start"
                :style="viewport < 600 ? 'padding: 6px' : null"
              >
                <v-btn
                  v-if="btnStates.visible"
                  :disabled="btnStates.disabled"
                  color="background-light"
                  icon
                  variant="elevated"
                  size="small"
                  :loading="btnLoading"
                  @click="() => this.$store.dispatch('reloadBtn/onclick')"
                >
                  <v-icon>mdi-reload</v-icon>
                </v-btn>
                <v-btn
                  class="ml-2"
                  color="background-light"
                  v-if="isFullscreenAvailable && !isFullscreen"
                  @click="toggleFullscreen"
                  icon
                  variant="elevated"
                  size="small"
                >
                  <v-icon>mdi-fullscreen</v-icon>
                </v-btn>
              </v-col>
            </template>
            <v-col
              class="d-flex justify-end align-center"
              :style="viewport < 600 ? 'padding: 6px' : null"
            >
              <languages v-if="false" />
              <themes />
              <notifications />
              <user-menu v-if="isLoggedIn" />
            </v-col>
          </v-row>
          <v-spacer></v-spacer>
        </v-app-bar>

        <v-main>
          <router-view />
        </v-main>
        <app-snackbar />
      </template>

      <v-main v-else-if="isFullscreen && isNotPlugin">
        <router-view />
      </v-main>
    </v-app>

    <router-view v-if="isFullscreen && !isNotPlugin" />
    <instances-table-modal
      v-if="overlay.uuid"
      type="menu"
      :uuid="overlay.uuid"
      :visible="overlay.isVisible"
      @close="overlay.isVisible = false"
      @hover="hoverOverlay"
    >
      <template #activator="{ props }">
        <v-btn
          v-bind="props"
          variant="outlined"
          color="success"
          :style="{
            position: 'absolute',
            top: `${overlay.y + (isFullscreen ? 0 : 80)}px`,
            right: `30px`,
            zIndex: 100,
            visibility: 'hidden',
          }"
        >
          {{ overlay.buttonTitle }}
        </v-btn>
      </template>
    </instances-table-modal>
  </div>
</template>

<script>
import { defineAsyncComponent, reactive, ref } from "vue";
import { mapGetters } from "vuex";
import useLoginClient from "@/hooks/useLoginInClient.js";
import api from "@/api.js";
import config from "@/config.js";
import languages from "@/components/header/languages.vue";
import appSearch from "@/components/search/search.vue";
import AppSnackbar from "@/components/snackbar.vue";
// only shown for the chats overlay; keeps the table and Sortable out of the entry chunk
const instancesTableModal = defineAsyncComponent(() =>
  import("@/components/instances_table_modal.vue"),
);
import Themes from "@/components/header/themes.vue";
import userMenu from "@/components/header/userMenu.vue";
import notifications from "@/components/header/notifications.vue";

// base64 of the UTF-8 bytes of a string, like Buffer.from(str).toString("base64")
const utf8Base64 = (str) => btoa(unescape(encodeURIComponent(str)));

export default {
  name: "App",
  components: {
    Themes,
    AppSnackbar,
    appSearch,
    languages,
    instancesTableModal,
    userMenu,
    notifications,
  },
  setup() {
    const { loginHandler } = useLoginClient();

    return {
      viewport: ref(window.innerWidth),
      isMenuMinimize: ref(true),
      isMouseOnMenu: ref(false),
      easterEgg: ref(false),
      config,
      navTitles: ref(config.navTitles ?? {}),
      overlay: reactive({
        timeoutId: null,
        isVisible: false,
        buttonTitle: "",
        uuid: "",
        x: 0,
        y: 0,
      }),
      loginHandler,
    };
  },
  methods: {
    logoutHandler() {
      this.$store.dispatch("auth/logout");
    },
    setTitle(value) {
      document.title = `${value} | NoCloud`;
    },
    onResize(e) {
      this.viewport = window.innerWidth;
      if (!(e instanceof UIEvent)) {
        this.isMenuMinimize = window.innerWidth <= 768;
      }
    },
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    configureHoverOnMenu() {
      document
        .getElementById("drawer-menu-hover")
        ?.addEventListener("mouseenter", () => {
          if (this.isMenuMinimize) {
            this.isMouseOnMenu = true;
            setTimeout(() => {
              if (this.isMouseOnMenu) {
                this.isMenuMinimize = false;
              }
            }, 5000);
          }
        });
      document
        .getElementById("drawer-menu-hover")
        ?.addEventListener("mouseleave", () => {
          if (this.isMouseOnMenu) {
            this.isMouseOnMenu = false;
            setTimeout(() => {
              this.isMenuMinimize = true;
            }, 1000);
          }
        });
    },
    hoverOverlay() {
      clearTimeout(this.overlay.timeoutId);
    },
    hiddenOverlay() {
      this.overlay.timeoutId = setTimeout(() => {
        this.overlay.isVisible = false;
      }, 100);
    },
    chatClick() {
      this.$store.commit("app/setChatClicks", 1);
      this.$router.push({ name: "Chats" });
    },
    openChatsInNewWindow() {
      const { title } = this.$store.getters["auth/userdata"];
      const { token } = this.$store.state.auth;
      const fullParams = JSON.stringify({
        title,
        token,
        api: location.host,
        theme: this.theme,
        fullscreen: this.$route.query["fullscreen"] === "true",
      });

      return window.open(
        `/cc.ui/?a=${utf8Base64(unescape(encodeURIComponent(fullParams)))}`,
        "_blanc"
      );
    },
    toggleFullscreen() {
      this.$router.push({
        path: this.$route.path,
        query: {
          ...this.$route.query,
          fullscreen: !this.isFullscreen,
        },
      });
    },
  },
  computed: {
    ...mapGetters("app", ["theme"]),
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
          theme: this.asideDark ? "dark" : undefined,
        };
      else return {};
    },
    isFullscreen() {
      return (
        this.$route.path.includes("vnc") ||
        this.$route.query["fullscreen"] === "true"
      );
    },
    isFullscreenAvailable() {
      return this.$route.query["fullscreen"] !== undefined;
    },
    isNotPlugin() {
      return this.$route.name !== "Plugin";
    },
    plugins() {
      return this.$store.getters["plugins/all"];
    },
    unreadChatsCount() {
      return this.$store.getters["chats/unreadChatsCount"];
    },
  },
  created() {
    window.addEventListener("message", ({ data, origin, source }) => {
      if (origin.includes("localhost") || !data) return;
      if (data === "ready") return;
      if (data.type === "send-user") {
        const setting = "plugin-chats-overlay";

        api.settings.get([setting]).then((res) => {
          const { title } = JSON.parse(res[setting]);

          setTimeout(() => {
            source.postMessage({ type: "button-title", value: title }, "*");
          }, 300);
          this.overlay.uuid = data.value.uuid;
        });
        return;
      }
      if (data.type === "click-on-button") {
        this.overlay.x = data.value.x;
        this.overlay.y = data.value.y;
        this.overlay.isVisible = !this.overlay.isVisible;
      }
      if (data.type === "open-user") {
        window.open(`/admin/accounts/${data.value.uuid}`, "_blank");
        return;
      }
      if (data.type === "open-chats-statistics") {
        this.$router.push({ name: "Statistics", query: { tab: "chats" } });
        return;
      }
      if (data.type === "open-chat") {
        this.loginHandler({
          accountUuid: this.userdata.uuid,
          chatId: data.value.uuid,
        });
        return;
      }
      if (data.type === "get-theme") {
        source.postMessage({ theme: this.theme }, "*");
        return;
      }
    });

    this.$store.dispatch("auth/load");

    this.$router.isReady().then(() => {
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
      this.overlay.uuid = "";
      this.overlay.buttonTitle = "";
      this.setTitle(to.name);
    });

    if (this.isLoggedIn) {
      this.$store.dispatch("auth/fetchUserData");
      this.$store.dispatch("chats/fetch");
    }
  },
  mounted() {
    this.isMenuMinimize =
      +localStorage.getItem("nocloud-menu-minimize") === 0 ? false : true;

    window.addEventListener("resize", this.onResize, { passive: true });

    this.configureHoverOnMenu();
  },
  watch: {
    isLoggedIn(newVal) {
      if (newVal) {
        this.$store.dispatch("plugins/fetch");
        this.$store.dispatch("currencies/fetch");
        this.$store.dispatch("settings/fetch");
      }
    },
    isMenuMinimize(newValue) {
      localStorage.setItem("nocloud-menu-minimize", +newValue);
    },
  },
};
</script>

<style lang="scss">
/* Vuetify 2 dense lists used smaller titles in the navigation drawer */
.v-navigation-drawer .v-list--density-compact .v-list-item-title {
  font-size: 0.8125rem;
  font-weight: 500;
}
</style>

<style scoped lang="scss">
@import "@/styles/globalStyles.scss";
</style>

<style lang="scss">
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
  border: none;
  border-radius: unset;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgb(var(--v-theme-background));
}

::-webkit-scrollbar-thumb:hover {
  background: rgb(var(--v-theme-primary));
}
</style>

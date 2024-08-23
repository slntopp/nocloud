<template>
  <div>
    <v-app
      :style="{
        background: $vuetify.theme.themes[theme].background,
        width: (isFullscreen && !isNotPlugin) ? '0px' : undefined,
        height: (isFullscreen && !isNotPlugin) ? '0px' : undefined,
      }"
    >
      <div
        v-if="isFullscreenAvailable && isFullscreen"
        style="position: fixed; right: 5vw; z-index: 100; margin-top: 1vh"
      >
        <v-btn @click="toggleFullscreen" color="background-light" fab small>
          <v-icon>mdi-fullscreen-exit</v-icon>
        </v-btn>
      </div>

      <template v-if="!isFullscreen">
        <v-navigation-drawer
          app
          permanent
          :color="asideColor"
          :mini-variant="isMenuMinimize"
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
                  v-if="!isMenuMinimize"
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
                  v-if="!isMenuMinimize"
                  transition="fade-transition"
                  alt=""
                  :src="config.logoSrc"
                  contain
                ></v-img>
              </template>
            </div>
          </router-link>

          <v-list
            style="height: 100%"
            v-if="isLoggedIn"
            dense
            :dark="asideDark"
          >
            <div
              :class="{
                'd-flex': true,
                'align-center': true,
                'justify-space-between': !isMenuMinimize,
                'flex-column-reverse': isMenuMinimize,
              }"
            >
              <v-subheader>MAIN</v-subheader>
              <v-btn @click="isMenuMinimize = !isMenuMinimize" icon>
                <v-icon v-if="isMenuMinimize">mdi-arrow-right</v-icon>
                <v-icon v-else>mdi-arrow-left</v-icon>
              </v-btn>
            </div>

            <div style="height: 100%" id="drawer-menu-hover">
              <v-list-item v-bind="listItemBind" :to="{ name: 'Dashboard' }">
                <v-list-item-icon>
                  <v-icon>mdi-view-dashboard-variant</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Dashboard")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Accounts' }">
                <v-list-item-icon>
                  <v-icon>mdi-account-multiple</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Accounts")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Instances' }">
                <v-list-item-icon>
                  <v-icon>mdi-server</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Instances")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Showcases' }">
                <v-list-item-icon>
                  <v-icon>mdi-store-search</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Showcases")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-subheader>BILLING</v-subheader>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Plans' }">
                <v-list-item-icon>
                  <v-icon>mdi-script-text</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Price Models")
                  }}</v-list-item-title>
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

              <v-list-item v-bind="listItemBind" :to="{ name: 'Currencies' }">
                <v-list-item-icon>
                  <v-icon>mdi-cash</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Currencies")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Reports' }">
                <v-list-item-icon>
                  <v-icon>mdi-chart-gantt</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Reports")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-subheader v-if="plugins?.length > 0">PLUGINS</v-subheader>

              <v-list-item
                v-bind="listItemBind"
                v-for="plugin of plugins"
                :key="plugin.url"
                :to="{
                  name: 'Plugin',
                  params: plugin,
                  query: { url: plugin.url, fullscreen: (viewport > 768) ? false : true },
                }"
              >
                <v-list-item-icon>
                  <v-icon>mdi-{{ plugin.icon }}</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle(plugin.title)
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-subheader>SYSTEM</v-subheader>

              <v-list-item
                v-bind="listItemBind"
                :to="{ name: 'Chats' }"
                @click="chatClick"
              >
                <v-list-item-icon>
                  <v-icon
                    :color="
                      unreadChatsCount && isMenuMinimize ? 'red' : undefined
                    "
                    >mdi-chat</v-icon
                  >
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title
                    >{{ navTitle("Chats") }}
                    <v-chip
                      color="red"
                      class="pa-2"
                      v-if="unreadChatsCount"
                      x-small
                    >
                      {{ unreadChatsCount }}
                    </v-chip>
                  </v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'History' }">
                <v-list-item-icon>
                  <v-icon>mdi-history</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Event log")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'DNS manager' }">
                <v-list-item-icon>
                  <v-icon>mdi-dns</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("DNS manager")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Namespaces' }">
                <v-list-item-icon>
                  <v-icon>mdi-form-textbox</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Groups (NameSpaces)")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item
                v-bind="listItemBind"
                :to="{ name: 'ServicesProviders' }"
              >
                <v-list-item-icon>
                  <v-icon>mdi-database-marker</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Services Providers")
                  }}</v-list-item-title>
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
                  <v-list-item-title>{{
                    navTitle("Services")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>

              <v-list-item v-bind="listItemBind" :to="{ name: 'Settings' }">
                <v-list-item-icon>
                  <v-icon>mdi-cogs</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>{{
                    navTitle("Settings")
                  }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </div>
          </v-list>
        </v-navigation-drawer>

        <v-app-bar app color="background" elevation="0">
          <v-row style="width: 100%" justify="center" align="center" class="flex-nowrap">
            <template v-if="isLoggedIn">
              <v-col :style="(viewport < 600) ? 'padding: 6px' : null">
                <app-search />
              </v-col>
              <v-col class="d-flex justify-start" :style="(viewport < 600) ? 'padding: 6px' : null">
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
                <v-btn
                  class="ml-2"
                  color="background-light"
                  v-if="isFullscreenAvailable && !isFullscreen"
                  @click="toggleFullscreen"
                  fab
                  small
                >
                  <v-icon>mdi-fullscreen</v-icon>
                </v-btn>
              </v-col>
            </template>
            <v-col class="d-flex justify-end align-center" :style="(viewport < 600) ? 'padding: 6px' : null">
              <languages v-if="false" />
              <themes v-if="viewport >= 600" />
              <v-menu
                v-if="isLoggedIn"
                offset-y
                transition="slide-y-transition"
              >
                <template v-slot:activator="{ on, attrs }">
                  <v-btn
                    class="mx-2"
                    fab
                    color="background-light"
                    v-bind="attrs"
                    v-on="on"
                    :small="viewport < 600"
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
                  <v-list-item>
                    <balance loged-in-user title="Balance: " style="margin-right: auto" />
                    <themes v-if="viewport < 600" />
                  </v-list-item>
                  <v-divider></v-divider>
                  <v-list-item @click="logoutHandler">
                    <v-list-item-title>Logout</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </v-col>
          </v-row>
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
      <template #activator>
        <v-btn
          outlined
          color="success"
          :style="{
            position: 'absolute',
            top: `${overlay.y + ((isFullscreen) ? 0 : 80)}px`,
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
import { reactive, ref } from "vue";
import { mapGetters } from "vuex";
import useLoginClient from "@/hooks/useLoginInClient.js";
import api from "@/api.js";
import config from "@/config.js";
import balance from "@/components/balance.vue";
import languages from "@/components/languages.vue";
import appSearch from "@/components/search/search.vue";
import AppSnackbar from "@/components/snackbar.vue";
import instancesTableModal from "@/components/instances_table_modal.vue";
import Themes from "@/components/themes.vue";

export default {
  name: "App",
  components: {
    Themes,
    AppSnackbar,
    balance,
    appSearch,
    languages,
    instancesTableModal,
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
      this.viewport = window.innerWidth
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
          dark: this.asideDark,
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
      return this.$route.name !== 'Plugin'
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
  background: var(--v-background-base);
}

::-webkit-scrollbar-thumb:hover {
  background: var(--v-primary-base);
}
</style>

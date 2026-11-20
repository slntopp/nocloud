import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import vuetify from "./plugins/vuetify";
import i18n from "./i18n";
import config from "./config";

// A tab left open across a deploy still runs the old build: its lazy chunks
// (assets/*-HASH.js) are gone from the server, so opening another page fails
// with "Failed to fetch dynamically imported module" until a reload. Reload
// once (to the page being opened) to get the new build; if that happened
// just before, let the error through instead of looping.
const RELOAD_KEY = "nocloud-chunk-reload";
let pendingRoute = null;
router.beforeEach((to) => {
  pendingRoute = to.fullPath;
});
router.afterEach(() => {
  pendingRoute = null;
});

window.addEventListener("vite:preloadError", (event) => {
  try {
    const lastReload = Number(sessionStorage.getItem(RELOAD_KEY)) || 0;
    if (Date.now() - lastReload < 10_000) return;
    sessionStorage.setItem(RELOAD_KEY, String(Date.now()));
  } catch {
    return; // no sessionStorage: can't guard against a reload loop
  }

  event.preventDefault();
  if (pendingRoute) {
    window.location.assign(router.resolve(pendingRoute).href);
  } else {
    window.location.reload();
  }
});

const app = createApp(App);

app.config.globalProperties.$config = config;

app.use(i18n).use(router).use(store).use(vuetify).mount("#app");

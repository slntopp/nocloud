import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import vuetify from "./plugins/vuetify";
import i18n from "./i18n";
import config from "./config";

const app = createApp(App);

app.config.globalProperties.$config = config;

app.use(i18n).use(router).use(store).use(vuetify).mount("#app");

import Vue from 'vue';
import Vuetify from 'vuetify/lib/framework';

Vue.use(Vuetify);

export default new Vuetify({
  theme: {
    themes: {
      dark: {
        background: "#000033",
        "background-dark": "#000020",
        accent: "#FF00FF",
				primary: "#FF00FF"
      }
    },
    dark: true,
  },
})
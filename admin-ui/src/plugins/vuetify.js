import Vue from 'vue';
import config from '@/config';
import Vuetify from 'vuetify/lib/framework';

Vue.use(Vuetify);

export default new Vuetify({
  theme: {
    themes: {
      dark: config.colors,
      light: config.colors,
    },
    options: {
      customProperties: true,
      variations: false
    },
    dark: config.dark ?? true, // тут поставить лайт
  },
})
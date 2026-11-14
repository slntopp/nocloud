import "vuetify/styles";
import { createVuetify } from "vuetify";
import { aliases, mdi } from "vuetify/iconsets/mdi";
import config from "@/config";

// Vuetify 2 inputs were underlined by default; keep that look
const underlined = { variant: "underlined" };

export default createVuetify({
  defaults: {
    VTextField: underlined,
    VTextarea: underlined,
    VSelect: underlined,
    VAutocomplete: underlined,
    VCombobox: underlined,
    VFileInput: underlined,
    // Vuetify 3 makes buttons in card actions flat text buttons; Vuetify 2 did not
    VCardActions: { VBtn: { variant: "elevated" } },
  },
  icons: {
    defaultSet: "mdi",
    aliases,
    sets: { mdi },
  },
  theme: {
    defaultTheme: config.dark ?? true ? "dark" : "light",
    themes: {
      dark: { dark: true, colors: config.colors?.dark ?? {} },
      light: { dark: false, colors: config.colors?.light ?? {} },
    },
    variations: false,
  },
});

import "vuetify/styles";
import "@/styles/vuetify/legacy.css";
import { createVuetify } from "vuetify";
import { aliases, mdi } from "vuetify/iconsets/mdi";
import config from "@/config";

// Vuetify 2 inputs were underlined by default; keep that look
const underlined = { variant: "underlined" };

export default createVuetify({
  // Vuetify 4 lowered the breakpoints; keep the Vuetify 3 ones (also set in Sass)
  display: {
    thresholds: { md: 960, lg: 1280, xl: 1920, xxl: 2560 },
  },
  defaults: {
    VTextField: underlined,
    VTextarea: underlined,
    VSelect: underlined,
    VAutocomplete: underlined,
    VCombobox: underlined,
    VFileInput: underlined,
    // Vuetify 3 makes buttons in card actions flat text buttons; Vuetify 2 did not
    VCardActions: { VBtn: { variant: "elevated" } },
    // Vuetify 2 colored selection controls and active tabs with primary
    VSwitch: { color: "primary" },
    VCheckbox: { color: "primary" },
    VCheckboxBtn: { color: "primary" },
    VRadio: { color: "primary" },
    VRadioGroup: { color: "primary" },
    VTabs: { color: "primary" },
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

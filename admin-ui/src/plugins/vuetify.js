import "vuetify/styles";
import "@/styles/vuetify/legacy.css";
import { createVuetify } from "vuetify";
import { aliases, mdi } from "vuetify/iconsets/mdi";
import config from "@/config";

// Vuetify 2 inputs were underlined by default; keep that look
const underlined = { variant: "underlined" };

// Vuetify 2 read item labels from `text`, Vuetify 3+ from `title`. Many items
// in the app are still { text, value }, so accept both.
// An object without a label (e.g. an empty {} model) renders as nothing,
// as in Vuetify 2, instead of "[object Object]".
const itemTitle = (item, fallback) =>
  item?.title ?? item?.text ?? (item && typeof item === "object" ? "" : fallback);
// Vuetify 2 honoured `disabled`, `header` and `divider` on item objects;
// Vuetify 3+ ignores them unless they come through item-props.
const itemProps = (item) => {
  if (!item || typeof item !== "object") return undefined;
  if (item.header) return { type: "subheader", title: item.header };
  if (item.divider) return { type: "divider" };
  if (item.disabled) return { disabled: true };
  return undefined;
};
const selectDefaults = { ...underlined, itemTitle, itemProps };

export default createVuetify({
  // Vuetify 4 lowered the breakpoints; keep the Vuetify 3 ones (also set in Sass)
  display: {
    thresholds: { md: 960, lg: 1280, xl: 1920, xxl: 2560 },
  },
  defaults: {
    VTextField: underlined,
    VTextarea: underlined,
    VSelect: selectDefaults,
    VAutocomplete: selectDefaults,
    VCombobox: selectDefaults,
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
    // Vuetify 2 skeletons had no background of their own
    VSkeletonLoader: { color: "transparent" },
  },
  icons: {
    defaultSet: "mdi",
    aliases,
    sets: { mdi },
  },
  theme: {
    defaultTheme: config.dark ?? true ? "dark" : "light",
    themes: {
      // "neutral": the default chip color of Vuetify 2, used for neutral states
      dark: {
        dark: true,
        // surface: Vuetify 2 dark card color (Vuetify 4 uses #212121)
        colors: {
          neutral: "#555555",
          surface: "#1E1E1E",
          ...(config.colors?.dark ?? {}),
        },
      },
      light: {
        dark: false,
        colors: { neutral: "#e0e0e0", ...(config.colors?.light ?? {}) },
      },
    },
    variations: false,
  },
});

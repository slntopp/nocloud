import js from "@eslint/js";
import pluginVue from "eslint-plugin-vue";
import globals from "globals";

export default [
  js.configs.recommended,
  ...pluginVue.configs["flat/essential"],
  {
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      globals: { ...globals.browser },
    },
    rules: {
      "vue/valid-v-slot": "off",
      "vue/multi-word-component-names": "off",
      "no-unsafe-optional-chaining": "off",
    },
  },
];

import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vuetify, { transformAssetUrls } from "vite-plugin-vuetify";

export default defineConfig({
  // the app is served under /admin (see apiserver_web and traefik)
  base: "/admin/",
  plugins: [
    vue({
      template: {
        transformAssetUrls: {
          // resolve src="@/assets/..." on Vuetify components such as v-img too
          tags: transformAssetUrls,
          // absolute URLs (/admin/img/...) point at files in public/, leave them as is
          includeAbsolute: false,
        },
      },
    }),
    // auto-imports only the Vuetify components and directives that are used;
    // styles are compiled from Sass with our settings (MD2 look, v3 breakpoints)
    vuetify({
      autoImport: true,
      styles: { configFile: "src/styles/vuetify/settings.scss" },
    }),
  ],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
    // the codebase imports .vue files without the extension in many places
    extensions: [".mjs", ".js", ".json", ".vue"],
  },
  css: {
    preprocessorOptions: {
      scss: {
        // the styles still use @import; migrating them to @use is a separate step
        silenceDeprecations: ["import", "global-builtin", "color-functions"],
      },
      sass: {
        silenceDeprecations: ["import", "global-builtin", "color-functions"],
      },
    },
  },
  server: {
    port: 8080,
  },
  build: {
    target: "es2020",
    chunkSizeWarningLimit: 1500,
  },
});

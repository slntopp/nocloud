import { readdirSync } from "node:fs";
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vuetify, { transformAssetUrls } from "vite-plugin-vuetify";

// Pre-bundle every Vuetify component for the dev server. Otherwise Vite finds
// them one by one while you browse and reloads the page each time, and split
// pre-bundling breaks overlay z-index stacking (see Vuetify 4 upgrade guide).
const vuetifyComponents = readdirSync(
  fileURLToPath(new URL("./node_modules/vuetify/lib/components", import.meta.url)),
  { withFileTypes: true },
)
  .filter((entry) => entry.isDirectory() && /^V[A-Z]/.test(entry.name))
  .map((entry) => `vuetify/components/${entry.name}`);

// Vuetify 4 utility classes (pa-3, d-flex, text-*) are no longer !important,
// so any app rule for the same property would now beat them. Vuetify 2/3
// utilities always won. Put the app's own CSS into the "app" cascade layer,
// ordered between Vuetify components and utilities, and third-party CSS into
// the low "external" layer (see public/layers.css).
const appCssLayer = {
  postcssPlugin: "app-css-layer",
  Once(root, { AtRule }) {
    const file = root.source?.input?.file || "";
    let name = null;
    if (file.includes("/node_modules/")) {
      // third-party CSS (jodit, editors, charts): low "external" layer, so
      // global rules in it cannot restyle Vuetify components or the app
      if (!file.includes("/node_modules/vuetify/")) name = "external";
    } else if (file.includes("/src/") && !file.includes("/src/styles/vuetify/")) {
      name = "app";
    }
    if (!name) return;
    const layer = new AtRule({ name: "layer", params: name });
    root.each((node) => {
      if (node.type === "atrule" && ["import", "charset", "layer"].includes(node.name) && !node.nodes) return;
      layer.append(node.clone());
      node.remove();
    });
    if (layer.nodes?.length) root.append(layer);
  },
};

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
    postcss: { plugins: [appCssLayer] },
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
  optimizeDeps: {
    include: ["vuetify", "vuetify/directives", ...vuetifyComponents],
  },
  build: {
    target: "es2020",
    chunkSizeWarningLimit: 1500,
  },
});

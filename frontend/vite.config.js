import vue from "@vitejs/plugin-vue";
import vueJsx from "@vitejs/plugin-vue-jsx";
import wails from "@wailsio/runtime/plugins/vite";
import { codeInspectorPlugin } from "code-inspector-plugin";
import path from "path";
import { defineConfig } from "vite";
import topLevelAwait from "vite-plugin-top-level-await";
import { staticI18nPlugin } from "./plugins/static-i18n-plugin.js";

const isDev = process.env.NODE_ENV === "development";
const noAds = process.env.NO_ADS === "true";

// https://vitejs.dev/config/
export default defineConfig({
  define: {
    __NO_ADS__: JSON.stringify(noAds),
  },
  resolve: {
    alias: [
      {
        find: /^@\/services\/adApi$/,
        replacement: path.resolve(
          __dirname,
          noAds ? "./src/services/adApi.noads.js" : "./src/services/adApi.js",
        ),
      },
      { find: "@", replacement: path.resolve(__dirname, "./src") },
      { find: "@bindings", replacement: path.resolve(__dirname, "./bindings") },
    ],
  },
  build: {
    target: ["es2019", "safari13"],
    cssTarget: "safari13",
  },
  plugins: [
    isDev &&
      codeInspectorPlugin({
        bundler: "vite",
        editor: "code",
        hotKeys: ["ctrlKey"],
      }),
    wails("./bindings"),
    topLevelAwait(),
    staticI18nPlugin(),
    vue(),
    vueJsx(),
  ].filter(Boolean),
});

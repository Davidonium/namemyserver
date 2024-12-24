import * as path from "@std/path";
import { defineConfig } from "vite";
import deno from "@deno/vite-plugin";

export default defineConfig({
  server: {
    // TODO parameterize when in production
    origin: "http://127.0.0.1:8080",
  },
  build: {
    manifest: true,
    rollupOptions: {
      input: "./src/js/entries/app.js",
    },
  },
  resolve: {
    alias: {
      "~": path.join(import.meta.dirname!, "src"),
    },
  },
  plugins: [deno()],
});

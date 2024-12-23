import { defineConfig } from "vite";

export default defineConfig({
  server: {
    // TODO parameterize when in production
    origin: "http://127.0.0.1:8080",
  },
  build: {
    manifest: true,
    rollupOptions: {
      input: "./src/js/entries/app.ts",
    },
  },
});

import path from "node:path";
import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";

const __dirname = import.meta.dirname;

export default defineConfig({
  server: {
    // TODO parameterize when in production
    origin: "http://127.0.0.1:8080",
  },
  build: {
    manifest: true,
    rollupOptions: {
      input: "./src/js/entries/app.js",
      // remove eval warning when compiling the application because htmx uses it and there's no plan to not using it.
      // see https://github.com/bigskysoftware/htmx/pull/1988#issuecomment-1806290317
      onwarn: (entry, next) => {
        if (
          entry.loc?.file &&
          /htmx\.esm\.js$/.test(entry.loc.file) &&
          /Use of eval in/.test(entry.message)
        )
          return;
        return next(entry);
      },
    },
  },
  plugins: [tailwindcss()],
  resolve: {
    alias: {
      "~": path.resolve(__dirname, "./src/js"),
    },
  },
});

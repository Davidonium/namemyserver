const plugin = require('tailwindcss/plugin')

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["../internal/templates/**/*.templ"],
  theme: {
    extend: {},
  },
  plugins: [
    // credits: https://www.crocodile.dev/blog/css-transitions-with-tailwind-and-htmx
    // allows for modifiers in tailwind classes
    plugin(function ({ addVariant }) {
      addVariant('htmx-settling', ['&.htmx-settling', '.htmx-settling &'])
      addVariant('htmx-request', ['&.htmx-request', '.htmx-request &'])
      addVariant('htmx-swapping', ['&.htmx-swapping', '.htmx-swapping &'])
      addVariant('htmx-added', ['&.htmx-added', '.htmx-added &'])
    }),
  ],
};

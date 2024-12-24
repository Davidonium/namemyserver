import "~/css/app.css"
import "htmx.org";
import "vite/modulepreload-polyfill";
import { writeClipboardText } from "~/js/lib/clipboard.js"


document.body.addEventListener("htmx:load", (e) => {
  e.detail.elt.querySelectorAll(".js-copy").forEach((el) => {
    el.addEventListener("click", (elm) => {
      writeClipboardText(elm.currentTarget.dataset.copyValue);
    })
  })
})

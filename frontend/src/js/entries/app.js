import "../../css/app.css";
import "htmx.org";
import "vite/modulepreload-polyfill";
import { writeTextToClipboard } from "~/lib/clipboard";

document.body.addEventListener("htmx:load", (e) => {
  e.detail.elt.querySelectorAll(".js-copy").forEach((el) => {
    el.addEventListener("click", (elm) => {
      writeTextToClipboard(elm.currentTarget.dataset.copyValue);
    });
  });
});

import "../../css/app.css";
import "vite/modulepreload-polyfill";
import "htmx.org";
import { on, onLoad, find } from "htmx.org";
import { writeTextToClipboard } from "~/lib/clipboard";

onLoad((elm) => {
  const copyEl = find(elm, ".js-copy");
  if (!copyEl) {
    return;
  }

  on(copyEl, "click", (e) => {
    const target = e.currentTarget;
    writeTextToClipboard(target.dataset.copyValue);
    const checkmark = find(target, ".js-checkmark");
    checkmark.classList.remove("opacity-0");
    setTimeout(() => {
      checkmark.classList.add("opacity-0");
    }, 2000);
  });
});

on(".js-drawer-open", "click", () => {
  find("#drawer").classList.remove("translate-x-full")
});

on(".js-drawer-close", "click", () => {
  find("#drawer").classList.add("translate-x-full")
})

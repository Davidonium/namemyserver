import "../../css/app.css";
import "vite/modulepreload-polyfill";
import "htmx.org";
import { on, onLoad, find } from "htmx.org";
import { writeTextToClipboard } from "~/lib/clipboard";

onLoad((elm) => {
  registerCopyBtnCallback(elm);

  safeOn(".js-drawer-open", "click", () => {
    find("#drawer").classList.remove("translate-x-full")
  });

  safeOn(".js-drawer-close", "click", () => {
    find("#drawer").classList.add("translate-x-full")
  })
});


function registerCopyBtnCallback(elm) {
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

}

function safeOn(selector, eventName, callback) {
  const el = find(selector);
  if (!el) {
    return;
  }

  on(el, eventName, callback);
}

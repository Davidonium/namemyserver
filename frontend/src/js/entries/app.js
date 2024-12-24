import "../../css/app.css";
import "htmx.org";
import { on, onLoad, find } from "htmx.org";
import "vite/modulepreload-polyfill";
import { writeTextToClipboard } from "~/lib/clipboard";

onLoad((elm) => {
  const copyEl = find(elm, ".js-copy");
  if (!copyEl) {
    return;
  }

  on(copyEl, "click", (e) => {
    writeTextToClipboard(e.currentTarget.dataset.copyValue);
    e.currentTarget.classList.remove("opacity-0");
    setTimeout(() => {
      e.currentTarget.classList.add("opacity-0");
    });
  });
});

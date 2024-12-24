import "../../css/app.css";
import "htmx.org";
import { on, onLoad, find } from "htmx.org";
import "vite/modulepreload-polyfill";
import { writeTextToClipboard } from "~/lib/clipboard";

onLoad((elm) => {
  const els = find(elm, ".js-copy")
  if (!els) {
    return;
  }

  on(els, "click", (e) => {
    writeTextToClipboard(e.currentTarget.dataset.copyValue);
  })
})

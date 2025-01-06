import "../../css/app.css";
import "vite/modulepreload-polyfill";
import "htmx.org";
import u from "umbrellajs";
import { writeTextToClipboard } from "~/lib/clipboard";

u(document).on("htmx:load", (ev) => {
  const el = u(ev.currentTarget);
  el.find(".js-copy").on("click", (ev) => {
    const target = ev.currentTarget;
    writeTextToClipboard(target.dataset.copyValue);
    const checkmark = u(target).find(".js-checkmark");
    checkmark.removeClass("opacity-0");
    setTimeout(() => {
      checkmark.addClass("opacity-0");
    }, 2000);
  });

  el.find(".js-drawer-open").on("click", () => {
    u("#drawer").removeClass("translate-x-full", "opacity-0");
  });

  el.find(".js-drawer-close").on("click", () => {
    u("#drawer").addClass("translate-x-full");
  });
});

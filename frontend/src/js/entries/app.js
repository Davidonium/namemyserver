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

  const lengthValueEl = el.find(".js-length-range-value").first();
  el.find(".js-length-range-slider").on("input", (ev) => {
    lengthValueEl.textContent = ev.currentTarget.value;
  });

  el.find(".js-config-length-toggle").on("change", (ev) => {
    if (ev.currentTarget.checked) {
      el.find(".js-config-length-opacity").removeClass("opacity-40");
      el.find(".js-length-linked").each((el) => {
        el.disabled = false;
      });
    } else {
      el.find(".js-config-length-opacity").addClass("opacity-40");
      el.find(".js-length-linked").each((el) => {
        el.disabled = true;
      });
    }
  });

  el.find("#archiveButton").on("click", () => {
    u("#archiveDialog").first().showModal();
  });

  el.find("#recoverButton").on("click", () => {
    u("#recoverDialog").first().showModal();
  });

  el.find(".js-close-dialog").on("click", (ev) => {
    u(ev.currentTarget).closest("dialog").first().close();
  });

  // Bucket create page filter controls
  const filterLengthToggle = el.find(".js-filter-length-toggle").first();
  const filterLengthControls = el.find(".js-filter-length-controls").first();
  const filterLengthLinked = el.find(".js-filter-length-linked");
  const filterLengthSlider = el.find(".js-filter-length-range-slider").first();
  const filterLengthValue = el.find(".js-filter-length-range-value").first();

  if (filterLengthToggle) {
    filterLengthToggle.addEventListener("change", () => {
      const isEnabled = filterLengthToggle.checked;

      if (isEnabled) {
        filterLengthControls.classList.remove("opacity-40");
      } else {
        filterLengthControls.classList.add("opacity-40");
      }

      filterLengthLinked.each((input) => {
        input.disabled = !isEnabled;
      });
    });
  }

  if (filterLengthSlider && filterLengthValue) {
    filterLengthSlider.addEventListener("input", () => {
      filterLengthValue.textContent = filterLengthSlider.value;
    });
  }
});

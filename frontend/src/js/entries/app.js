import "../../css/app.css"
import "htmx.org";
import "vite/modulepreload-polyfill";

console.log("Javascript is loaded!");


document.body.addEventListener("htmx:load", (e) => {
  console.log("htmx load trigered");
  e.detail.elt.querySelectorAll(".js-copy").forEach((el) => {
    el.addEventListener("click", (elm) => {
      writeClipboardText(elm.currentTarget.dataset.copyValue);
    })
  })
})

async function writeClipboardText(text) {
  try {
    await navigator.clipboard.writeText(text);
  } catch (error) {
    console.error(error.message);
  }
}

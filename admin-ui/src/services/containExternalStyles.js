// Third-party scripts (e.g. the map widget) inject global CSS such as
// `a { color }`. Vuetify 4 keeps its styles in cascade layers, and unlayered
// CSS always beats layered CSS, so such rules would restyle the whole app
// until a reload (Vuetify 2's selectors were simply more specific).
// While active, this moves injected <style> contents into the low "external"
// layer (see public/layers.css). Returns a function that stops watching.
export function containExternalStyles() {
  const wrap = (node) => {
    if (node.nodeName !== "STYLE") return;
    if (node.hasAttribute("data-vite-dev-id")) return; // the app's own dev styles
    const css = node.textContent || "";
    // already layered (incl. Vuetify's own theme stylesheet) or empty
    if (!css.trim() || css.trimStart().startsWith("@layer")) return;
    node.textContent = `@layer external {\n${css}\n}`;
  };

  const observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      if (mutation.type === "childList") mutation.addedNodes.forEach(wrap);
      // a style tag filled after insertion
      if (mutation.type === "characterData" || mutation.target?.nodeName === "STYLE") {
        const target = mutation.target.nodeName === "STYLE" ? mutation.target : mutation.target.parentNode;
        if (target) wrap(target);
      }
    }
  });
  observer.observe(document.head, { childList: true, subtree: true, characterData: true });
  observer.observe(document.body, { childList: true, subtree: true });

  return () => observer.disconnect();
}

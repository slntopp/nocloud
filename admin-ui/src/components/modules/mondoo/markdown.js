// A small Markdown renderer for the documentation Mondoo attaches to findings
// (paragraphs, headings, lists, code, bold, links). Everything is escaped
// first, so the source can never inject markup.

const escape = (text) =>
  text.replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`);

function inline(text) {
  return escape(text)
    .replace(/`([^`]+)`/g, "<code>$1</code>")
    .replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>")
    .replace(
      /\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g,
      '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>'
    );
}

export function renderMarkdown(source) {
  const lines = String(source ?? "").replace(/\r/g, "").split("\n");
  const html = [];
  let paragraph = [];
  let list = null;

  const flush = () => {
    if (paragraph.length) html.push(`<p>${inline(paragraph.join(" "))}</p>`);
    if (list) {
      // A list interrupted by a code block continues its numbering.
      const start = list.start > 1 ? ` start="${list.start}"` : "";
      html.push(`<${list.tag}${start}>${list.items.join("")}</${list.tag}>`);
    }
    paragraph = [];
    list = null;
  };

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();

    if (trimmed.startsWith("```")) {
      flush();
      const code = [];
      while (++i < lines.length && !lines[i].trim().startsWith("```")) code.push(lines[i]);
      html.push(`<pre><code>${escape(code.join("\n"))}</code></pre>`);
      continue;
    }
    if (!trimmed) {
      // A blank line inside a list keeps the list going if more items follow.
      if (list && /^\s*([-*]|\d+\.)\s+/.test(lines[i + 1] ?? "")) continue;
      flush();
      continue;
    }
    const heading = trimmed.match(/^(#{1,6})\s+(.*)$/);
    if (heading) {
      flush();
      html.push(`<h4>${inline(heading[2])}</h4>`);
      continue;
    }
    const item = trimmed.match(/^([-*]|\d+\.)\s+(.*)$/);
    if (item) {
      const tag = /\d/.test(item[1]) ? "ol" : "ul";
      if (paragraph.length || (list && list.tag !== tag)) flush();
      list ??= { tag, items: [], start: parseInt(item[1]) || 1 };
      list.items.push(`<li>${inline(item[2])}</li>`);
      continue;
    }
    if (list) {
      // Continuation of the previous list item.
      list.items[list.items.length - 1] = list.items.at(-1).replace(/<\/li>$/, ` ${inline(trimmed)}</li>`);
      continue;
    }
    paragraph.push(trimmed);
  }
  flush();
  return html.join("");
}

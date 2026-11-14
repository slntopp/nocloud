// SVG icons that used to come from the nocloud-ui package, loaded as URLs on demand
const icons = import.meta.glob("@/assets/nocloud-icons/*.svg", {
  query: "?url",
  import: "default",
});

const nameOf = (path) => path.split("/").pop().replace(".svg", "");
const loaders = Object.fromEntries(
  Object.entries(icons).map(([path, load]) => [nameOf(path), load]),
);

export const nocloudIconNames = Object.keys(loaders);

export function loadNocloudIcon(name) {
  const load = loaders[name];
  return load ? load() : Promise.reject(new Error(`unknown icon ${name}`));
}

// all @ant-design/icons-vue icon names (CloudOutlined, CloudFilled, ...)
export async function loadAntIcons() {
  return import("@ant-design/icons-vue");
}

export const isAntIconName = (name) => /(Outlined|Filled|TwoTone)$/.test(name);

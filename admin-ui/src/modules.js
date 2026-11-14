import { defineAsyncComponent } from "vue";

// Per-provider components live in src/components/modules/<type>/<kind>.vue.
// import.meta.glob replaces webpack's require.context: the keys tell which
// types implement a kind, the values load the component lazily.
const kinds = {
  instanceCreate: import.meta.glob("@/components/modules/*/instanceCreate.vue"),
  instanceCard: import.meta.glob("@/components/modules/*/instanceCard.vue"),
  serviceProviders: import.meta.glob(
    "@/components/modules/*/serviceProviders.vue",
  ),
  serviceCreate: import.meta.glob("@/components/modules/*/serviceCreate.vue"),
  additionalInstanceInfo: import.meta.glob(
    "@/components/modules/*/additionalInstanceInfo.vue",
  ),
  billingInfo: import.meta.glob("@/components/modules/*/billingInfo.vue"),
  billingLabel: import.meta.glob("@/components/modules/*/billingLabel.vue"),
  serviceProviderVars: import.meta.glob(
    "@/components/modules/*/serviceProviderVars.vue",
  ),
};

const typeOf = (path) => path.split("/").slice(-2)[0];
const cache = new Map();

// Types that have a component of the given kind, e.g. moduleTypes("instanceCreate")
export function moduleTypes(kind) {
  return Object.keys(kinds[kind]).map(typeOf);
}

// Async component for a type, or null when the type has no such component
export function moduleComponent(kind, type) {
  const key = `${kind}/${type}`;
  if (cache.has(key)) return cache.get(key);

  const entry = Object.entries(kinds[kind]).find(
    ([path]) => typeOf(path) === type,
  );
  const component = entry ? defineAsyncComponent(entry[1]) : null;

  cache.set(key, component);
  return component;
}

// { [type]: asyncComponent } for every type that implements the kind
export function moduleComponents(kind) {
  return Object.fromEntries(
    moduleTypes(kind).map((type) => [type, moduleComponent(kind, type)]),
  );
}

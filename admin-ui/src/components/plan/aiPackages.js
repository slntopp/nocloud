/**
 * AI packages are the products of an "empty" plan whose meta carries the keys the OpenAI driver
 * reads: ai_credit (NCU per paid period), ai_seats (how many people may spend it; absent is the
 * whole organization) and ai_models (the model keys it pays for; absent is every model).
 * Constructor packages (`ai_site`) may also split models into survey (`ai_models_survey`) and
 * build (`ai_models_build`); `ai_models` stays the union so billing still covers both.
 * The storefront reads one more: ai_promocode, the code it applies to this package on its own.
 */
export const AI_KEYS = [
  "ai_credit",
  "ai_seats",
  "ai_models",
  "ai_models_survey",
  "ai_models_build",
  "ai_site",
  "ai_promocode",
];

export const DEFAULT_PERIOD = 3600 * 24 * 30;

const KEY_PATTERN = /^[a-z0-9][a-z0-9_-]*$/;

/** A key the driver reads, from meta or, as an earlier mistake left it, from resources. */
const aiValue = (product, key) => product?.meta?.[key] ?? product?.resources?.[key];

const modelList = (value) =>
  Array.isArray(value) ? value.filter((m) => typeof m === "string" && m) : [];

const uniqueModels = (keys) => [...new Set(keys.filter((m) => typeof m === "string" && m))];

export const isAiPackage = (product) => Number(aiValue(product, "ai_credit")) > 0;

/** Whether some AI keys sit in resources, where the driver does not read them. */
export const hasMisplacedKeys = (product) =>
  AI_KEYS.some((key) => product?.resources?.[key] !== undefined);

export function toForm(key, product = {}) {
  const seats = Math.floor(Number(aiValue(product, "ai_seats")) || 0);
  const models = modelList(aiValue(product, "ai_models"));
  const site = Boolean(aiValue(product, "ai_site"));
  const survey = modelList(aiValue(product, "ai_models_survey"));
  const build = modelList(aiValue(product, "ai_models_build"));
  /** Older constructor packages only had `ai_models`; seed both roles from it. */
  const surveyModels = site && survey.length === 0 && build.length === 0 ? [...models] : survey;
  const buildModels = site && survey.length === 0 && build.length === 0 ? [...models] : build;
  return {
    key,
    title: product.title ?? "",
    price: Number(product.price) || 0,
    period: Number(product.period) || DEFAULT_PERIOD,
    periodKind: product.periodKind || (product.period ? "DEFAULT" : "CALENDAR_MONTH"),
    credit: Number(aiValue(product, "ai_credit")) || 0,
    seats: seats > 0 ? seats : "",
    models,
    surveyModels,
    buildModels,
    site,
    promocode: aiValue(product, "ai_promocode") ?? "",
    public: key ? Boolean(product.public) : product.public !== false,
    sorter: Number(product.sorter) || 0,
  };
}

const withoutAiKeys = (object) => {
  const rest = { ...object };
  AI_KEYS.forEach((key) => delete rest[key]);
  return rest;
};

/**
 * The product the form describes, over the one it edits: every field the form does not know
 * (descriptionId, group, addons, other meta and resources) is kept, and AI keys found in resources
 * move to meta.
 */
export function toProduct(form, original = {}) {
  const meta = withoutAiKeys(original.meta);
  meta.ai_credit = Number(form.credit);
  if (Number(form.seats) > 0) {
    meta.ai_seats = Math.floor(Number(form.seats));
  }
  if (form.site) {
    meta.ai_site = true;
    const survey = uniqueModels(form.surveyModels || []);
    const build = uniqueModels(form.buildModels || []);
    if (survey.length > 0) {
      meta.ai_models_survey = survey;
    }
    if (build.length > 0) {
      meta.ai_models_build = build;
    }
    const billed = uniqueModels([...survey, ...build]);
    if (billed.length > 0) {
      meta.ai_models = billed;
    }
  } else if (form.models.length > 0) {
    meta.ai_models = uniqueModels(form.models);
  }
  if (form.promocode.trim()) {
    meta.ai_promocode = form.promocode.trim().toUpperCase();
  }
  const product = {
    ...original,
    kind: original.kind || "PREPAID",
    title: form.title.trim(),
    price: Number(form.price),
    period: Number(form.period),
    periodKind: form.periodKind || "DEFAULT",
    public: Boolean(form.public),
    sorter: Number(form.sorter) || 0,
    meta,
  };
  if (original.resources) {
    product.resources = withoutAiKeys(original.resources);
  }
  return product;
}

/** Why the form cannot be applied, or "" when it can. */
export function formError(form, otherKeys) {
  if (!KEY_PATTERN.test(form.key)) {
    return "The key takes lowercase latin letters, digits, - and _";
  }
  if (otherKeys.includes(form.key)) {
    return "Another product of this plan has this key";
  }
  if (!form.title.trim()) {
    return "The package needs a title";
  }
  if (!(Number(form.price) >= 0)) {
    return "The price cannot be negative";
  }
  if (!(Number(form.period) > 0)) {
    return "The package needs a period";
  }
  if (!(Number(form.credit) > 0)) {
    return "The credit must be above 0 NCU";
  }
  if (form.seats !== "" && !(Number.isInteger(Number(form.seats)) && Number(form.seats) >= 1)) {
    return "Seats are a whole number from 1, or empty for no limit";
  }
  return "";
}

/** A free product key after the given one: "broke-boy" → "broke-boy-copy", "-copy-2", … */
export function copyKey(key, taken) {
  let candidate = `${key}-copy`;
  for (let n = 2; taken.includes(candidate); n++) {
    candidate = `${key}-copy-${n}`;
  }
  return candidate;
}

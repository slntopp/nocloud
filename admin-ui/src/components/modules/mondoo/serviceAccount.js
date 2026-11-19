// Mondoo service account credentials as the driver accepts them: the
// credentials object, its JSON, or the base64 of that JSON.
export function parseServiceAccount(value) {
  if (!value) return { account: null, error: "Service account is required" };
  if (typeof value === "object") return check(value);

  let text = String(value).trim();
  if (!text.startsWith("{")) {
    try {
      text = atob(text.replace(/\s+/g, ""));
    } catch {
      return { account: null, error: "Neither JSON nor base64" };
    }
  }
  try {
    return check(JSON.parse(text));
  } catch {
    return { account: null, error: "Service account is not valid JSON" };
  }
}

function check(account) {
  if (!account?.mrn || !account?.private_key) {
    return { account: null, error: "Service account must have mrn and private_key" };
  }
  return { account, error: null };
}

export const importModes = [
  {
    value: "",
    title: "Off: always create a new integration",
    hint: "Each instance gets its own new integration. Use this for a space shared between customers.",
  },
  {
    value: "prefer",
    title: "Prefer: adopt a matching integration, otherwise create",
    hint: "An instance adopts the integration that already exists for exactly its domains. Not for shared spaces: a customer could claim another's results by naming their domain.",
  },
  {
    value: "only",
    title: "Only: adopt a matching integration, never create",
    hint: "Instances only adopt existing integrations and never change them. Works with a Viewer service account. Not for shared spaces.",
  },
];

export const ratings = [
  { key: "CRITICAL", title: "Critical", color: "#b71c1c" },
  { key: "HIGH", title: "High", color: "#e65100" },
  { key: "MEDIUM", title: "Medium", color: "#f9a825" },
  { key: "LOW", title: "Low", color: "#1565c0" },
  { key: "NONE", title: "None", color: "#757575" },
];

export const ratingColor = (rating) =>
  ratings.find((r) => r.key === rating)?.color ?? "#757575";

import api from "@/api.js";

// Services provider secrets come from the API (GET /sp/:uuid, create, update)
// with every value replaced by this placeholder. Sent back in an update or a
// config test, it keeps the stored value.
export const REDACTED_SECRET = "_nocloud.redacted";
// Sent as a secret's value in an update, it removes that key.
export const REMOVE_SECRET = "_nocloud.remove";

export const isRedacted = (value) => value === REDACTED_SECRET;

export const hasRedactedSecrets = (secrets) =>
  Object.values(secrets ?? {}).some(isRedacted);

// The provider's real secret values. Only root may read them; anyone else
// gets a 403.
export async function fetchSpSecrets(uuid) {
  const response = await api.get(`/sp/${uuid}/secrets`);
  return response?.secrets ?? {};
}

// The secrets to send in an update: what the editor holds, plus a removal for
// every stored key the editor dropped (the core keeps keys that are left out).
export function secretsPatch(original, edited) {
  const patch = { ...(edited ?? {}) };
  for (const key of Object.keys(original ?? {})) {
    if (!(key in patch)) patch[key] = REMOVE_SECRET;
  }
  return patch;
}

// The two things a Mondoo product can sell, as product meta.kind. These are
// the names used everywhere: plans, instances, docs.
export const productKinds = [
  {
    value: "attack_surface",
    title: "Domain scanning",
    subtitle: "One domain and its subdomains, scanned from outside",
  },
  {
    value: "ssh",
    title: "SSH scanning",
    subtitle: "One of the customer's servers, scanned over SSH",
  },
];

export function productKind(product) {
  return product?.meta?.kind === "ssh" ? "ssh" : "attack_surface";
}

export function kindTitle(product) {
  return productKinds.find(({ value }) => value === productKind(product)).title;
}

const month = 30 * 24 * 3600;
const week = 7 * 24 * 3600;

// Starting points for a new product. The price is left at 0 for the admin to
// set. Each kind has its own trial group, so a customer gets one free trial
// of each, once per account: the first domain on trial, the next one paid.
export function productPreset(kind) {
  const trial = {
    period: week,
    scope: "account",
    invoice_before: 2 * 24 * 3600,
    group: kind === "ssh" ? "mondoo-ssh" : "mondoo-domain",
  };
  if (kind === "ssh") {
    return {
      key: "ssh_scan",
      title: "SSH scanning",
      kind: "PREPAID",
      price: 0,
      period: month,
      resources: {},
      public: true,
      meta: { kind: "ssh", schedule: "24h", trial },
    };
  }
  return {
    key: "domain_scan",
    title: "Domain scanning",
    kind: "PREPAID",
    price: 0,
    period: month,
    resources: {},
    public: true,
    meta: { kind: "attack_surface", trial },
  };
}

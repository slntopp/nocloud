import { computed, onMounted, ref } from "vue";
import Cookies from "js-cookie";
import api from "@/api";
import config from "@/config";
import { useStore } from "@/store";

// ponytail: the billing info and the billing label mount this hook for the same instance at
// once, and every usage read mints a client session. Share the in-flight request per key and
// drop it when it settles, so a page reload still refetches.
const inFlight = new Map();
const once = (key, fn) => {
  let request = inFlight.get(key);
  if (!request) {
    request = fn();
    inFlight.set(key, request);
    request.catch(() => {}).finally(() => inFlight.delete(key));
  }
  return request;
};

/**
 * An instance's AI package (a product with meta.ai_credit) as the billing info shows it: the credit
 * of the current period, how much of it is spent, the seats and models, and signing into the chat
 * as the client. Everything is empty for any other product.
 */
export default function useAiPackage(template, account) {
  const store = useStore();

  /** The instance's product when it is an AI package: one with meta.ai_credit. */
  const product = computed(() => {
    const p = template.value.billingPlan?.products?.[template.value.product];
    return +p?.meta?.ai_credit > 0 ? p : null;
  });
  const credit = computed(() => +product.value?.meta?.ai_credit || 0);
  const models = computed(() =>
    Array.isArray(product.value?.meta?.ai_models)
      ? product.value.meta.ai_models
      : []
  );
  const seatsText = computed(() => {
    const seats = +product.value?.meta?.ai_seats || 0;
    if (!seats) return "Whole organization";
    const picked = template.value.config?.ai_members?.length || 0;
    // the driver gives a package nobody was picked for to its payer
    const members = usage.value?.members?.length ?? (picked || 1);
    return `${members} / ${seats}`;
  });

  const usage = ref(null);
  const usageError = ref("");
  const isUsageLoading = ref(false);
  const isChatLoading = ref(false);
  const modelNames = ref({});

  const amount = (value) =>
    value > 0 && value < 0.01 ? "< 0.01" : (+value || 0).toFixed(2);

  // ponytail: the driver answers /packages for the token's own account only, so the usage is read
  // with the client's token; a driver endpoint for admins would spare minting a client session
  const clientToken = async () => {
    const { token } = await store.dispatch("auth/loginToApp", {
      uuid: template.value.account ?? account.value?.uuid,
      type: "whmcs",
    });
    return token;
  };

  const fetchUsage = async () => {
    isUsageLoading.value = true;
    try {
      const data = await once(`packages:${template.value.uuid}`, async () => {
        const token = await clientToken();
        const { data } = await api.axios.get("/api/openai/packages", {
          headers: { Authorization: `Bearer ${token}` },
        });
        return data;
      });
      usage.value =
        data.packages?.find((p) => p.instance === template.value.uuid) ?? null;
      if (!usage.value) {
        // the driver lists only paid, running, not suspended periods, and a one-time package
        // until its credit is spent
        usageError.value = +product.value.period > 0
          ? "No active period (stopped, suspended or unpaid)"
          : "Credit spent, or stopped or suspended";
      }
    } catch (e) {
      usageError.value = `Failed to read usage: ${
        e.response?.data?.answer || e.response?.data?.message || e.message
      }`;
    } finally {
      isUsageLoading.value = false;
    }
  };

  const fetchModelNames = async () => {
    if (!models.value.length) return;
    try {
      const { cfg } = await once("openai-config", () =>
        api.get("/api/openai/get_config")
      );
      modelNames.value = Object.fromEntries(
        Object.entries(cfg?.models ?? {}).map(([key, m]) => [key, m.name])
      );
    } catch {
      // the chips fall back to the model keys
    }
  };

  const chatUrl = `${config.aiChatUrl.replace(/\/$/, "")}/oauth/nocloud`;

  /** The domain this page and the chat share (api.support.by and ai.support.by → support.by), or "". */
  const sharedDomain = (a, b) => {
    const x = a.split(".").reverse();
    const y = b.split(".").reverse();
    let n = 0;
    while (n < x.length && x[n] === y[n]) n++;
    return n >= 2 ? x.slice(0, n).reverse().join(".") : "";
  };

  /**
   * Signs into the chat as the client, as nocloud-app does: the client's token in LibreChat's SSO
   * cookie on the shared domain for a minute, then /oauth/nocloud. Without a shared domain the chat
   * shows its login form.
   */
  const openChat = async () => {
    // opened before the await, or the browser blocks it as a popup
    const win = window.open("", "_blank");
    isChatLoading.value = true;
    try {
      const token = await clientToken();
      const domain = sharedDomain(location.hostname, new URL(chatUrl).hostname);
      if (domain) {
        Cookies.set("nocloud_token", token, {
          domain,
          expires: 1 / 1440,
          secure: location.protocol === "https:",
          sameSite: "lax",
        });
      }
      win.opener = null;
      win.location = chatUrl;
    } catch (e) {
      win?.close();
      store.commit("snackbar/showSnackbarError", {
        message: e.response?.data?.message || "Error during login to AI chat",
      });
    } finally {
      isChatLoading.value = false;
    }
  };

  onMounted(() => {
    if (!product.value) return;
    fetchUsage();
    fetchModelNames();
  });

  return {
    isAiPackage: computed(() => !!product.value),
    credit,
    usage,
    usageError,
    isUsageLoading,
    seatsText,
    models,
    modelNames,
    amount,
    openChat,
    isChatLoading,
  };
}

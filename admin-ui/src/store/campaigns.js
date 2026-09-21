import api from "@/api.js";
import store from "@/store/index.js";

// The campaigns plugin keeps a journal of everything that has gone out to a
// customer: the tickets its own chains opened, and the letters nocloud asked
// WHMCS to send. This module is what the notifications column reads.
//
// Nothing here is configured by hand. The plugin's address is already in
// nocloud's settings, where plugins are listed as {title, url, icon}, and the
// url of this one ends in /campaigns.ui/ — the API sits at its root. No
// plugin in the list, no column: other installations do not have it.
const uiSuffix = "campaigns.ui";

function base() {
  const plugin = (store.getters["plugins/all"] || []).find((p) =>
    (p.url || "").includes(uiSuffix),
  );
  if (!plugin) return null;

  try {
    const url = new URL(plugin.url, location.origin);
    return url.origin + url.pathname.slice(0, url.pathname.indexOf(uiSuffix));
  } catch {
    return null;
  }
}

// connect-RPC speaks JSON over a plain POST, so no client library is needed:
// the method is a path, the message is the body. The token is the one axios
// already carries for nocloud — the plugin checks it with the same signing
// key and asks for the same admin rights.
async function call(method, body) {
  const url = base();
  if (!url) throw new Error("the campaigns plugin is not installed");

  const { data } = await api.axios.post(
    `${url}campaigns.SendsAPI/${method}`,
    body,
    {
      headers: { "Content-Type": "application/json" },
    },
  );
  return data;
}

// A table draws its rows one by one, and a page is twenty five of them. Each
// cell asks for its own count, and these collect into one request: a promise
// per uuid, a single call once the page has finished asking.
const pending = { accounts: new Map(), instances: new Map() };
let flushing = null;

function flush() {
  flushing = null;

  for (const kind of ["accounts", "instances"]) {
    const waiting = pending[kind];
    if (!waiting.size) continue;

    const uuids = [...waiting.keys()];
    const resolvers = [...waiting.values()];
    waiting.clear();

    call("Notifications", { [kind]: uuids })
      .then(({ counts = [] }) => {
        const found = new Map(counts.map((c) => [c.uuid, c]));
        uuids.forEach((uuid, i) =>
          // Everything asked about comes back, so a uuid missing from the
          // answer means nothing was written to it rather than that the
          // question was not asked.
          resolvers[i](
            found.get(uuid) || { uuid, letters: 0, tickets: 0, last: 0 },
          ),
        );
      })
      .catch(() => resolvers.forEach((resolve) => resolve(null)));
  }
}

function ask(kind, uuid) {
  return new Promise((resolve) => {
    pending[kind].set(uuid, resolve);
    if (!flushing) flushing = setTimeout(flush, 30);
  });
}

export default {
  namespaced: true,
  state: {
    // What has already been counted, so paging back and forth does not ask
    // again: { [uuid]: { letters, tickets, last } }.
    counts: {},
    // The last ten of a subject, fetched when its tag is first hovered.
    recent: {},
  },
  getters: {
    // Whether the plugin is installed at all — the column hides itself
    // otherwise rather than showing a row of failures.
    installed: () => base() !== null,
    count: (state) => (uuid) => state.counts[uuid],
    recent: (state) => (uuid) => state.recent[uuid],
  },
  mutations: {
    setCount(state, { uuid, count }) {
      state.counts = { ...state.counts, [uuid]: count };
    },
    setRecent(state, { uuid, rows }) {
      state.recent = { ...state.recent, [uuid]: rows };
    },
  },
  actions: {
    // count answers one row of a table, batching with whatever else the page
    // is asking for at the same moment.
    async count({ state, commit, getters }, { uuid, kind = "accounts" }) {
      if (!uuid || !getters.installed) return null;
      if (state.counts[uuid]) return state.counts[uuid];

      const count = await ask(kind, uuid);
      if (count) commit("setCount", { uuid, count });
      return count;
    },

    // recent is what the tag shows when hovered: the last ten, both halves of
    // the journal together, newest first.
    async recent({ state, commit, getters }, { uuid, kind = "accounts" }) {
      if (!uuid || !getters.installed) return [];
      if (state.recent[uuid]) return state.recent[uuid];

      const filter =
        kind === "instances" ? { instance: uuid } : { account: uuid };
      const { sends = [] } = await call("List", { ...filter, limit: 10 });

      commit("setRecent", { uuid, rows: sends });
      return sends;
    },
  },
};

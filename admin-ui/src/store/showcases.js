import api from "@/api.js";
import { createPromiseClient } from "@connectrpc/connect";
import { ShowcaseCategoriesService } from "nocloud-proto/proto/es/services_providers/services_providers_connect";
import { createListCache } from "@/store/cache.js";

// categories are loaded together with showcases, so their changes count too
const showcasesCache = createListCache({
  invalidateOn: /^\/?(showcases|showcase_categories)(\/|\?|$)/,
});

const load = async (commit, params) => {
  commit("setShowcases", []);
  commit("setCategories", []);
  commit("setLoading", true);

  params = { ...(params ?? {}), omitPromos: false };

  try {
    const [{ showcases }, { categories }] = await Promise.all([
      api.showcases.list(params),
      api.get("showcase_categories"),
    ]);

    showcases.forEach((showcase) => {
      showcase.categories = categories.reduce((acc, category) => {
        if (category.showcases?.includes(showcase.uuid)) {
          acc.push(category);
        }
        return acc;
      }, []);
    });

    commit("setShowcases", showcases);
    commit("setCategories", categories);
  } catch (e) {
    throw new Error(e);
  } finally {
    commit("setLoading", false);
  }
};

export default {
  namespaced: true,
  state: {
    showcases: [],
    categories: [],
    loading: false,
  },
  mutations: {
    setShowcases(state, showcases) {
      state.showcases = showcases;
    },
    pushShowcase(state, showcase) {
      const index = state.showcases.findIndex((a) => a.uuid === showcase.uuid);

      if (index !== -1) {
        state.showcases[index] = showcase;
      } else {
        state.showcases.push(showcase);
      }
    },
    removeShowcase(state, uuid) {
      state.showcases = state.showcases.filter((s) => s.uuid !== uuid);
    },
    replaceShowcase(state, value) {
      state.showcases = state.showcases.map((s) =>
        s.uuid === value.uuid ? value : s
      );
    },
    repaceCategory(state, value) {
      state.categories = state.categories.map((s) =>
        s.uuid === value.uuid ? value : s
      );
    },
    setLoading(state, data) {
      state.loading = data;
    },
    setCategories(state, categories) {
      state.categories = categories;
    },
  },
  actions: {
    // Always goes to the server (reload button, after changes)
    fetch({ commit }, params) {
      return showcasesCache.run(params, () => load(commit, params), {
        force: true,
      });
    },
    // Reuses the loaded list if it's fresh and was loaded with the same params
    ensure({ commit }, params) {
      return showcasesCache.run(params, () => load(commit, params));
    },
    async fetchById({ commit }, id) {
      commit("setLoading", true);

      try {
        const response = await api.showcases.get(id);
        commit("pushShowcase", response);
      } catch (e) {
        throw new Error(e);
      } finally {
        commit("setLoading", false);
      }
    },
    async delete({ commit }, uuid) {
      await api.showcases.delete(uuid);
      commit("removeShowcase", uuid);
    },
  },
  getters: {
    all(state) {
      return state.showcases;
    },
    isLoading(state) {
      return state.loading;
    },
    categories(state) {
      return state.categories;
    },
    showcaseCategoriesClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(
        ShowcaseCategoriesService,
        rootGetters["app/transport"]
      );
    },
  },
};

import api from "@/api";
import { formatToYYMMDD } from "@/functions";
import qs from 'qs';

const getCacheKey = (params) => JSON.stringify(params);

export default {
  namespaced: true,
  state: {
    cached: {},
    loading: false,
  },
  mutations: {
    setToCached(state, { key, value }) {
      state.cached[key] = value;
    },
    clearCache(state) {
      state.cached = {};
    },
    setLoading(state, val) {
      state.loading = !!val;
    },
  },
  actions: {
    async fetch({ commit, state }, params) {
      const cacheKey = getCacheKey(params);

      if (state.cached[cacheKey]) {
        return state.cached[cacheKey];
      }

      try {
        commit("setLoading", true);
        console.log(params.params);

        commit("setToCached", {
          key: cacheKey,
          value: api.get(`/statistic/${params.entity}`, {
            params: params.params,
            paramsSerializer: (params) =>
              qs.stringify(params, { arrayFormat: "repeat" }),
          }),
        });

        const response = await state.cached[cacheKey];

        commit("setToCached", { key: cacheKey, value: response });

        return response;
      } catch {
        commit("setToCached", { key: cacheKey, value: null });
      } finally {
        commit(
          "setLoading",
          Object.keys(state.cached).some(
            (key) => state.cached[key] instanceof Promise
          )
        );
      }
    },
    get(_, params) {
      return api.get(`/statistic/${params.entity}`, {
        params: params.params,
      });
    },
    async getForChart({ dispatch }, { periods, periodType, entity, params }) {
      let interval = "1 day";

      if (periodType.split("-")[1]) {
        interval = periodType.split("-")[1].replace("_", " ");
      }

      const dataParams = {
        entity,
        params: {
          ...(params || {}),
          with_timeseries: true,
          bucket_interval: interval,
        },
      };

      const data = await Promise.all(
        periods.map((period) => {
          dataParams.params.start_date = formatToYYMMDD(period[0]);
          dataParams.params.end_date = formatToYYMMDD(period[1]);

          return dispatch("fetch", dataParams);
        })
      );

      data.forEach((data) => {
        if (!data.timeseries) {
          data.timeseries = [];
        }

        if (!data.summary) {
          data.summary = {};
        }
      });

      return data;
    },
  },
  getters: {
    loading(state) {
      return state.loading;
    },
    cached(state) {
      return (params) => {
        const cacheKey = getCacheKey(params);

        return state.cached[cacheKey];
      };
    },
  },
};

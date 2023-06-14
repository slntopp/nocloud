import api from "@/api";

const currencyRate = {
  data: () => ({ rate: null }),
  computed: {
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
  },
  methods: {
    fetchRate(currency = "PLN") {
      api
        .get(`/billing/currencies/rates/${currency}/${this.defaultCurrency}`)
        .then((res) => {
          this.rate = res.rate;
        })
        .catch(() =>
          api.get(
            `/billing/currencies/rates/${this.defaultCurrency}/${currency}`
          )
        )
        .then((res) => {
          if (res) this.rate = 1 / res.rate;
        })
        .catch((err) => console.error(err));
    },
  },
};

export default currencyRate;

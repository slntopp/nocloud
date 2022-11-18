<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-icon class="group-icon">mdi-format-list-group</v-icon>
    <v-btn @click="editPlan">Save</v-btn>
    <v-select
      label="Plan"
      item-text="title"
      item-value="uuid"
      class="d-inline-block ml-4"
      v-model="nocloudPlan"
      :items="nocloudPlans"
    />

    <v-row>
      <v-col lg="6" md="12">
        <nocloud-table
          item-key="id"
          group-by="name"
          :show-expand="true"
          :show-select="false"
          :show-group-by="true"
          :items="plans"
          :headers="headers"
          :expanded.sync="expanded"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.price.value`]="{ item, value }">
            {{ value }} {{ 'NCU' || item.price.currencyCode }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field v-model="item.value" />
          </template>
          <template v-slot:[`item.sell`]="{ item }">
            <v-switch v-model="item.sell" />
          </template>
          <template v-slot:expanded-item="{ headers, item }">
            <template v-if="item.windows">
              <td></td>
              <td :colspan="headers.length - 4">{{ item.windows.name }}</td>
              <td>
                {{ item.windows.price.value }}
                {{ 'NCU' || item.price.currencyCode }}
              </td>
              <td><v-text-field v-model="item.windows.value" /></td>
              <td></td>
            </template>
            <template v-else>
              <td></td>
              <td :colspan="headers.length - 1">{{ $t('Windows is none') }}</td>
            </template>
          </template>
        </nocloud-table>
      </v-col>
      <v-col lg="6" md="12">
        <nocloud-table
          group-by="name"
          :show-select="false"
          :show-group-by="true"
          :items="addons"
          :headers="addonsHeaders"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.price.value`]="{ item, value }">
            {{ value }} {{ 'NCU' || item.price.currencyCode }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field v-model="item.value" />
          </template>
          <template v-slot:[`item.sell`]="{ item }">
            <v-switch v-model="item.sell" />
          </template>
        </nocloud-table>
      </v-col>
    </v-row>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </v-card>
</template>

<script>
import nocloudTable from "@/components/table.vue";
import snackbar from "@/mixins/snackbar.js";
import api from "@/api.js";

export default {
  name: 'sevices-provider-table',
  components: { nocloudTable },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    plans: [],
    expanded: [],
    headers: [
      { text: '', value: 'data-table-expand', groupable: false },
      { text: 'Plan', value: 'name', sortable: false, class: 'groupable' },
      { text: 'Duration', value: 'duration', sortable: false, class: 'groupable' },
      { text: 'OVH price', value: 'price.value', groupable: false },
      { text: 'Price', value: 'value', groupable: false, width: 150 },
      { text: 'Sell', value: 'sell', sortable: false, class: 'groupable', width: 100 }
    ],

    addons: [],
    addonsHeaders: [
      { text: 'Addon', value: 'name', sortable: false, class: 'groupable' },
      { text: 'Duration', value: 'duration', sortable: false, class: 'groupable' },
      { text: 'OVH price', value: 'price.value', groupable: false },
      { text: 'Price', value: 'value', groupable: false, width: 150 },
      { text: 'Sell', value: 'sell', sortable: false, class: 'groupable', width: 100 }
    ],

    isPlansLoading: false,
    nocloudPlan: '',
    fetchError: ''
  }),
  methods: {
    changeIcon() {
      setTimeout(() => {
        const headers = document.querySelectorAll('.groupable');

        headers.forEach(({ lastElementChild }) => {
          const icon = document.querySelector('.group-icon').cloneNode(true);

          lastElementChild.innerHTML = '';
          lastElementChild.append(icon);

          icon.style = 'display: inline-flex';
          icon.addEventListener('click', () => {
            this.changeClose();
            this.changeIcon();
          });
        });
      }, 100);
    },
    changeClose() {
      setTimeout(() => {
        const close = document.querySelectorAll('.v-row-group__header .v-btn__content');

        close.forEach((element) => {
          element.addEventListener('click', this.changeIcon);
        });
      }, 100);
    },
    changePlans({ plans, windows }) {
      const result = [];

      plans.forEach(({ prices, planCode, productName }) => {
        prices.forEach(({ pricingMode, price, duration }) => {
          const isMonthly = duration === 'P1M' && pricingMode === 'default';
          const isYearly = duration === 'P1Y' && pricingMode === 'upfront12';

          if (isMonthly || isYearly) {
            const code = planCode.split('-').slice(1).join('-');
            const option = windows.find((el) => el.planCode.includes(code));
            const plan = { windows: null };

            if (option) {
              const { price: { value } } = option.prices.find((el) =>
                el.duration === duration && el.pricingMode === pricingMode);

              plan.windows = {
                value, price: { value },
                name: option.productName,
                code: option.planCode
              };
            }

            result.push({
              ...plan,
              planCode,
              pricingMode,
              price,
              duration,
              name: productName,
              value: price.value,
              sell: false,
              id: `${duration} ${planCode}`
            });
          }
        });
      });
      this.plans = result;

      this.plans.sort((a, b) => {
        const resA = a.planCode.split('-');
        const resB = b.planCode.split('-');

        const isCpuEqual = resB.at(-3) === resA.at(-3);
        const isRamEqual = resB.at(-2) === resA.at(-2);

        if (isCpuEqual && isRamEqual) return resA.at(-1) - resB.at(-1);
        if (isCpuEqual) return resA.at(-2) - resB.at(-2);
        return resA.at(-3) - resB.at(-3);
      });
    },
    changeAddons({ backup, disk, snapshot }) {
      const result = [];

      [backup, disk, snapshot].forEach((el) => {
        el.forEach(({ prices, planCode, productName }) => {
          prices.forEach(({ pricingMode, price, duration }) => {
            const isMonthly = duration === 'P1M' && pricingMode === 'default';
            const isYearly = duration === 'P1Y' && pricingMode === 'upfront12';

            if (isMonthly || isYearly) {
              result.push({
                planCode,
                pricingMode,
                price,
                duration,
                name: productName,
                value: price.value,
                sell: false,
                id: `${duration} ${planCode}`
              });
            }
          });
        });
      });

      this.addons = result;
    },
    editPlan() {
      if (!this.nocloudPlan) {
        this.showSnackbarError({ message: 'Please select plan' });
        return;
      }
      const plan = {
        ...this.nocloudPlans.find(({ uuid }) => uuid === this.nocloudPlan),
        resources: [], products: {}
      };

      this.plans.forEach((el) => {
        if (el.sell) {
          const [,, cpu, ram, disk] = el.planCode.split('-');
          const meta = {};

          if (el.windows) meta[el.windows.code] = el.windows.value;
          plan.products[el.id] = {
            kind: 'PREPAID',
            title: el.name,
            price: el.value,
            period: this.getPeriod(el.duration),
            resources: { cpu: +cpu, ram: ram * 1024, disk: disk * 1024 },
            sorter: Object.keys(plan.products).length,
            meta
          }
        }
      });

      this.addons.forEach((el) => {
        if (el.sell) {
          plan.resources.push({
            key: el.id,
            kind: 'PREPAID',
            price: el.value,
            period: this.getPeriod(el.duration),
            except: false,
            on: []
          });
        }
      });

      this.isLoading = true;
      api.plans.update(plan.uuid, plan)
        .then(() => {
          this.showSnackbarSuccess({ message: 'Plan edited successfully' });
        })
        .catch((err) => {
          const message = err.response?.data?.message ?? err.message ?? err;

          this.showSnackbarError({ message });
          console.error(err);
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
    getPeriod(duration) {
      switch (duration) {
        case 'P1M':
          return 3600 * 24 * 30;
        case 'P1Y':
          return 3600 * 24 * 30 * 12;
      }
    }
  },
  created() {
    this.changeIcon();
    this.isPlansLoading = true;
    this.$store.dispatch('plans/fetch', {
      sp_uuid: this.template.uuid,
      anonymously: false
    })
      .then(() => this.fetchError = '')
      .catch((err) => {
        this.fetchError = err.response?.data?.message ?? err.message ?? err;
        console.error(err);
      });

    api.post(`/sp/${this.template.uuid}/invoke`, {
      method: 'get_plans'
    })
      .then(({ meta }) => {
        this.changePlans(meta);
        this.changeAddons(meta);

        const footerButtons = document.querySelectorAll('.v-data-footer .v-btn__content');

        footerButtons.forEach((element) => {
          element.addEventListener('click', this.changeClose);
        });
        this.changeClose();
      })
      .catch((err) => {
        this.fetchError = err.response?.data?.message ?? err.message ?? err;
        console.error(err);
      })
      .finally(() => {
        this.isPlansLoading = false;
      });
  },
  computed: {
    nocloudPlans() {
      return this.$store.getters['plans/all'].filter(({ type }) => type === 'ovh');
    }
  },
  watch: {
    nocloudPlan(value) {
      console.log(value);
      const plan = this.nocloudPlans.find(({ uuid }) => uuid === value);

      plan.resources.forEach(({ key, price }) => {
        const addon = this.addons.find((el) => el.id === key);

        addon.value = price;
        addon.sell = true;
      });

      Object.entries(plan.products).forEach(([key, product]) => {
        const ovhPlan = this.plans.find((el) => el.id === key);
        const winKey = Object.keys(product.meta).find((el) => el.includes('windows'));

        ovhPlan.value = product.price;
        ovhPlan.sell = true;
        if (winKey) ovhPlan.windows.value = product.meta[winKey];
      });
    }
  }
}
</script>

<style>
.v-card .v-icon.group-icon {
  display: none;
  margin: 0 0 2px 4px;
  font-size: 18px;
  opacity: 0.5;
}

.v-data-table__expanded__content {
  background: var(--v-background-base);
}
</style>

<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-icon class="group-icon">mdi-format-list-group</v-icon>
    <v-row>
      <v-col cols="6">
        <v-expansion-panels>
          <v-expansion-panel>
            <v-expansion-panel-header color="background-light">
              Fee:
            </v-expansion-panel-header>
            <v-expansion-panel-content color="background-light">
              <plan-opensrs
                @changeFee="(data) => (fee = data)"
                @onValid="(data) => (isValid = data)"
              />
              <confirm-dialog
                text="This will apply the fee markup parameters to all prices"
                @confirm="setFee"
              >
                <v-btn class="mt-4">Set fee</v-btn>
              </confirm-dialog>
            </v-expansion-panel-content>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-col>
    </v-row>

    <nocloud-table
      item-key="id"
      :show-select="false"
      :show-group-by="true"
      :show-expand="true"
      :items="plans"
      :headers="headers"
      :expanded.sync="expanded"
      :loading="isPlansLoading"
      :footer-error="fetchError"
    >
      <template v-slot:expanded-item="{ headers, item }">
        <td :colspan="headers.length" style="padding: 0">
          <nocloud-table
            class="mx-8"
            style="background: var(--v-background-base) !important"
            :show-select="false"
            :items="item.prices"
            :headers="pricesHeaders"
          >
            <template v-slot:[`item.period`]="{ value }">
              {{ value }} months
            </template>
            <template v-slot:[`item.price`]="{ value }">
              {{ value }} {{ 'NCU' }}
            </template>
            <template v-slot:[`item.value`]="{ item }">
              <v-text-field dense style="width: 150px" v-model="item.value" />
            </template>
            <template v-slot:[`item.sell`]="{ item }">
              <v-switch v-model="item.sell" />
            </template>
          </nocloud-table>
        </td>
      </template>
    </nocloud-table>
    <v-btn class="mt-4" @click="editPlan">Save</v-btn>

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
import planOpensrs from '@/components/plan/opensrs/planOpensrs.vue';
import nocloudTable from "@/components/table.vue";
import confirmDialog from '@/components/confirmDialog.vue';
import snackbar from "@/mixins/snackbar.js";
import api from "@/api.js";

export default {
  name: 'sevices-provider-table',
  components: { nocloudTable, planOpensrs, confirmDialog },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({
    plans: [],
    expanded: [],
    headers: [
      { text: '', value: 'data-table-expand', groupable: false },
      { text: 'Brand', value: 'brand', sortable: false, class: 'groupable' },
      { text: 'Product', value: 'product', groupable: false },
      { text: 'Type', value: 'product_type', sortable: false, class: 'groupable' }
    ],
    pricesHeaders: [
      { text: 'Period', value: 'period' },
      { text: 'Price', value: 'price' },
      { text: 'New price', value: 'value' },
      { text: 'Sell', value: 'sell', width: 100 }
    ],

    isValid: true,
    isPlansLoading: false,
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
    setFee() {
      this.plans.forEach((plan) => {
        plan.prices.forEach((price, i, arr) => {
          const n = Math.pow(10, this.fee.precision);
          let percent = (this.fee?.default ?? 0) / 100 + 1;
          let round;

          switch (this.fee.round) {
            case 1:
              round = 'floor';
              break;
            case 2:
              round = 'round';
              break;
            case 3:
              round = 'ceil';
          }

          for (let range of this.fee.ranges) {
            if (price.value <= range.from) continue;
            if (price.value > range.to) continue;
            percent = range.factor / 100 + 1;
          }
          arr[i].value = Math[round](price.price * percent * n) / n;
        });
      });
    },
    editPlan() {
      if (!this.isValid) {
        this.showSnackbarError({ message: 'Fee is not valid' });
        return;
      }

      const newPlan = { ...this.template, fee: this.fee, resources: [], products: {} };

      this.plans.forEach((plan) => {
        plan.prices.forEach((el) => {
          if (el.sell) {
            const id = `${el.period} ${plan.id}`;

            newPlan.products[id] = {
              kind: 'PREPAID',
              title: plan.product,
              price: el.value,
              period: el.period * 30 * 24 * 3600,
              sorter: Object.keys(newPlan.products).length
            }
          }
        });
      });

      this.isLoading = true;
      api.plans.update(newPlan.uuid, newPlan).then(() => {
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
    getPrices(obj) {
      const result = [];

      Object.entries(obj).forEach(([period, value]) => {
        if (isFinite(+period)) result.push({ period, value, price: value });
      });
      return result;
    }
  },
  created() {
    this.isPlansLoading = true;
    this.$store.dispatch('servicesProviders/fetch')
      .then(({ pool }) => {
        const sp = pool.find(({ type }) => type === 'goget');

        return api.post(`/sp/${sp.uuid}/invoke`, { method: 'get_certificate' });
      })
      .then(({ meta: { cert } }) => {
        this.plans = cert.products;

        this.plans.forEach(({ prices }, i) => {
          this.plans[i].prices = this.getPrices(prices);
        });

        const footerButtons = document.querySelectorAll('.v-data-footer .v-btn__content');

        footerButtons.forEach((element) => {
          element.addEventListener('click', this.changeClose);
        });

        this.fetchError = '';
        this.changeIcon();
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
  watch: {
    plans() {
      Object.entries(this.template.products).forEach(([key, value]) => {
        const [period, id] = key.split(' ');
        const product = this.plans.find((el) => el.id === id);

        product[period].value = value.price;
        product[period].sell = true;
      });

      this.fee = this.template.fee;
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

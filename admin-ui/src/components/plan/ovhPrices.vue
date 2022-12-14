<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-icon class="group-icon">mdi-format-list-group</v-icon>
    <v-expansion-panels v-if="!isPlansLoading" :value="0">
      <v-expansion-panel>
        <v-expansion-panel-header color="background-light">
          Margin rules:
        </v-expansion-panel-header>
        <v-expansion-panel-content color="background-light">
          <plan-opensrs
            :fee="fee"
            :isEdit="true"
            @changeFee="(data) => (fee = data)"
            @onValid="(data) => (isValid = data)"
          />
          <confirm-dialog
            text="This will apply the rules markup parameters to all prices"
            @confirm="setFee"
          >
            <v-btn class="mt-4">Set rules</v-btn>
          </confirm-dialog>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>

    <v-tabs
      class="rounded-t-lg"
      v-model="tabsIndex"
      background-color="background-light"
    >
      <v-tab v-for="tab in tabs" :key="tab">{{ tab }}</v-tab>
    </v-tabs>

    <v-tabs-items
      v-model="tabsIndex"
      style="background: var(--v-background-light-base)"
      class="rounded-b-lg"
    >
      <v-tab-item v-for="tab in tabs" :key="tab">
        <v-progress-linear v-if="isPlansLoading" indeterminate class="pt-1" />

        <nocloud-table
          item-key="id"
          v-else-if="tab === 'Tariffs'"
          :show-expand="true"
          :show-select="false"
          :items="plans"
          :headers="headers"
          :expanded.sync="expanded"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.group`]="{ item }">
            <template v-if="mode === 'edit' && planId === item.id">
              <v-text-field dense class="d-inline-block mr-1" style="width: 200px" v-model="newGroupName" />
              <v-icon @click="editGroup(item.group)">mdi-content-save</v-icon>
              <v-icon @click="mode = 'none'">mdi-close</v-icon>
            </template>

            <template v-if="mode === 'create' && planId === item.id">
              <v-text-field dense class="d-inline-block mr-1" style="width: 200px" v-model="newGroupName" />
              <v-icon @click="createGroup(item)">mdi-content-save</v-icon>
              <v-icon @click="mode = 'none'">mdi-close</v-icon>
            </template>

            <template v-if="mode === 'none'">
              <v-select
                dense 
                class="d-inline-block"
                style="width: 200px"
                v-model="item.group"
                :items="groups"
                @change="item.name = getName(item)"
              />
              <v-icon @click="changeMode('create', item)">mdi-plus</v-icon>
              <v-icon @click="changeMode('edit', item)">mdi-pencil</v-icon>
              <v-icon v-if="groups.length > 1" @click="deleteGroup(item.group)">mdi-delete</v-icon>
            </template>

            <template v-else-if="planId !== item.id">{{ item.group }}</template>
          </template>
          <template v-slot:[`item.margin`]="{ item }">
            {{ getMargin(item) }}
          </template>
          <template v-slot:[`item.duration`]="{ value }">
            {{ (value === 'P1M') ? 'monthly' : 'yearly' }}
          </template>
          <template v-slot:[`item.price.value`]="{ item, value }">
            {{ value }} {{ 'NCU' || item.price.currencyCode }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field dense style="width: 150px" v-model="item.value" />
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
              <td><v-text-field dense style="width: 150px" v-model="item.windows.value" /></td>
              <td></td>
            </template>
            <template v-else>
              <td></td>
              <td :colspan="headers.length - 1">{{ $t('Windows is none') }}</td>
            </template>
          </template>
        </nocloud-table>

        <nocloud-table
          v-else-if="tab === 'Addons'"
          :show-select="false"
          :items="addons"
          :headers="addonsHeaders"
          :loading="isPlansLoading"
          :footer-error="fetchError"
        >
          <template v-slot:[`item.duration`]="{ value }">
            {{ (value === 'P1M') ? 'monthly' : 'yearly' }}
          </template>
          <template v-slot:[`item.price.value`]="{ item, value }">
            {{ value }} {{ 'NCU' || item.price.currencyCode }}
          </template>
          <template v-slot:[`item.value`]="{ item }">
            <v-text-field dense style="width: 150px" v-model="item.value" />
          </template>
          <template v-slot:[`item.sell`]="{ item }">
            <v-switch v-model="item.sell" />
          </template>
        </nocloud-table>
      </v-tab-item>
    </v-tabs-items>
    <v-btn class="mt-4 mr-2" color="secondary" :disabled="!isTestSuccess" @click="editPlan">Save</v-btn>
    <v-btn class="mt-4" :color="testButtonColor" @click="testConfig">Test</v-btn>

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
      { text: '', value: 'data-table-expand' },
      { text: 'Tariff', value: 'name', sortable: false, class: 'groupable' },
      { text: 'Group', value: 'group', sortable: false, class: 'groupable' },
      { text: 'Margin', value: 'margin', sortable: false, class: 'groupable' },
      { text: 'Payment', value: 'duration', sortable: false, class: 'groupable' },
      { text: 'Income price', value: 'price.value' },
      { text: 'Sale price', value: 'value' },
      { text: 'Sell', value: 'sell', sortable: false, class: 'groupable', width: 100 }
    ],

    addons: [],
    addonsHeaders: [
      { text: 'Addon', value: 'name', sortable: false, class: 'groupable' },
      { text: 'Payment', value: 'duration', sortable: false, class: 'groupable' },
      { text: 'Income price', value: 'price.value' },
      { text: 'Sale price', value: 'value' },
      { text: 'Sell', value: 'sell', sortable: false, class: 'groupable', width: 100 }
    ],

    fee: {},
    groups: [],
    newGroupName: '',
    planId: -1,
    mode: 'none',
    tabs: ['Tariffs', 'Addons'],
    tabsIndex: 0,
    isValid: true,
    isPlansLoading: false,
    isTestSuccess: false,
    testButtonColor: 'secondary',
    fetchError: ''
  }),
  methods: {
    changeIcon() {
      setTimeout(() => {
        const headers = document.querySelectorAll('.groupable');

        headers.forEach(({ firstElementChild, children }) => {
          if (!children[1]?.className.includes('group-icon')) {
            const icon = document.querySelector('.group-icon').cloneNode(true);

            firstElementChild.after(icon);
            icon.style = 'display: inline-flex';
          }
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
              price,
              duration,
              name: productName,
              group: productName.split(' ')[1],
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
    setFee() {
      const windows = [];

      this.plans.forEach((el) => {
        if (el.windows) windows.push(el.windows);
      });

      [this.plans, this.addons, windows].forEach((el) => {
        el.forEach((plan, i, arr) => {
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
            if (plan.price.value <= range.from) continue;
            if (plan.price.value > range.to) continue;
            percent = range.factor / 100 + 1;
          }
          arr[i].value = Math[round](plan.price.value * percent * n) / n;
        });
      });
    },
    editPlan() {
      const newPlan = { ...this.template, fee: this.fee, resources: [], products: {} };

      this.plans.forEach((el) => {
        if (el.sell) {
          const [,, cpu, ram, disk] = el.planCode.split('-');
          const meta = {};

          if (el.windows) meta[el.windows.code] = el.windows.value;
          newPlan.products[el.id] = {
            kind: 'PREPAID',
            title: el.name,
            price: el.value,
            period: this.getPeriod(el.duration),
            resources: { cpu: +cpu, ram: ram * 1024, disk: disk * 1024 },
            sorter: Object.keys(newPlan.products).length,
            meta
          }
        }
      });

      this.addons.forEach((el) => {
        if (el.sell) {
          newPlan.resources.push({
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
      api.plans.update(newPlan.uuid, newPlan)
        .then(() => {
          this.showSnackbarSuccess({ message: 'Price model edited successfully' });
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
    testConfig() {
      let message = '';
      if (!this.isValid) {
        message = 'Margin rules is not valid';
      }
      if (!this.plans.every(({ group }) => this.groups.includes(group))) {
        message = 'You must select a group for the tariff!';
      }

      if (message) {
        this.testButtonColor = 'secondary';
        this.isTestSuccess = false;
        this.showSnackbarError({ message });
        return;
      }

      this.testButtonColor = 'success';
      this.isTestSuccess = true;
    },
    getPeriod(duration) {
      switch (duration) {
        case 'P1M':
          return 3600 * 24 * 30;
        case 'P1Y':
          return 3600 * 24 * 30 * 12;
      }
    },
    getName({ name, group }) {
      const newGroup = `${group[0].toUpperCase()}${group.slice(1)}`;

      return `VPS ${newGroup} ${name.split(' ').at(-1)}`;
    },
    getMargin({ value, price }) {
      if (!this.fee.ranges) return 'none';
      const range = this.fee.ranges.find(({ from, to }) =>
        from <= price.value && to >= price.value
      );
      const n = Math.pow(10, this.fee.precision);
      let percent = range?.factor / 100 + 1;
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

      if (value === Math[round](price.value * percent * n) / n) return 'ranged';
      else percent = this.fee.default / 100 + 1;

      switch (value) {
        case price.value:
          return 'none';
        case Math[round](price.value * percent * n) / n:
          return 'fixed';
        default:
          return 'manual';
      }
    },
    editGroup(group) {
      const i = this.groups.indexOf(group);

      this.groups.splice(i, 1, this.newGroupName);
      this.plans.forEach((plan, index) => {
        if (plan.group !== group) return;
        this.plans[index].group = this.newGroupName;
        this.plans[index].name = this.getName(plan);
      });

      this.changeMode('none', { id: -1, group: '' });
    },
    createGroup(plan) {
      this.groups.push(this.newGroupName);
      plan.group = this.newGroupName;
      plan.name = this.getName(plan);

      this.changeMode('none', { id: -1, group: '' });
    },
    deleteGroup(group) {
      this.groups = this.groups.filter((el) => el !== group);
      this.plans.forEach((plan, i) => {
        if (plan.group !== group) return;
        this.plans[i].group = this.groups[0];
        this.plans[i].name = this.getName(plan);
      });
    },
    changeMode(mode, { id, group }) {
      this.mode = mode;
      this.planId = id;
      this.newGroupName = group;
    }
  },
  created() {
    this.isPlansLoading = true;
    this.$store.dispatch('servicesProviders/fetch')
      .then(({ pool }) => {
        const sp = pool.find(({ type }) => type === 'ovh');

        return api.post(`/sp/${sp.uuid}/invoke`, { method: 'get_plans' });
      })
      .then(({ meta }) => {
        this.changePlans(meta);
        this.changeAddons(meta);

        this.fetchError = '';
        this.changeIcon();
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
    tabsIndex() { this.changeIcon() },
    addons() {
      this.template.resources.forEach(({ key, price }) => {
        const addon = this.addons.find((el) => el.id === key);

        addon.value = price;
        addon.sell = true;
      });

      this.groups = [];
      this.plans.forEach(({ id, name }, i) => {
        const product = this.template.products[id];
        const winKey = Object.keys(product?.meta || {}).find((el) => el.includes('windows'));
        const group = product?.title.split(' ')[1] || name.split(' ')[1];

        if (product) {
          this.plans[i].name = product.title;
          this.plans[i].value = product.price;
          this.plans[i].group = group;
          this.plans[i].sell = true;

          if (winKey) this.plans[i].windows.value = product.meta[winKey];
        }
        if (!this.groups.includes(group)) this.groups.push(group);
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
  cursor: pointer;
}

.v-data-table__expanded__content {
  background: var(--v-background-base);
}
</style>

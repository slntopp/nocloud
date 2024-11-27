<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="Price model"
          :value="template.billingPlan.title"
        >
          <template v-slot:append>
            <v-icon @click="priceModelDialog = true">mdi-pencil</v-icon>
            <v-icon
              @click="
                $router.push({
                  name: 'Plan',
                  params: { planId: template.billingPlan.uuid },
                })
              "
              >mdi-login</v-icon
            >
          </template>
        </v-text-field>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Product name"
          :value="
            !isDynamicPlan
              ? template.billingPlan.products[template.product]?.title ||
                template.product
              : '-'
          "
          append-icon="mdi-pencil"
          @click:append="changeTariffDialog = true"
        />
      </v-col>
      <v-col>
        <date-picker
          edit-icon
          label="Date (create)"
          :value="formatSecondsToDateString(template.created, false, '-')"
          :placeholder="formatSecondsToDate(template.created, true)"
          :clearable="false"
          @input="
            emit('update', {
              key: 'created',
              value: formatDateToTimestamp($event),
            })
          "
        />
      </v-col>

      <v-col>
        <v-text-field
          label="Deleted date"
          readonly
          :value="formatSecondsToDate(template.deleted, true, '-')"
        />
      </v-col>

      <v-col v-if="!isMonitoringEmpty">
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="!isDynamicPlan ? dueDate : 'PAYG'"
          :append-icon="!isMonitoringEmpty ? 'mdi-pencil' : null"
          @click:append="changeDatesDialog = true"
        />
      </v-col>
    </v-row>
    <instances-panels title="Prices">
      <nocloud-table
        class="mb-5"
        :headers="billingHeaders"
        :items="billingItems"
        no-hide-uuid
        :show-select="false"
        hide-default-footer
      >
        <template v-slot:[`item.name`]="{ item }">
          <span v-html="item.name" />
          <v-chip v-if="item.isAddon" small class="ml-1">Addon</v-chip>
        </template>

        <template v-slot:[`item.price`]="{ item }">
          <div class="d-flex">
            <v-text-field
              class="mr-2"
              :suffix="defaultCurrency?.title"
              v-model="item.price"
              type="number"
              @input="updatePrice(item, false)"
              append-icon="mdi-pencil"
            />
            <v-text-field
              style="color: var(--v-primary-base)"
              class="ml-2"
              type="number"
              :suffix="accountCurrency?.title"
              v-model="item.accountPrice"
              @input="updatePrice(item, true)"
              append-icon="mdi-pencil"
            />
          </div>
        </template>
        <template v-slot:[`item.quantity`]="{ item }">
          {{ item.quantity?.toFixed(2) }} {{ item.unit }}
        </template>
        <template v-slot:[`item.total`]="{ item }">
          {{ totalPrices[item.name]?.toFixed(2) }}
          {{ defaultCurrency?.title }} /
          {{ totalAccountPrices[item.name]?.toFixed(2) }}
          {{ accountCurrency?.title }}
        </template>
        <template v-slot:body.append>
          <tr>
            <td></td>
            <td></td>
            <td>
              {{
                billingItems.find((i) => i.name === template.product)?.period
              }}
            </td>
            <td></td>
            <td>
              <v-dialog v-model="isAddonsDialog" max-width="60%">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn class="ml-2" color="primary" v-bind="attrs" v-on="on"
                    >addons</v-btn
                  >
                </template>
                <instance-change-addons
                  v-if="isAddonsDialog"
                  :instance="template"
                  :instance-addons="addons"
                  @update="
                    emit('update', {
                      key: 'addons',
                      value: $event,
                    })
                  "
                />
              </v-dialog>
            </td>
            <td>
              <v-chip color="primary" outlined>
                {{ totalPrice }}
                {{ defaultCurrency?.title }} / {{ totalAccountPrice }}
                {{ accountCurrency?.title }}
              </v-chip>
            </td>
          </tr>
        </template>
      </nocloud-table>
    </instances-panels>
    <change-ione-monitorings
      :template="template"
      :service="service"
      v-model="changeDatesDialog"
      @refresh="emit('refresh')"
      v-if="
        template.billingPlan.title.toLowerCase() !== 'payg' || isMonitoringEmpty
      "
    />
    <change-ione-tarrif
      v-if="availableTarrifs?.length > 0"
      v-model="changeTariffDialog"
      @refresh="emit('refresh')"
      :template="template"
      :service="service"
      :sp="sp"
      :available-tarrifs="availableTarrifs"
      :billing-plan="billingPlan"
    />
    <edit-price-model
      :account-rate="accountRate"
      :account-currency="accountCurrency"
      v-model="priceModelDialog"
      :template="template"
      @refresh="emit('refresh')"
      :service="service"
    />
  </div>
</template>

<script setup>
import {
  defineProps,
  ref,
  defineEmits,
  toRefs,
  computed,
  onMounted,
  watch,
} from "vue";
import {
  formatSecondsToDate,
  getBillingPeriod,
  formatSecondsToDateString,
  formatDateToTimestamp,
} from "@/functions";
import ChangeIoneMonitorings from "@/components/dialogs/changeMonitorings.vue";
import ChangeIoneTarrif from "@/components/dialogs/changeIoneTarrif.vue";
import NocloudTable from "@/components/table.vue";
import EditPriceModel from "@/components/dialogs/editPriceModel.vue";
import InstanceChangeAddons from "@/components/InstanceChangeAddons.vue";
import useInstancePrices from "@/hooks/useInstancePrices";
import { useStore } from "@/store";
import InstancesPanels from "../../ui/nocloudExpansionPanels.vue";
import DatePicker from "../../ui/datePicker.vue";

const props = defineProps(["template", "service", "sp", "account", "addons"]);
const emit = defineEmits(["refresh", "update"]);

const { template, service, sp, account, addons } = toRefs(props);

const store = useStore();
const { accountCurrency, toAccountPrice, accountRate, fromAccountPrice } =
  useInstancePrices(template.value, account.value);

const changeDatesDialog = ref(false);
const changeTariffDialog = ref(false);
const isAddonsDialog = ref(false);
const priceModelDialog = ref(false);
const billingItems = ref([]);
const billingHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Payment term", value: "kind" },
  { text: "Billing period", value: "period" },
  { text: "Price per unit", value: "price" },
  { text: "Unit quantity", value: "quantity" },
  { text: "Total price", value: "total" },
]);

onMounted(() => {
  billingItems.value = getBillingItems();
});

const dueDate = computed(() =>
  formatSecondsToDate(+template.value?.data?.next_payment_date, true)
);
const defaultCurrency = computed(() => store.getters["currencies/default"]);
const isMonitoringEmpty = computed(() => dueDate.value === "-");
const isDynamicPlan = computed(() => fullPlan.value?.kind === "DYNAMIC");
const fullPlan = computed(() => store.getters["plans/one"]);
const availableTarrifs = computed(() =>
  Object.keys(fullPlan.value?.products || {}).map((key) => ({
    title: key,
    resources: fullPlan.value.products[key].resources,
  }))
);

const billingPlan = computed(() => template.value.billingPlan);
const totalPrice = computed(() =>
  Object.keys(totalPrices.value || {})
    .reduce((acc, key) => acc + totalPrices.value[key], 0)
    .toFixed(2)
);

const totalPrices = computed(() => {
  const prices = {};

  billingItems.value.forEach(
    (i) => (prices[i.name] = i.price * i.quantity || 0)
  );

  return prices;
});

const totalAccountPrices = computed(() => {
  const prices = {};

  billingItems.value.forEach(
    (i) => (prices[i.name] = i.accountPrice * i.quantity)
  );

  return prices;
});

const totalAccountPrice = computed(() => toAccountPrice(totalPrice.value));
const getBillingItems = () => {
  const items = [];

  if (billingPlan.value.products[template.value.product]) {
    const { price, kind, period } =
      billingPlan.value.products[template.value.product];
    items.push({
      name: template.value.product,
      price,
      accountPrice: toAccountPrice(price),
      path: `billingPlan.products.${template.value.product}.price`,
      quantity: 1,
      unit: "pcs",
      kind,
      period,
    });
  }

  const driveType = template.value.resources.drive_type?.toLowerCase();

  if (driveType) {
    const driveIndex = billingPlan.value.resources.findIndex(
      (r) => r.key === `drive_${driveType}`
    );
    const drive = billingPlan.value.resources[driveIndex];
    items.push({
      name: driveType,
      price: drive?.price,
      accountPrice: toAccountPrice(drive.price),
      path: `billingPlan.resources.${driveIndex}.price`,
      kind: drive?.kind,
      quantity: template.value.resources.drive_size / 1024,
      unit: "GB",
      period: drive?.period,
    });
  }

  addons.value.forEach((addon, index) => {
    const { title, periods } = addon;
    const { period, kind } = billingPlan.value.products[template.value.product];
    items.push({
      name: title,
      price: periods[period],
      path: `${index}.periods.${period}`,
      unit: "pcs",
      accountPrice: toAccountPrice(periods[period]),
      quantity: 1,
      isAddon: true,
      kind,
      period,
    });
  });

  return items.map((i) => {
    i.period = getBillingPeriod(i.period);
    return i;
  });
};

const updatePrice = (item, isAccount) => {
  if (isAccount) {
    emit("update", {
      key: item.path,
      value: fromAccountPrice(item.accountPrice),
      type: item.isAddon ? "addons" : "template",
    });
    billingItems.value = billingItems.value.map((p) => {
      if (p.path === item.path) {
        p.price = fromAccountPrice(item.accountPrice);
      }
      return p;
    });
  } else {
    emit("update", {
      key: item.path,
      value: item.price,
      type: item.isAddon ? "addons" : "template",
    });
    billingItems.value = billingItems.value.map((p) => {
      if (p.path === item.path) {
        p.accountPrice = toAccountPrice(item.price);
      }
      return p;
    });
  }
};

watch(accountRate, () => {
  billingItems.value = billingItems.value.map((i) => {
    i.accountPrice = toAccountPrice(i.price);
    return i;
  });
});
</script>

<style lang="scss">
.ione-billing {
  .v-expansion-panel-content__wrap {
    padding: 0px !important;
  }
}
</style>

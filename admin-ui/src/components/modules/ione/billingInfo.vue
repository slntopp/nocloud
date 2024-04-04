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
        <v-text-field
          readonly
          label="Date (create)"
          :value="formatSecondsToDate(template.created, true)"
        />
      </v-col>
      <v-col>
        <!-- <v-text-field
          label="Date (create)"
          type="datetime-local"
          :value="formatSecondsToDateTime(template.created, true)"
        /> -->

        <date-picker
          label="Date (create)"
          :value="formatSecondsToDateString(template.created, false, '-')"
          :placeholder="formatSecondsToDate(template.created, true)"
          @input="
            emit('update', {
              key: 'created',
              value: formatDateToTimestamp($event),
            })
          "
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
        <template v-slot:[`item.price`]="{ item }">
          <div class="d-flex">
            <v-text-field
              class="mr-2"
              :suffix="defaultCurrency"
              v-model="item.price"
              type="number"
              @input="updatePrice(item, false)"
              append-icon="mdi-pencil"
            />
            <v-text-field
              style="color: var(--v-primary-base)"
              class="ml-2"
              type="number"
              :suffix="accountCurrency"
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
          {{ defaultCurrency }} /
          {{ totalAccountPrices[item.name]?.toFixed(2) }}
          {{ accountCurrency }}
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
            <td></td>
            <td>
              <v-chip color="primary" outlined>
                {{ totalPrice }}
                {{ defaultCurrency }} / {{ totalAccountPrice }}
                {{ accountCurrency }}
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
      :plans="filteredPlans"
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
import useInstancePrices from "@/hooks/useInstancePrices";
import { useStore } from "@/store";
import InstancesPanels from "../../ui/nocloudExpansionPanels.vue";
import DatePicker from "../../ui/datePicker.vue";

const props = defineProps(["template", "plans", "service", "sp", "account"]);
const emit = defineEmits(["refresh", "update"]);

const { template, service, sp, plans, account } = toRefs(props);

const store = useStore();
const { accountCurrency, toAccountPrice, accountRate, fromAccountPrice } =
  useInstancePrices(template.value, account.value);

const changeDatesDialog = ref(false);
const changeTariffDialog = ref(false);
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
const fullPlan = computed(() =>
  plans.value.find((p) => p.uuid === template.value.billingPlan.uuid)
);
const availableTarrifs = computed(() =>
  Object.keys(fullPlan.value?.products || {}).map((key) => ({
    title: key,
    resources: fullPlan.value.products[key].resources,
  }))
);
const filteredPlans = computed(() =>
  plans.value.filter((p) => p.type === "ione")
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

  const gbAddons = ["ram"];
  Object.keys(template.value.resources).forEach((resourceKey) => {
    let quantity = template.value.resources[resourceKey];
    if (!quantity) {
      return;
    }

    if (gbAddons.includes(resourceKey)) {
      quantity = quantity / 1024;
    }

    const addonIndex = billingPlan.value.resources.findIndex(
      (r) => r.key === resourceKey
    );
    const addon = billingPlan.value.resources[addonIndex];
    if (addon) {
      items.push({
        name: resourceKey,
        price: addon.price,
        accountPrice: toAccountPrice(addon.price),
        kind: addon.kind,
        period: addon.period,
        quantity,
        unit: "pcs",
        path: `billingPlan.resources.${addonIndex}.price`,
      });
    }
  });

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
    });
    billingItems.value = billingItems.value.map((p) => {
      if (p.path === item.path) {
        p.price = fromAccountPrice(item.accountPrice);
      }
      return p;
    });
  } else {
    emit("update", { key: item.path, value: item.price });
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

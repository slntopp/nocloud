<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="Price model"
          :model-value="template.billingPlan.title"
        >
          <template v-slot:append-inner>
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
          :model-value="
            !isDynamicPlan
              ? template.billingPlan.products[template.product]?.title ||
                template.product
              : '-'
          "
          append-inner-icon="mdi-pencil"
          @click:append-inner="changeTariffDialog = true"
        />
      </v-col>
      <v-col>
        <date-picker
          label="Date (create)"
          :model-value="timestampToDateTimeLocal(template?.created)"
          :clearable="false"
          @update:model-value="
            emit('update', {
              key: 'created',
              value: formatDateToTimestamp($event),
            })
          "
        />
      </v-col>

      <v-col>
        <date-picker
          label="Deleted date"
          :model-value="timestampToDateTimeLocal(template?.deleted)"
          :clearable="false"
          @update:model-value="
            emit('update', {
              key: 'deleted',
              value: formatDateToTimestamp($event),
            })
          "
        />
      </v-col>

      <v-col v-if="!isMonitoringEmpty">
        <v-text-field
          readonly
          label="Due to date/next payment"
          :model-value="!isDynamicPlan ? dueDate : 'PAYG'"
          :append-inner-icon="!isMonitoringEmpty ? 'mdi-pencil' : null"
          @click="changeDatesDialog = true"
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
          <template v-if="item.isAddon">
            <v-chip size="small" class="ml-1">Addon</v-chip>
            <confirm-dialog @confirm="removeAddon(item)">
              <v-btn icon variant="text" size="small" class="ml-1">
                <v-icon size="small">mdi-delete</v-icon>
              </v-btn>
            </confirm-dialog>
          </template>
        </template>

        <template v-slot:[`item.price`]="{ item }">
          <div class="d-flex">
            <v-text-field
              class="mr-2"
              :suffix="defaultCurrency?.code"
              v-model="item.price"
              type="number"
              @update:model-value="updatePrice(item, false)"
              append-inner-icon="mdi-pencil"
            />
            <v-text-field
              style="color: rgb(var(--v-theme-primary))"
              class="ml-2"
              type="number"
              :suffix="accountCurrency?.code"
              v-model="item.accountPrice"
              @update:model-value="updatePrice(item, true)"
              append-inner-icon="mdi-pencil"
            />
          </div>
        </template>
        <template v-slot:[`item.quantity`]="{ item }">
          {{ item.quantity?.toFixed(2) }} {{ item.unit }}
        </template>
        <template v-slot:[`item.total`]="{ item }">
          {{ formatPrice(totalPrices[item.name], defaultCurrency) }}
          {{ defaultCurrency?.code }} /
          {{ formatPrice(totalAccountPrices[item.name], accountCurrency) }}
          {{ accountCurrency?.code }}
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
                <template v-slot:activator="{ props }">
                  <v-btn class="ml-2" color="primary" v-bind="props"
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
              <v-chip color="primary" variant="outlined">
                {{ totalPrice }}
                {{ defaultCurrency?.code }} / {{ totalAccountPrice }}
                {{ accountCurrency?.code }}
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
  ref,
  toRefs,
  computed,
  onMounted,
  watch,
} from "vue";
import {
  formatSecondsToDate,
  getBillingPeriod,
  formatDateToTimestamp,
} from "@/functions";
import ChangeIoneMonitorings from "@/components/dialogs/changeMonitorings.vue";
import ChangeIoneTarrif from "@/components/dialogs/changeIoneTarrif.vue";
import NocloudTable from "@/components/table.vue";
import EditPriceModel from "@/components/dialogs/editPriceModel.vue";
import InstanceChangeAddons from "@/components/InstanceChangeAddons.vue";
import ConfirmDialog from "@/components/confirmDialog.vue";
import useInstancePrices from "@/hooks/useInstancePrices";
import { useStore } from "@/store";
import InstancesPanels from "../../ui/nocloudExpansionPanels.vue";
import DatePicker from "../../ui/dateTimePicker.vue";
import { formatPrice, timestampToDateTimeLocal } from "../../../functions";

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
  Object.keys(totalPrices.value || {}).reduce(
    (acc, key) => acc + totalPrices.value[key],
    0
  )
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

  const ipsKey = "ips_public";
  const ipsIndex = billingPlan.value.resources.findIndex(
    (r) => r.key === ipsKey
  );
  if (ipsIndex != -1) {
    const ips = billingPlan.value.resources[ipsIndex];
    items.push({
      name: ips.title,
      price: ips.price,
      accountPrice: toAccountPrice(ips.price),
      path: `billingPlan.resources.${ipsIndex}.price`,
      kind: ips.kind,
      quantity: template.value.resources[ipsKey],
      unit: "pcs",
      period: ips.period,
    });
  }

  addons.value.forEach((addon, index) => {
    // the instance may have dropped the addon since it was fetched, the index
    // still has to point at the fetched array because item.path uses it
    if (!template.value.addons?.includes(addon.uuid)) {
      return;
    }

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
      uuid: addon.uuid,
      kind,
      period,
    });
  });

  return items.map((i) => {
    i.period = getBillingPeriod(i.period);
    return i;
  });
};

const removeAddon = (item) => {
  emit("update", {
    key: "addons",
    value: template.value.addons.filter((uuid) => uuid !== item.uuid),
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

// the addons dialog and the delete button both only change the uuid list
watch(
  () => template.value.addons,
  () => {
    billingItems.value = getBillingItems();
  }
);

watch(accountRate, () => {
  billingItems.value = billingItems.value.map((i) => {
    i.accountPrice = toAccountPrice(i.price);
    return i;
  });
});
</script>

<style lang="scss">
.ione-billing {
  .v-expansion-panel-text__wrapper {
    padding: 0px !important;
  }
}
</style>

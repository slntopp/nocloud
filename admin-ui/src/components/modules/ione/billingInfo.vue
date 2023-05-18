<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="price model"
          :value="template.billingPlan.title"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Tarif (product plan)"
          :value="template.product"
          append-icon="mdi-pencil"
          @click:append="changeTarrifDialog = true"
        />
      </v-col>
      <v-col>
        <v-text-field readonly label="Price instance total" :value="price" />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Date (create)"
          :value="template.data.creation"
        />
      </v-col>

      <v-col
        v-if="
          template.billingPlan.title.toLowerCase() !== 'payg' ||
          isMonitoringsEmpty
        "
      >
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="date"
          :append-icon="!isMonitoringsEmpty ? 'mdi-pencil' : null"
          @click:append="changeDatesDialog = true"
        />
      </v-col>
      ></v-row
    >
    <nocloud-table
      class="mb-5"
      :headers="billingHeaders"
      :items="billingItems"
      no-hide-uuid
      table-name="ione-billing"
      :show-select="false"
      hide-default-footer
    >
      <template v-slot:[`item.price`]="{ item }">
        <v-text-field v-model="item.price" />
      </template>
      <template v-slot:[`item.total`]="{ item }">
        {{ totalPrices[item.name] }}
      </template>
      <template v-slot:body.append>
        <tr>
          <td>Total instance price</td>
          <td></td>
          <td></td>
          <td></td>
          <td></td>
          <td>
            {{ billingItems.find((i) => i.name == template.product)?.period }}
          </td>
          <td>{{ totalPrice }}</td>
        </tr>
      </template>
    </nocloud-table>
    <change-ione-monitorings
      :template="template"
      :service="service"
      :value="changeDatesDialog"
      @input="changeDatesDialog = $event"
      @refresh="emit('refresh')"
      v-if="
        template.billingPlan.title.toLowerCase() !== 'payg' ||
        isMonitoringsEmpty
      "
    />
    <change-ione-tarrif
      v-if="availableTarrifs?.length > 0"
      :value="changeTarrifDialog"
      @input="changeTarrifDialog = $event"
      @refresh="emit('refresh')"
      :template="template"
      :service="service"
      :sp="sp"
      :available-tarrifs="availableTarrifs"
      :billing-plan="billingPlan"
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
} from "vue";
import { formatSecondsToDate, getFullDate } from "@/functions";
import ChangeIoneMonitorings from "@/components/dialogs/changeIoneMonitorings.vue";
import ChangeIoneTarrif from "@/components/dialogs/changeIoneTarrif.vue";
import NocloudTable from "@/components/table.vue";

const props = defineProps(["template", "plans", "service", "sp"]);
const emit = defineEmits(["refresh"]);

const { template, service, sp } = toRefs(props);
const changeDatesDialog = ref(false);
const changeTarrifDialog = ref(false);
const price = ref(0);
const billingItems = ref([]);
const billingHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Unit name", value: "unit" },
  { text: "Price per unit", value: "price" },
  { text: "Unit quantity", value: "quantity" },
  { text: "Payment term", value: "kind" },
  { text: "Billing period", value: "period" },
  { text: "Total price", value: "total" },
]);

const date = computed(() =>
  formatSecondsToDate(template.value?.data?.last_monitoring)
);
const isMonitoringsEmpty = computed(() => date.value === "-");
const availableTarrifs = computed(() =>
  Object.keys(billingPlan.value?.products || {}).map((key) => ({
    title: key,
    resources: billingPlan.value.products[key].resources,
  }))
);
const billingPlan = computed(() => template.value.billingPlan);
const totalPrice = computed(() =>
  Object.keys(totalPrices.value || {}).reduce(
    (acc, key) => acc + totalPrices.value[key],
    0
  )
);

onMounted(() => {
  billingItems.value = getBillingItems();
  price.value = billingItems.value.reduce(
    (acc, bi) => acc + +(bi.price || 0),
    0
  );
});

const totalPrices = computed(() => {
  const prices = {};

  billingItems.value.forEach(
    (i) => (prices[i.name] = i.price * i.quantity || 0)
  );

  return prices;
});

const getBillingItems = () => {
  const items = [];

  items.push({
    name: template.value.product,
    price: billingPlan.value.products[template.value.product]?.price,
    quantity: 1,
    unit: "pcs",
    kind: billingPlan.value.products[template.value.product]?.kind,
    period: billingPlan.value.products[template.value.product]?.period,
  });

  Object.keys(template.value.resources).forEach((resourceKey) => {
    const quantity = template.value.resources[resourceKey];
    if (!quantity) {
      return;
    }
    const addon = billingPlan.value.resources.find(
      (r) => r.key === resourceKey
    );
    if (addon) {
      items.push({
        name: resourceKey,
        price: addon.price,
        kind: addon.kind,
        period: addon.period,
        quantity,
        unit: "pcs",
      });
    }
  });

  const driveType = template.value.resources.drive_type?.toLowerCase();
  if (driveType) {
    items.push({
      name: driveType,
      price: billingPlan.value.resources[`drive_${driveType}`]?.price,
      kind: billingPlan.value.resources[`drive_${driveType}`]?.kind,
      quantity: template.value.resources.drive_size / 1024,
      unit: "GB",
      period: billingPlan.value.resources[`drive_${driveType}`]?.period,
    });
  }

  return items.map((i) => {
    const fullPeriod = i.period && getFullDate(i.period);
    if (fullPeriod) {
      i.period = Object.keys(fullPeriod)
        .filter((key) => +fullPeriod[key])
        .map((key) => `${fullPeriod[key]} (${key})`)
        .join(", ");
    }

    return i;
  });
};
</script>

<style scoped></style>

<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="price model"
          :value="template.billingPlan.title"
          @click:append="priceModelDialog = true"
          append-icon="mdi-pencil"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Tarif (product plan)"
          :value="billingPlan.products[template.product].title"
          @click:append="priceModelDialog = true"
          append-icon="mdi-pencil"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Price instance total"
          :suffix="defaultCurrency"
          :value="price"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Account price instance total"
          :value="accountPrice"
          :suffix="accountCurrency"
        />
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
    </v-row>

    <v-expansion-panels>
      <v-expansion-panel>
        <v-expansion-panel-header color="background-light"
          >Prices</v-expansion-panel-header
        >
        <v-expansion-panel-content
          class="ione-billing"
          color="background-light"
        >
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
                  v-model="item.price"
                  :suffix="defaultCurrency"
                  @input="updatePrice(item, false)"
                  append-icon="mdi-pencil"
                />
                <v-text-field
                  class="ml-2"
                  :suffix="accountCurrency"
                  style="color: #c921c9"
                  v-model="item.accountPrice"
                  @input="updatePrice(item, true)"
                  append-icon="mdi-pencil"
                />
              </div>
            </template>
            <template v-slot:body.append>
              <tr>
                <td></td>
                <td />
                <td>
                  {{
                    billingItems.find((i) => i.name === template.product)
                      ?.period
                  }}
                </td>
                <td>
                  <div class="d-flex justify-end mr-4">
                    {{ [totalPrice, defaultCurrency].join(" ") }}
                    /
                    {{ [totalAccountPrice, accountCurrency].join(" ") }}
                  </div>
                </td>
              </tr>
            </template>
          </nocloud-table>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>

    <change-monitorings
      :template="template"
      :service="service"
      v-model="changeDatesDialog"
      @refresh="emit('refresh')"
    />
    <edit-price-model
      :account-rate="accountRate"
      :account-currency="accountCurrency"
      v-model="priceModelDialog"
      :template="template"
      :plans="filtredPlans"
      @refresh="emit('refresh')"
      :service="service"
    />
  </div>
</template>

<script setup>
import { computed, defineProps, toRefs, ref, watch, onMounted } from "vue";
import { formatSecondsToDate, getBillingPeriod } from "@/functions";
import ChangeMonitorings from "@/components/dialogs/changeMonitorings.vue";
import EditPriceModel from "@/components/dialogs/editPriceModel.vue";
import useAccountConverter from "@/hooks/useAccountConverter";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";

const props = defineProps(["template", "plans", "service", "sp"]);
const emit = defineEmits(["refresh"]);

const { template, plans, service } = toRefs(props);

const store = useStore();
const {
  fetchAccountRate,
  accountCurrency,
  toAccountPrice,
  accountRate,
  fromAccountPrice,
} = useAccountConverter(template.value);

const changeDatesDialog = ref(false);
const priceModelDialog = ref(false);
const accountPrice = ref(0);
const billingHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Payment term", value: "kind" },
  { text: "Billing period", value: "period" },
  { text: "Price", value: "price" },
]);
const billingItems = ref([]);

const date = computed(() =>
  formatSecondsToDate(template.value?.data?.last_monitoring)
);
const isMonitoringsEmpty = computed(() => date.value === "-");

const price = computed(() => {
  return template.value.billingPlan.products[template.value.product]?.price;
});

const filtredPlans = computed(() =>
  plans.value.filter((p) => p.type === "virtual")
);

const totalAccountPrice = computed(() => {
  return toAccountPrice(totalPrice.value);
});
const totalPrice = computed(() => {
  return billingItems.value.reduce((acc, i) => acc + +i.price, 0);
});

const billingPlan = computed(() => template.value.billingPlan);

const defaultCurrency = computed(() => store.getters["currencies/default"]);

onMounted(() => {
  fetchAccountRate();
  billingItems.value = getBillingItems();
});

watch(accountRate, () => {
  accountPrice.value = totalAccountPrice.value;
  billingItems.value = billingItems.value.map((i) => {
    i.accountPrice = toAccountPrice(i.price);
    return i;
  });
});

const getBillingItems = () => {
  const items = [];

  items.push({
    name: template.value.product,
    price: billingPlan.value.products[template.value.product]?.price,
    path: `billingPlan.products.${template.value.product}.price`,
    kind: billingPlan.value.products[template.value.product]?.kind,
    period: billingPlan.value.products[template.value.product]?.period,
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
</script>

<style scoped></style>

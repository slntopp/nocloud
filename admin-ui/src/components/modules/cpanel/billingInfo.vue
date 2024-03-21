<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="price model"
          :value="template.billingPlan.title"
          append-icon="mdi-pencil"
          @click:append="priceModelDialog = true"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Product name"
          :value="
            template.billingPlan.products[template.resources.plan]?.title ||
            template.resources.plan
          "
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Date (create)"
          :value="formatSecondsToDate(template.created, true)"
        />
      </v-col>

      <v-col
        v-if="
          template.billingPlan.title.toLowerCase() !== 'payg' ||
          isMonitoringEmpty
        "
      >
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="dueDate"
          :append-icon="!isMonitoringEmpty ? 'mdi-pencil' : null"
          @click:append="changeDatesDialog = true"
        />
      </v-col>
    </v-row>
    <instances-prices-panels title="Prices">
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
              v-model="item.price"
              :suffix="defaultCurrency"
              @input="updatePrice(item, false)"
              append-icon="mdi-pencil"
            />
            <v-text-field
              class="ml-2"
              :suffix="accountCurrency"
              style="color: var(--v-primary-base)"
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
                billingItems.find((i) => i.name === template.product)?.period
              }}
            </td>
            <td>
              <div class="d-flex justify-end mr-4">
                <v-chip color="primary" outlined>
                  {{ [totalPrice, defaultCurrency].join(" ") }}
                  /
                  {{ [totalAccountPrice, accountCurrency].join(" ") }}
                </v-chip>
              </div>
            </td>
          </tr>
        </template>
      </nocloud-table>
    </instances-prices-panels>
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
import { defineProps, toRefs, computed, ref, onMounted, watch } from "vue";
import { formatSecondsToDate, getBillingPeriod } from "@/functions";
import NocloudTable from "@/components/table.vue";
import InstancesPricesPanels from "@/components/ui/nocloudExpansionPanels.vue";
import { useStore } from "@/store";
import useInstancePrices from "@/hooks/useInstancePrices";
import EditPriceModel from "@/components/dialogs/editPriceModel.vue";

const props = defineProps(["template", "plans", "service", "sp"]);
const emit = defineEmits(["refresh"]);

const { template, plans } = toRefs(props);

const store = useStore();
const { accountCurrency, toAccountPrice, accountRate, fromAccountPrice } =
  useInstancePrices(template.value);

const billingHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Payment term", value: "kind" },
  { text: "Billing period", value: "period" },
  { text: "Price", value: "price" },
]);
const billingItems = ref([]);
const priceModelDialog = ref(false);

const filtredPlans = computed(() =>
  plans.value.filter((p) => p.type === "cpanel")
);
const totalPrice = computed(() => {
  return billingItems.value.reduce((acc, i) => acc + +i.price, 0)?.toFixed(2);
});

const totalAccountPrice = computed(() => {
  return billingItems.value
    .reduce((acc, i) => acc + +i.accountPrice, 0)
    ?.toFixed(2);
});

const billingPlan = computed(() => template.value.billingPlan);

const defaultCurrency = computed(() => store.getters["currencies/default"]);

onMounted(() => {
  billingItems.value = getBillingItems();
});

watch(accountRate, () => {
  billingItems.value = billingItems.value.map((i) => {
    i.accountPrice = toAccountPrice(i.price);
    return i;
  });
});

const dueDate = computed(() =>
  formatSecondsToDate(+template.value?.data?.next_payment_date, true)
);
const isMonitoringEmpty = computed(() => dueDate.value === "-");

const getBillingItems = () => {
  const items = [];
  const product = billingPlan.value.products[template.value.product];
  items.push({
    name: product.title,
    price: product?.price,
    path: `billingPlan.products.${template.value.product}.price`,
    kind: product?.kind,
    period: getBillingPeriod(product?.period),
    accountPrice: toAccountPrice(product?.price),
  });

  return items.map((i) => {
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

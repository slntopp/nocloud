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
          label="Product name"
          :value="billingPlan.products[template.product].title"
          @click:append="priceModelDialog = true"
          append-icon="mdi-pencil"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Date (create)"
          :value="template.data.creation"
        />
      </v-col>

      <v-col>
        <v-text-field
          readonly
          label="Start date"
          :value="template.data.start || '-'"
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
        <v-expansion-panel-header color="background-light">
          <div>
            <span style="color: var(--v-primary-base)" class="text-h6"
              >Prices
            </span>
            <v-dialog max-width="60%">
              <template v-slot:activator="{ on, attrs }">
                <v-btn class="ml-2" color="primary" v-bind="attrs" v-on="on"
                  >Change addons</v-btn
                >
              </template>
              <v-card>
                <nocloud-table
                  :items="addons"
                  :headers="addonsHeaders"
                  no-hide-uuid
                  :show-select="false"
                  hide-default-footer
                >
                  <template v-slot:[`item.name`]="{ item }">
                    <span v-html="item.name" />
                  </template>
                  <template v-slot:[`item.enabled`]="{ item }">
                    <v-switch
                      @change="changeAddons"
                      v-model="item.enabled"
                    ></v-switch>
                  </template>
                </nocloud-table>
              </v-card>
            </v-dialog>
          </div>
          <template v-slot:actions>
            <v-icon color="primary" x-large> $expand </v-icon>
          </template>
        </v-expansion-panel-header>
        <v-expansion-panel-content color="background-light">
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
                    billingItems.find((i) => i.name === template.product)
                      ?.period
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
import useInstancePrices from "@/hooks/useInstancePrices";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";

const props = defineProps(["template", "plans", "service", "sp"]);
const emit = defineEmits(["refresh"]);

const { template, plans, service } = toRefs(props);

const store = useStore();
const { accountCurrency, toAccountPrice, accountRate, fromAccountPrice } =
  useInstancePrices(template.value);

const changeDatesDialog = ref(false);
const priceModelDialog = ref(false);

function getBillingHeaders() {
  return [
    { text: "Name", value: "name" },
    { text: "Payment term", value: "kind" },
    { text: "Billing period", value: "period" },
    { text: "Price", value: "price" },
  ];
}

const billingHeaders = ref(getBillingHeaders());
const addonsHeaders = ref([
  ...getBillingHeaders(),
  { text: "Enabled", value: "enabled" },
]);
const billingItems = ref([]);

const date = computed(() =>
  formatSecondsToDate(template.value?.data?.next_payment_date)
);
const isMonitoringsEmpty = computed(() => date.value === "-");

const filtredPlans = computed(() =>
  plans.value.filter((p) => p.type === "empty")
);

const totalPrice = computed(() => {
  return billingItems.value.reduce((acc, i) => acc + +i.price, 0);
});

const totalAccountPrice = computed(() => {
  return billingItems.value.reduce((acc, i) => acc + +i.accountPrice, 0);
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

const addons = computed(() => {
  return billingPlan.value.resources.map(
    ({ price, title, kind, period, key }, index) => ({
      name: title,
      price,
      enabled: !!template.value.config?.addons?.find((a) => a === key),
      path: `billingPlan.resources.${index}.price`,
      kind,
      key,
      period,
      accountPrice: toAccountPrice(price),
    })
  );
});

const getBillingItems = () => {
  const items = [];
  const product = billingPlan.value.products[template.value.product];
  items.push({
    name: template.value.product,
    price: product?.price,
    path: `billingPlan.products.${template.value.product}.price`,
    kind: product?.kind,
    period: product?.period,
    accountPrice: toAccountPrice(product?.price),
  });

  items.push(...addons.value.filter((a) => a.enabled));

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

const changeAddons = () => {
  emit("update", {
    key: "config.addons",
    value: addons.value.filter((a) => a.enabled).map((a) => a.key),
  });
};
</script>

<style scoped></style>

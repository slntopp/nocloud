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
            <v-icon
              v-if="isPriceModelCanBeChange"
              @click="priceModelDialog = true"
              >mdi-pencil</v-icon
            >
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
          :value="tarrif.title"
          :append-icon="isPriceModelCanBeChange ? 'mdi-pencil' : undefined"
          @click:append="
            isPriceModelCanBeChange ? (priceModelDialog = true) : undefined
          "
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

      <v-col>
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="dueDate"
        />
      </v-col>
    </v-row>

    <instances-panels title="Prices">
      <nocloud-table
        hide-default-footer
        sort-by="index"
        item-key="key"
        :show-select="false"
        :headers="pricesHeaders"
        :items="pricesItems"
      >
        <template v-slot:[`item.prices`]="{ item }">
          <div class="d-flex">
            <v-text-field
              class="mr-2"
              v-model="item.price"
              @change="onUpdatePrice(item, false)"
              :suffix="defaultCurrency?.code"
              type="number"
              append-icon="mdi-pencil"
            ></v-text-field>
            <v-text-field
              class="ml-2"
              style="color: var(--v-primary-base)"
              v-model="item.accountPrice"
              @change="onUpdatePrice(item, true)"
              :suffix="accountCurrency?.code"
              type="number"
              append-icon="mdi-pencil"
            ></v-text-field>
          </div>
        </template>
        <template v-slot:[`item.basePrice`]="{ item }">
          <span v-if="+item.basePrice"> {{ item.basePrice }} PLN </span>
        </template>
        <template v-slot:[`item.title`]="{ item }">
          <span v-html="item.title" />
          <v-chip v-if="item.isAddon" small class="ml-1">Addon</v-chip>
        </template>
        <template v-slot:body.append>
          <tr>
            <td></td>
            <td></td>
            <td>
              {{ getBillingPeriod(tarrif.period) }}
            </td>
            <td>
              {{ [totalBasePrice, "PLN"].join(" ") }}
            </td>
            <td>
              <div class="d-flex justify-end">
                <v-dialog v-model="isAddonsDialog" max-width="60%">
                  <template v-slot:activator="{ on, attrs }">
                    <v-btn class="mr-4" color="primary" v-bind="attrs" v-on="on"
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
                <v-chip outlined color="primary" class="mr-4">
                  {{
                    [
                      formatPrice(totalNewPrice, defaultCurrency),
                      defaultCurrency?.code,
                    ].join(" ")
                  }}
                  /
                  {{
                    [
                      formatPrice(accountTotalNewPrice, accountCurrency),
                      accountCurrency?.code,
                    ].join(" ")
                  }}
                </v-chip>
              </div>
            </td>
          </tr>
        </template>
      </nocloud-table>
    </instances-panels>

    <edit-price-model
      v-if="isPriceModelCanBeChange"
      @refresh="emit('refresh')"
      :template="template"
      :account-currency="accountCurrency"
      :account-rate="accountRate"
      :service="service"
      v-model="priceModelDialog"
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
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";
import EditPriceModel from "@/components/dialogs/editPriceModel.vue";
import usePlnRate from "@/hooks/usePlnRate";
import {
  getBillingPeriod,
  formatDateToTimestamp,
  formatSecondsToDate,
  formatSecondsToDateString,
} from "@/functions";
import useInstancePrices from "@/hooks/useInstancePrices";
import InstancesPanels from "@/components/ui/nocloudExpansionPanels.vue";
import DatePicker from "../../ui/datePicker.vue";
import InstanceChangeAddons from "@/components/InstanceChangeAddons.vue";
import { formatPrice } from "../../../functions";

const props = defineProps(["template", "account", "addons"]);
const emit = defineEmits(["refresh", "update"]);

const { template, account, addons } = toRefs(props);

const store = useStore();
const plnRate = usePlnRate();
const { toAccountPrice, fromAccountPrice, accountCurrency, accountRate } =
  useInstancePrices(template.value, account.value);

const pricesItems = ref([]);
const pricesHeaders = ref([
  { text: "Name", value: "title" },
  { text: "Payment term", value: "kind" },
  { text: "Billing period", value: "period" },
  { text: "Base price", value: "basePrice" },
  { text: "Price", value: "prices" },
]);
const totalNewPrice = ref(0);
const totalBasePrice = ref(0);
const priceModelDialog = ref(false);
const isAddonsDialog = ref(false);

onMounted(() => {
  initPrices();
});

const accountTotalNewPrice = computed(() =>
  toAccountPrice(totalNewPrice.value)
);

const defaultCurrency = computed(() => store.getters["currencies/default"]);

const dueDate = computed(() => {
  return formatSecondsToDate(+template.value?.data?.next_payment_date, true);
});

const planCode = computed(() => template.value.config.planCode);
const type = computed(() => template.value.config.type);
const tarrif = computed(() => {
  const key = template.value.product;
  return { ...(template.value.billingPlan.products[key] || {}), key };
});

const service = computed(() =>
  store.getters["services/all"].find((s) => s.uuid === template.value.service)
);

const isPriceModelCanBeChange = computed(() => ["cloud"].includes(type.value));

const setTotalNewPrice = () => {
  totalNewPrice.value = +pricesItems.value.reduce(
    (acc, i) => +i.price + acc,
    0
  );
};

const onUpdatePrice = (item, isAccount) => {
  if (isAccount) {
    emit("update", {
      key: item.path,
      value: fromAccountPrice(item.accountPrice),
      type: item.isAddon ? "addons" : "template",
    });
    pricesItems.value = pricesItems.value.map((p) => {
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
    pricesItems.value = pricesItems.value.map((p) => {
      if (p.path === item.path) {
        p.accountPrice = toAccountPrice(item.price);
      }
      return p;
    });
  }
  setTotalNewPrice();
};

const initPrices = () => {
  pricesItems.value.push({
    title: tarrif.value.title,
    key: planCode.value,
    ind: 0,
    path: `billingPlan.products.${tarrif.value.key}.price`,
    kind: tarrif.value.kind,
    price: tarrif.value?.price,
    period: tarrif.value?.period,
    basePrice: toAccountPrice(tarrif.value.meta?.basePrice),
  });

  addons.value.forEach((addon, index) => {
    const { title, periods } = addon;
    const { period, kind } = tarrif.value;
    const price = periods[period];

    pricesItems.value.push({
      title,
      price,
      accountPrice: toAccountPrice(price),
      path: `${index}.periods.${period}`,
      quantity: 1,
      isAddon: true,
      basePrice: toAccountPrice(addon.meta?.basePrice),
      kind,
      period: tarrif.value?.period,
    });
  });

  pricesItems.value = pricesItems.value.map((i) => {
    i.period = getBillingPeriod(i.period);

    return i;
  });

  setAccountsPrices();
  setTotalNewPrice();
  setBasePrices();
};

const setAccountsPrices = () => {
  pricesItems.value = pricesItems.value.map((i) => {
    i.accountPrice = toAccountPrice(i.price);

    return i;
  });
};

const setBasePrices = () => {
  let total = 0;

  pricesItems.value = pricesItems.value.map((i) => {
    i.basePrice = i.basePrice / plnRate.value;
    total += +i.basePrice || 0;
    return i;
  });

  totalBasePrice.value = total;
};

watch(accountRate, () => {
  setAccountsPrices();
});
</script>

<style scoped></style>

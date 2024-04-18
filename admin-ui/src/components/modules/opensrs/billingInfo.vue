<template>
  <div>
    <v-row>
      <v-col>
        <route-text-field
          readonly
          label="Price model"
          :value="template.billingPlan.title"
          :to="{ name: 'Plan', params: { planId: template.billingPlan.uuid } }"
        />
      </v-col>

      <v-col>
        <v-text-field
          readonly
          label="Product name"
          :value="template.resources.domain"
        />
      </v-col>

      <v-col>
        <v-text-field
          readonly
          label="Price"
          :value="price"
          :suffix="account.currency"
        />
      </v-col>

      <v-col>
        <date-picker
          edit-icon
          label="Date (create)"
          :value="formatSecondsToDateString(template.created, false, '-')"
          :placeholder="formatSecondsToDate(template.created, true)"
          @input="
            emit('update', {
              key: 'created',
              value: formatDateToTimestamp($event),
            })
          "
          :clearable="false"
        />
      </v-col>

      <v-col>
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="formatSecondsToDate(template.data.expiry.expiredate, true)"
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
        <template v-slot:[`item.total`]="{ item }">
          {{ totalPrices[item.name] }}
          {{ defaultCurrency?.title }} /
          {{ totalAccountPrices[item.name] }}
          {{ accountCurrency?.title }}
        </template>
        <template v-slot:body.append>
          <tr>
            <td></td>
            <td>
              {{
                billingItems.find((i) => i.name === template.product)?.period
              }}
            </td>
            <td></td>
            <td>
              <div class="d-flex justify-end">
                <v-chip color="primary" outlined>
                  {{ totalPrice }}
                  {{ defaultCurrency?.title }} / {{ totalAccountPrice }}
                  {{ accountCurrency?.title }}
                </v-chip>
              </div>
            </td>
          </tr>
        </template>
      </nocloud-table>
    </instances-panels>
  </div>
</template>

<script setup>
import {
  computed,
  defineEmits,
  defineProps,
  onMounted,
  ref,
  toRefs,
} from "vue";
import {
  getBillingPeriod,
  formatDateToTimestamp,
  formatSecondsToDate,
  formatSecondsToDateString,
} from "@/functions";
import useCurrency from "@/hooks/useCurrency";
import InstancesPanels from "@/components/ui/nocloudExpansionPanels.vue";
import routeTextField from "@/components/ui/routeTextField.vue";
import NocloudTable from "@/components/table.vue";
import useInstancePrices from "@/hooks/useInstancePrices";
import DatePicker from "../../ui/datePicker.vue";

const props = defineProps(["template", "plans", "service", "sp", "account"]);
const emit = defineEmits(["refresh", "update"]);

const { template, account } = toRefs(props);

const { convertTo, defaultCurrency } = useCurrency();
const { accountCurrency, toAccountPrice, fromAccountPrice } = useInstancePrices(
  template.value,
  account.value
);

const billingItems = ref([]);
const billingHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Payment term", value: "kind" },
  { text: "Billing period", value: "period" },
  { text: "Price", value: "price" },
]);

onMounted(() => {
  billingItems.value = getBillingItems();
});

const totalPrices = computed(() =>
  billingItems.value.reduce((acc, i) => {
    acc[i.name] = i.price;
    return acc;
  }, {})
);

const totalAccountPrices = computed(() =>
  billingItems.value.reduce((acc, i) => {
    acc[i.name] = i.accountPrice;
    return acc;
  }, {})
);

const totalPrice = computed(() =>
  Object.keys(totalPrices.value || {}).reduce(
    (acc, key) => acc + totalPrices.value[key],
    0
  )
);
const totalAccountPrice = computed(() => toAccountPrice(totalPrice.value));

const price = computed(() => {
  return convertTo(
    template.value.billingPlan.resources[0]?.price || 0,
    account.value.currency
  );
});

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

const getBillingItems = () => {
  const items = [];

  if (template.value.billingPlan.resources[0]) {
    const { price, kind, period, key } =
      template.value.billingPlan.resources[0];
    items.push({
      name: key,
      price,
      accountPrice: toAccountPrice(price),
      path: `billingPlan.resources.0.price`,
      kind,
      period: getBillingPeriod(period),
    });
  }
  return items;
};
</script>

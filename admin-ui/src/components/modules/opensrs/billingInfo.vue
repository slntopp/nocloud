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
        <v-text-field
          readonly
          label="Date (create)"
          :value="formatSecondsToDate(template.data.creation)"
        />
      </v-col>

      <v-col>
        <v-text-field
          readonly
          label="Due to date/next payment"
          :value="template.data.expiry.expiredate"
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
        <template v-slot:[`item.total`]="{ item }">
          {{ totalPrices[item.name] }}
          {{ defaultCurrency }} /
          {{ totalAccountPrices[item.name] }}
          {{ accountCurrency }}
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
                 {{ defaultCurrency }} / {{ totalAccountPrice }}
                 {{ accountCurrency }}
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
import { formatSecondsToDate, getBillingPeriod } from "@/functions";
import useCurrency from "@/hooks/useCurrency";
import { useStore } from "@/store";
import InstancesPanels from "@/components/ui/nocloudExpansionPanels.vue";
import NocloudTable from "@/components/table.vue";
import useInstancePrices from "@/hooks/useInstancePrices";

const props = defineProps(["template", "plans", "service", "sp"]);
const emit = defineEmits(["refresh", "update"]);

const { template } = toRefs(props);

const { convertTo, defaultCurrency } = useCurrency();
const store = useStore();
const { accountCurrency, toAccountPrice, fromAccountPrice } = useInstancePrices(
  template.value
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

const account = computed(() => {
  const namespace = store.getters["namespaces/all"]?.find(
    (n) => n.uuid === template.value?.access.namespace
  );
  const account = store.getters["accounts/all"].find(
    (a) => a.uuid === namespace?.access.namespace
  );
  return account;
});

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

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
          :value="billingPlan.products[template.product].title"
          @click:append="priceModelDialog = true"
          append-icon="mdi-pencil"
        />
      </v-col>
      <v-col>
        <date-picker
          label="Date (create)"
          :value="timestampToDateTimeLocal(template?.created)"
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
        <date-picker
          label="Deleted date"
          :value="timestampToDateTimeLocal(template?.deleted)"
          :clearable="false"
          @input="
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
          :value="dueDate"
          :append-icon="!isMonitoringEmpty ? 'mdi-pencil' : null"
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
          <v-chip v-if="item.isAddon" small class="ml-1">Addon</v-chip>
        </template>
        <template v-slot:[`item.price`]="{ item }">
          <div class="d-flex">
            <v-text-field
              class="mr-2"
              v-model="item.price"
              :suffix="defaultCurrency?.code"
              @input="updatePrice(item, false)"
              append-icon="mdi-pencil"
            />
            <v-text-field
              class="ml-2"
              :suffix="accountCurrency?.code"
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

                <v-chip color="primary" outlined>
                  {{
                    [
                      formatPrice(totalPrice, defaultCurrency),
                      defaultCurrency?.code,
                    ].join(" ")
                  }}
                  /
                  {{
                    [
                      formatPrice(totalAccountPrice, accountCurrency),
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
      :account-rate="accountRate"
      :account-currency="accountCurrency"
      v-model="priceModelDialog"
      :template="template"
      @refresh="emit('refresh')"
      :service="service"
    />

    <change-ione-monitorings
      :template="template"
      :service="service"
      v-model="changeDatesDialog"
      @refresh="emit('refresh')"
      v-if="
        template.billingPlan.title.toLowerCase() !== 'payg' || isMonitoringEmpty
      "
    />
  </div>
</template>

<script setup>
import { computed, defineProps, toRefs, ref, watch, onMounted } from "vue";
import {
  formatSecondsToDate,
  getBillingPeriod,
  formatDateToTimestamp,
} from "@/functions";
import EditPriceModel from "@/components/dialogs/editPriceModel.vue";
import useInstancePrices from "@/hooks/useInstancePrices";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";
import InstancesPanels from "@/components/ui/nocloudExpansionPanels.vue";
import DatePicker from "../../ui/dateTimePicker.vue";
import InstanceChangeAddons from "@/components/InstanceChangeAddons.vue";
import { formatPrice, timestampToDateTimeLocal } from "../../../functions";
import ChangeIoneMonitorings from "@/components/dialogs/changeMonitorings.vue";

const props = defineProps(["template", "service", "sp", "account", "addons"]);
const emit = defineEmits(["refresh"]);

const { template, service, account, addons } = toRefs(props);

const store = useStore();
const { accountCurrency, toAccountPrice, accountRate, fromAccountPrice } =
  useInstancePrices(template.value, account.value);

const priceModelDialog = ref(false);
const isAddonsDialog = ref(false);
const changeDatesDialog = ref(false);

const billingHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Payment term", value: "kind" },
  { text: "Billing period", value: "period" },
  { text: "Price", value: "price" },
]);
const billingItems = ref([]);

const dueDate = computed(() =>
  formatSecondsToDate(template.value?.data?.next_payment_date, true)
);

const isMonitoringEmpty = computed(() => dueDate.value === "-");

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

const getBillingItems = () => {
  const items = [];
  const product = billingPlan.value.products[template.value.product];
  items.push({
    name: product.title,
    price: product?.price,
    path: `billingPlan.products.${template.value.product}.price`,
    kind: product?.kind,
    period: product?.period,
    accountPrice: toAccountPrice(product?.price),
  });

  items.push(
    ...addons.value.map((a, index) => ({
      name: a.title,
      price: a?.periods?.[product?.period] || 0,
      kind: a?.kind,
      path: `${index}.periods.${product?.period}`,
      period: product?.period,
      accountPrice: toAccountPrice(a?.periods?.[product?.period] || 0),
      isAddon: true,
    }))
  );

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
</script>

<style scoped></style>

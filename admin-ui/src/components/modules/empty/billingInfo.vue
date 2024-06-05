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
          label="Start date"
          :value="formatSecondsToDate(template.data.start) || '-'"
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
          :value="dueDate"
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
            <v-dialog
              v-model="isAddonsOpen"
              persistent
              v-if="addons?.length"
              max-width="60%"
            >
              <template v-slot:activator="{ on, attrs }">
                <v-btn class="ml-2" color="primary" v-bind="attrs" v-on="on"
                  >addons</v-btn
                >
              </template>
              <v-card color="background-light" class="pa-5">
                <nocloud-table
                  :items="newAddons"
                  :headers="addonsHeaders"
                  no-hide-uuid
                  :show-select="false"
                  hide-default-footer
                >
                  <template v-slot:[`item.name`]="{ item }">
                    <span v-html="item.name" />
                  </template>
                  <template v-slot:[`item.enabled`]="{ item }">
                    <v-switch v-model="item.enabled"></v-switch>
                  </template>
                </nocloud-table>
                <div class="d-flex justify-end mt-3">
                  <v-btn
                    class="mx-1"
                    @click="resetAddons"
                    :disabled="isAddonsLoading"
                    >Cancel</v-btn
                  >
                  <v-menu offset-y>
                    <template v-slot:activator="{ on, attrs }">
                      <v-btn
                        class="mx-1"
                        v-bind="newEnabledAddons.length ? attrs : undefined"
                        v-on="newEnabledAddons.length ? on : undefined"
                        :disabled="!isAddonsEdited"
                        :loading="isAddonsLoading"
                        @click="saveAddonsClick"
                        >Save</v-btn
                      >
                    </template>
                    <v-card color="background-light">
                      <v-card-title
                        >Make a payment now (balance will be
                        debited)?</v-card-title
                      >
                      <v-card-actions class="d-flex justify-end">
                        <v-btn class="mr-2" @click="saveAddons(true)">
                          No
                        </v-btn>
                        <v-btn class="mr-2" @click="saveAddons(false)">
                          Yes
                        </v-btn>
                      </v-card-actions>
                    </v-card>
                  </v-menu>
                </div>
              </v-card>
            </v-dialog>
            <v-chip class="ml-1" v-if="enabledAddonsCount > 0">{{
              enabledAddonsCount
            }}</v-chip>
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
import {
  getBillingPeriod,
  formatDateToTimestamp,
  formatSecondsToDate,
  formatSecondsToDateString,
} from "@/functions";
import ChangeMonitorings from "@/components/dialogs/changeMonitorings.vue";
import EditPriceModel from "@/components/dialogs/editPriceModel.vue";
import useInstancePrices from "@/hooks/useInstancePrices";
import NocloudTable from "@/components/table.vue";
import { useStore } from "@/store";
import api from "@/api";
import DatePicker from "../../ui/datePicker.vue";

const props = defineProps(["template", "plans", "service", "sp", "account"]);
const emit = defineEmits(["refresh"]);

const { template, plans, service, account } = toRefs(props);

const store = useStore();
const { accountCurrency, toAccountPrice, accountRate, fromAccountPrice } =
  useInstancePrices(template.value, account.value);

const changeDatesDialog = ref(false);
const priceModelDialog = ref(false);
const newAddons = ref([]);
const isAddonsOpen = ref(false);
const isAddonsLoading = ref(false);

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

const dueDate = computed(() =>
  formatSecondsToDate(template.value?.data?.next_payment_date, true)
);
const isMonitoringsEmpty = computed(() => dueDate.value === "-");

const filtredPlans = computed(() =>
  plans.value.filter((p) => p.type === "empty")
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

const isAddonsEdited = computed(
  () => JSON.stringify(newAddons.value) !== JSON.stringify(addons.value)
);

onMounted(() => {
  billingItems.value = getBillingItems();
  newAddons.value = JSON.parse(JSON.stringify(addons.value));
});

watch(accountRate, () => {
  billingItems.value = billingItems.value?.map((i) => {
    i.accountPrice = toAccountPrice(i.price);
    return i;
  });
});

const addons = computed(() => {
  return (
    billingPlan.value.products[template.value.product].meta.addons
      ?.map((key) => billingPlan.value.resources.find((r) => r.key === key))
      ?.filter((a) => !!a)
      ?.map(({ price, title, kind, period, key }, index) => ({
        name: title,
        price,
        enabled: !!template.value.config?.addons?.find((a) => a === key),
        path: `billingPlan.resources.${index}.price`,
        kind,
        key,
        period: getBillingPeriod(period),
        accountPrice: toAccountPrice(price),
      })) || []
  );
});

const enabledAddonsCount = computed(() => {
  return addons.value.filter((a) => a.enabled).length;
});

const newEnabledAddons = computed(() =>
  newAddons.value.filter(
    (a, index) => a.enabled === true && addons.value[index].enabled === false
  )
);

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

  items.push(
    ...addons.value
      .filter((a) => a.enabled)
      .map((a) => ({ ...a, isAddon: true }))
  );

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

const resetAddons = () => {
  newAddons.value = JSON.parse(JSON.stringify(addons.value));
  isAddonsOpen.value = false;
};

const saveAddons = async (isSkipped) => {
  const tempService = JSON.parse(JSON.stringify(service.value));
  const instance = JSON.parse(JSON.stringify(template.value));
  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  const skipped = [];
  skipped.push(template.value.product);
  skipped.push(...(template.value?.config?.addons || []));

  const addons = newAddons.value.filter((a) => a.enabled).map((a) => a.key);

  if (isSkipped) {
    instance.config = {
      ...instance.config,
      skip_next_payment: [
        ...new Set(
          (instance.config?.skip_next_payment || []).concat(
            ...newEnabledAddons.value.map((a) => a.key)
          )
        ),
      ],
    };
  }
  instance.config.addons = addons;

  tempService.instancesGroups[igIndex].instances[instanceIndex] = instance;
  try {
    isAddonsLoading.value = true;
    await api.services._update(tempService);
    emit("refresh");
  } catch (err) {
    store.commit("snackbar/showSnackbarError", { message: err });
  } finally {
    isAddonsLoading.value = false;
  }
};
const saveAddonsClick = () => {
  if (newEnabledAddons.value.length === 0) {
    saveAddons();
  }
};
</script>

<style scoped></style>

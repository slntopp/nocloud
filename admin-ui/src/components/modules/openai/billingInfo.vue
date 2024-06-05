<template>
  <div>
    <v-row>
      <v-col cols="2">
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
      <v-col cols="2">
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
          <v-text-field
            :suffix="defaultCurrency"
            v-model="item.price"
            @input="updatePrice(item, false)"
            append-icon="mdi-pencil"
          />
        </template>
        <template v-slot:[`item.accountPrice`]="{ item }">
          <v-text-field
            :suffix="accountCurrency"
            style="color: var(--v-primary-base)"
            v-model="item.accountPrice"
            @input="updatePrice(item, true)"
            append-icon="mdi-pencil"
          />
        </template>
      </nocloud-table>
    </instances-panels>
    <v-dialog persistent :value="priceModelDialog" max-width="60%">
      <v-card class="pa-5">
        <v-card-title class="text-center">Change price model</v-card-title>
        <v-row align="center">
          <v-col cols="12">
            <v-autocomplete
              label="price model"
              item-text="title"
              item-value="uuid"
              return-object
              v-model="newPlan"
              :items="filteredPlans"
            />
          </v-col>
        </v-row>

        <v-row justify="end">
          <v-btn class="mx-3" @click="priceModelDialog = false">Close</v-btn>
          <v-btn
            class="mx-3"
            :loading="isChangePMLoading"
            :disabled="isChangeBtnDisabled"
            @click="changePM"
            >Change price model</v-btn
          >
        </v-row>
      </v-card>
    </v-dialog>
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
import useInstancePrices from "@/hooks/useInstancePrices";
import { useStore } from "@/store";
import api from "@/api";
import {
  formatDateToTimestamp,
  formatSecondsToDate,
  formatSecondsToDateString,
} from "@/functions";
import InstancesPanels from "@/components/ui/nocloudExpansionPanels.vue";
import DatePicker from "../../ui/datePicker.vue";

const props = defineProps(["template", "plans", "service", "sp", "account"]);
const emit = defineEmits(["refresh", "update"]);

const { template, plans, service, account } = toRefs(props);

const store = useStore();
const { accountCurrency, toAccountPrice, accountRate, fromAccountPrice } =
  useInstancePrices(template.value, account.value);

const priceModelDialog = ref(false);
const newPlan = ref();
const isChangePMLoading = ref(false);

const billingItems = ref([]);
const billingHeaders = ref([
  { text: "Name", value: "name" },
  { text: "Price per thousand", value: "price" },
  { text: "Account price per thousand", value: "accountPrice" },
]);

onMounted(() => {
  billingItems.value = getBillingItems();
});
const filteredPlans = computed(() =>
  plans.value.filter((p) => p.type === "openai")
);
const defaultCurrency = computed(() => store.getters["currencies/default"]);
const isChangeBtnDisabled = computed(() => !newPlan.value);

const changePM = () => {
  const tempService = JSON.parse(JSON.stringify(service.value));

  const igIndex = tempService.instancesGroups.findIndex((ig) =>
    ig.instances.find((i) => i.uuid === template.value.uuid)
  );
  const instanceIndex = tempService.instancesGroups[
    igIndex
  ].instances.findIndex((i) => i.uuid === template.value.uuid);

  tempService.instancesGroups[igIndex].instances[instanceIndex].billingPlan =
    newPlan.value;

  isChangePMLoading.value = true;
  api.services
    ._update(tempService)
    .then(() => {
      emit("refresh");
    })
    .finally(() => {
      isChangePMLoading.value = false;
      emit("input", false);
    });
};

const getBillingItems = () => {
  const items = [];

  const acceptedResources = [
    "input_kilotoken",
    "output_kilotoken",
    "image_size_1024_1024_quality_standart",
    "image_size_1024_1024_quality_hd",
    "image_size_1024_1792_quality_standart",
    "image_size_1024_1792_quality_hd",
  ];

  template.value.billingPlan.resources.forEach((r, index) => {
    if (acceptedResources.includes(r.key)) {
      items.push({
        name: r.key.replace("_kilotoken", ""),
        price: r.price,
        accountPrice: toAccountPrice(r.price),
        path: `billingPlan.resources.${index}.price`,
      });
    }
  });

  return items;
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

watch(accountRate, () => {
  billingItems.value = billingItems.value.map((i) => {
    i.accountPrice = toAccountPrice(i.price);
    return i;
  });
});
</script>

<style lang="scss">
.openai-billing {
  .v-expansion-panel-content__wrap {
    padding: 0px !important;
  }
}
</style>

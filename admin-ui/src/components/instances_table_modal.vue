<template>
  <component
    offset-y
    max-width="1280"
    :is="type === 'dialog' ? VDialog : VMenu"
    :value="visible"
    :close-on-content-click="false"
    @input="emits('close')"
  >
    <template v-slot:activator="{ on, attrs }">
      <slot v-on="on" v-bind="attrs" name="activator" />
    </template>

    <v-card
      class="pa-4"
      elevation="0"
      color="background"
      @mouseenter="emits('hover')"
      @mouseleave="emits('close')"
    >
      <nocloud-table
        table-name="instances-table-modal-chats"
        style="margin-top: 0 !important"
        :items="instances"
        :headers="headers"
        :loading="isLoading"
        :server-items-length="total"
        :server-side-page="page"
        :show-select="false"
        @update:options="setOptions"
        :hide-default-footer="true"
      >
        <template v-slot:[`item.title`]="{ item }">
          <router-link
            target="_blank"
            :to="{ name: 'Instance', params: { instanceId: item.uuid } }"
          >
            {{ getShortName(item.title, 45) }}
          </router-link>
        </template>

        <template v-slot:[`item.data.next_payment_date`]="{ item }">
          {{ formatSecondsToDate(item.data?.next_payment_date) || "Unknown" }}
        </template>

        <template v-slot:[`item.state.state`]="{ item }">
          <instance-state small :template="item" />
        </template>

        <template v-slot:[`item.product`]="{ item }">
          <router-link
            :to="{ name: 'Plan', params: { planId: item.billingPlan?.uuid } }"
          >
            {{
              getShortName(item.billingPlan?.products?.[item.product]?.title)
            }}
          </router-link>
        </template>

        <template v-slot:[`item.accountPrice`]="{ item }">
          <template v-if="item.estimate">
            {{ formatPrice(item.estimate, defaultCurrency) }}
            {{ defaultCurrency?.code }}
          </template>
          <template v-else>-</template>
        </template>
      </nocloud-table>
    </v-card>
  </component>
</template>

<script setup>
import { VDialog, VMenu } from "vuetify/lib/components";
import nocloudTable from "@/components/table.vue";
import InstanceState from "@/components/ui/instanceState.vue";
import { createPromiseClient } from "@connectrpc/connect";
import { InstancesService } from "nocloud-proto/proto/es/instances/instances_connect";
import { ListInstancesRequest } from "nocloud-proto/proto/es/instances/instances_pb";
import { formatPrice, formatSecondsToDate, getShortName } from "@/functions";
import useCurrency from "@/hooks/useCurrency";
import { useStore } from "@/store";
import { ref, computed, watch, toRefs } from "vue";

const props = defineProps({
  uuid: { type: String, required: true },
  visible: { type: Boolean, required: true },
  type: { type: String, default: "dialog" },
});
const emits = defineEmits(["hover", "close"]);

const { uuid, visible } = toRefs(props);

const store = useStore();
const { defaultCurrency } = useCurrency();

const instances = ref([]);
const isLoading = ref(false);
const total = ref(0);
const page = ref(1);
const options = ref({});

const headers = [
  { text: "Title", value: "title" },
  { text: "Due date", value: "data.next_payment_date" },
  { text: "Status", value: "state.state" },
  { text: "Tariff", value: "product" },
  { text: "Price", value: "accountPrice" },
];

const instancesClient = computed(() =>
  createPromiseClient(InstancesService, store.getters["app/transport"]),
);

const setOptions = (newOptions) => {
  page.value = newOptions.page;
  if (JSON.stringify(newOptions) !== JSON.stringify(options.value)) {
    options.value = newOptions;
  }
};

const fetchInstances = async () => {
  if (!uuid.value) return;
  isLoading.value = true;
  try {
    const params = {
      filters: { account: [uuid.value] },
      page: page.value,
      limit: options.value.itemsPerPage || 10,
      field: options.value.sortBy?.[0],
      sort:
        options.value.sortBy?.[0] && options.value.sortDesc?.[0]
          ? "DESC"
          : "ASC",
    };
    const response = await instancesClient.value.list(
      ListInstancesRequest.fromJson(params),
    );
    instances.value = response.pool.map((i) => ({
      ...i,
      ...i.instance.toJson(),
      instance: undefined,
    }));
    total.value = Number(response.count);
  } catch (e) {
    console.error("Failed to fetch instances:", e);
  } finally {
    isLoading.value = false;
  }
};

watch([options, uuid], fetchInstances, { deep: true });
watch(visible, (val) => {
  if (val) fetchInstances();
});
</script>

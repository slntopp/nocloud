<template>
  <v-card color="background-light" class="mx-5 px-5" :loading="isFetchLoading">
    <v-card-title>Move accounts</v-card-title>
    <v-row>
      <v-col cols="3">
        <v-select
          v-model="sp"
          item-value="uuid"
          item-text="title"
          :items="ioneSps"
          label="Service provider"
        ></v-select>
      </v-col>
      <v-col cols="3">
        <v-select
          v-model="service"
          :items="services"
          label="Service"
          item-text="title"
          item-value="uuid"
        ></v-select>
      </v-col>
      <v-col cols="3">
        <v-text-field
          v-model.trim="instanceGroupName"
          label="Instance group name"
        ></v-text-field>
      </v-col>
    </v-row>
    <v-row>
      <v-col cols="3">
        <v-select
          v-model="plan"
          :items="plans"
          label="Plan"
          item-text="title"
          item-value="uuid"
        ></v-select>
      </v-col>
      <v-col cols="3">
        <v-select
          v-model="product"
          :items="products"
          label="Product"
        ></v-select>
      </v-col>
      <v-col class="d-flex align-center">
        <v-btn
          :loading="isMoveLoading"
          @click="move"
          :disabled="!isMoveAvailable"
          >Move</v-btn
        >
      </v-col>
    </v-row>
    <nocloud-table
      single-select
      :loading="isUsersLoading"
      no-hide-uuid
      v-model="selectedVM"
      :headers="headers"
      :items="users"
      table-name="nebulaAccounts"
      :footer-error="fetchError"
      show-expand
      :expanded.sync="expanded"
      item-key="data.userid"
    >
      <template v-slot:expanded-item="{ item, headers }">
        <td :colspan="headers.length" style="padding: 0">
          <template v-if="item.vms">
            <v-card
              class="px-5 ma-5"
              color="background-light"
              width="100%"
              v-for="machine in item.vms"
              :key="machine.vmid"
            >
              <v-row>
                <v-col cols="6">
                  <v-text-field
                    label="Instance title"
                    v-model="machine.title"
                  />
                </v-col>
              </v-row>
              <json-editor :json="machine"></json-editor>
            </v-card>
          </template>
          <v-card-title v-else class="text-center">No machines</v-card-title>
        </td>
      </template>
    </nocloud-table>
  </v-card>
</template>

<script setup>
import { computed, onMounted, watch, ref } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import NocloudTable from "@/components/table.vue";
import JsonEditor from "@/components/JsonEditor.vue";

const store = useStore();

const isFetchLoading = ref(false);
const sp = ref(null);
const plan = ref(null);
const service = ref(null);
const product = ref(null);
const instanceGroupName = ref("");
const selectedVM = ref([]);
const headers = ref([
  { text: "User id", value: "data.userid" },
  { text: "Public vn", value: "data.public_vn" },
]);
const users = ref([]);
const expanded = ref([]);
const fetchError = ref("");
const isUsersLoading = ref(false);
const isMoveLoading = ref(false);

onMounted(async () => {
  isFetchLoading.value = true;
  try {
    await Promise.all([
      store.dispatch("servicesProviders/fetch"),
      store.dispatch("services/fetch"),
      store.dispatch("plans/fetch"),
    ]);
  } catch (e) {
    fetchError.value = e.message;
  } finally {
    isFetchLoading.value = false;
  }
});

const ioneSps = computed(() =>
  store.getters["servicesProviders/all"].filter((sp) => sp.type === "ione")
);
const services = computed(() => store.getters["services/all"]);
const plans = computed(() =>
  store.getters["plans/all"].filter(
    (p) => p.type === "ione" && getPlanProducts(p)?.length
  )
);
const products = computed(() => getPlanProducts(selectedPlan.value));

const getPlanProducts = (plan) => {
  const products = [];
  if (!plan?.products || !selectedVM.value[0]) {
    return;
  }
  Object.keys(plan.products || {}).forEach((productKey) => {
    Object.keys(plan.products[productKey].resources || {}).forEach(
      (resourceKey) => {
        let stop = false;
        selectedVM.value[0]?.vms.forEach((vm) => {
          if (
            !stop &&
            vm.resources[resourceKey] !=
              plan.products[productKey].resources[resourceKey]
          ) {
            stop = true;
          }
        });
        if (!stop) {
          products.push(productKey);
        }
      }
    );
  });
  return products;
};

const selectedService = computed(() =>
  services.value.find((s) => s.uuid === service.value)
);
const selectedPlan = computed(() =>
  plans.value.find((p) => p.uuid === plan.value)
);

const isMoveAvailable = computed(
  () =>
    !!(sp.value &&
      service.value &&
      instanceGroupName.value &&
      selectedVM.value.length &&
      selectedVM.value[0].vms &&
      plan.value,
    product.value)
);

watch(sp, async () => {
  fetchError.value = "";
  isUsersLoading.value = true;
  expanded.value = [];
  selectedVM.value = [];
  try {
    const res = await api.servicesProviders.action({
      uuid: sp.value,
      action: "get_users",
    });
    users.value = res.meta.users;
    users.value.forEach((u) => {
      u.vms = u.vms?.map((vm, ind) => ({
        ...vm,
        title: "instance " + (ind + 1),
      }));
    });
  } catch (e) {
    fetchError.value = e.message;
  } finally {
    isUsersLoading.value = false;
  }
});

const move = async () => {
  const { data, resources, vms } = JSON.parse(
    JSON.stringify(selectedVM.value[0])
  );
  const service = JSON.parse(JSON.stringify(selectedService.value));
  try {
    isMoveLoading.value = true;
    service.instancesGroups.push({
      title: instanceGroupName.value,
      type: "ione",
      data,
      resources,
      sp: sp.value,
      instances: vms.map(({ config, resources, data, title }) => ({
        type: "ione",
        config,
        resources,
        data,
        title,
        product: product.value,
        billingPlan: selectedPlan.value,
      })),
    });
    await api.services._update(service);
  } finally {
    isMoveLoading.value = false;
  }
};
</script>

<style scoped></style>

<template>
  <v-card color="background-light" class="mx-5 px-5" :loading="isFetchLoading">
    <v-card-title>Move accounts</v-card-title>
    <v-row>
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
      table-name="nebula-accounts"
      :footer-error="fetchError"
      show-expand
      :expanded.sync="expanded"
      item-key="data.userid"
    >
      <template v-slot:expanded-item="{ item, headers }">
        <td :colspan="headers.length" style="padding: 0">
          <template v-if="item.vms">
            <nocloud-table
              :show-select="false"
              hide-default-footer
              :headers="machinesHeaders"
              :items="item.vms"
              item-key="data.vmid"
              table-name="nebula-machines"
              no-hide-uuid
            >
              <template v-slot:[`item.title`]="{ item }">
                <v-text-field v-model="item.title" />
              </template>
              <template v-slot:[`item.pass`]="{ item }">
                <password-text-field :value="item.config.password" />
              </template>
              <template v-slot:[`item.os`]="{ item }">
                {{
                  template?.publicData?.templates[item.config.template_id]?.name
                }}
              </template>
            </nocloud-table>
          </template>
          <v-card-title v-else class="text-center">No machines</v-card-title>
        </td>
      </template>
    </nocloud-table>
  </v-card>
</template>

<script>
export default {
  name: "nebula-tab",
};
</script>

<script setup>
import { computed, onMounted, ref, defineProps, toRefs } from "vue";
import { useStore } from "@/store";
import api from "@/api";
import NocloudTable from "@/components/table.vue";
import PasswordTextField from "@/components/ui/passwordTextField.vue";

const props = defineProps(["template"]);

const { template } = toRefs(props);

const store = useStore();

const isFetchLoading = ref(false);
const plan = ref(null);
const service = ref(null);
const product = ref(null);
const instanceGroupName = ref("");
const selectedVM = ref([]);
const headers = ref([
  { text: "User id", value: "data.userid" },
  { text: "Public vn", value: "data.public_vn" },
]);
const machinesHeaders = ref([
  { text: "Title", value: "title" },
  { text: "Vm ID", value: "data.vmid" },
  { text: "Vm name", value: "data.vm_name" },
  { text: "CPU", value: "resources.cpu" },
  { text: "RAM", value: "resources.ram" },
  { text: "OS", value: "os" },
  { text: "Drive size", value: "resources.drive_size" },
  { text: "Drive type", value: "resources.drive_type" },
  { text: "Ips private", value: "resources.ips_private" },
  { text: "Ips public", value: "resources.ips_public" },
  { text: "Password", value: "pass" },
]);
const users = ref([]);
const expanded = ref([]);
const fetchError = ref("");
const isUsersLoading = ref(false);
const isMoveLoading = ref(false);

onMounted(async () => {
  isUsersLoading.value = true;
  fetchError.value = "";
  isFetchLoading.value = true;
  try {
    await Promise.all([
      store.dispatch("services/fetch",{showDeleted:true}),
      store.dispatch("plans/fetch"),
    ]);

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
    store.commit("snackbar/showSnackbarError", {
      message: e.response.data.message || "Error during fetch vms",
    });
    fetchError.value = e.message;
  } finally {
    isUsersLoading.value = false;
    isFetchLoading.value = false;
  }
});

const services = computed(() => store.getters["services/all"]);
const plans = computed(() =>
  store.getters["plans/all"].filter(
    (p) => p.type === "ione" && getPlanProducts(p)?.length
  )
);
const products = computed(() => getPlanProducts(selectedPlan.value));
const sp = computed(() => template.value.uuid);

const getPlanProducts = (plan) => {
  const products = [];
  if (!plan?.products || !selectedVM.value[0]) {
    return;
  }
  Object.keys(plan.products || {}).forEach((productKey) => {
    const isValid = Object.keys(
      plan.products[productKey].resources || {}
    ).every((resourceKey) => {
      const isResourceValid = selectedVM.value[0]?.vms.every((vm) => {
        if (
          vm.resources[resourceKey] !=
          plan.products[productKey].resources[resourceKey]
        ) {
          return false;
        }
        return true;
      });
      if (isResourceValid) {
        return true;
      }
    });
    if (isValid) {
      products.push(productKey);
    }
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
    store.commit("snackbar/showSnackbarSuccess", {
      message: "Vm created successfully",
    });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during move nebula account",
    });
  } finally {
    isMoveLoading.value = false;
  }
};
</script>

<style scoped></style>

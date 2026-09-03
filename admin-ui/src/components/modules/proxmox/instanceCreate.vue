<template>
  <div class="module">
    <v-card
      v-if="Object.keys(instance).length > 1"
      class="mb-4 pa-2"
      elevation="0"
      color="background"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('title', newVal)"
            label="Name"
            :value="instance.title"
            :rules="requiredRule"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <plans-autocomplete
            :value="bilingPlan"
            :custom-params="{
              filters: { type: ['proxmox'], 'meta.isIndividual': [false] },
              anonymously: false,
            }"
            @input="changeBilling"
            return-object
            label="Price model"
            :rules="planRules"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Product"
            :disabled="isDynamicPlan"
            :rules="!isDynamicPlan ? requiredRule : []"
            :value="instance.product"
            :items="products"
            @change="setProduct"
          />
        </v-col>

        <v-col cols="6">
          <v-autocomplete
            @change="(newVal) => changeOS(newVal)"
            label="Template"
            :rules="existing ? [] : requiredRule"
            :items="osNames"
            :value="selectedTemplate?.name"
            :hint="selectedTemplate?.warning"
            persistent-hint
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('config.password', newVal)"
            label="Password"
            :rules="existing ? [] : requiredRule"
            :value="instance.config?.password"
          />
        </v-col>

        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('config.username', newVal)"
            label="Username"
            :value="instance.config?.username"
          />
        </v-col>
        <v-col cols="6">
          <v-textarea
            rows="1"
            auto-grow
            @change="(newVal) => setValue('config.ssh_public_key', newVal)"
            label="SSH public key"
            :value="instance.config?.ssh_public_key"
          />
        </v-col>

        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.cpu', +newVal)"
            label="CPU"
            :value="instance.resources.cpu"
            type="number"
            :rules="requiredRule"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            :rules="requiredRule"
            @change="(newVal) => setValue('resources.ram', +newVal)"
            label="RAM (MB)"
            :value="instance.resources.ram"
            type="number"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            :items="driveTypes"
            :rules="requiredRule"
            @change="(newVal) => setValue('resources.drive_type', newVal)"
            label="Drive type"
            :value="instance.resources.drive_type"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.drive_size', +newVal * 1024)"
            :label="`Drive size (minimum ${driveSizeConfig?.minDisk} GB, maximum ${driveSizeConfig?.maxDisk} GB)`"
            :rules="[driveSizeRule]"
            :value="instance.resources.drive_size / 1024"
            type="number"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.ips_public', +newVal)"
            label="IPs public"
            :value="instance.resources.ips_public"
            type="number"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue('resources.ips_private', +newVal)"
            label="IPs private"
            :value="instance.resources.ips_private"
            type="number"
            :disabled="privateBanned"
          />
        </v-col>
        <v-col cols="12">
          <v-combobox
            multiple
            small-chips
            deletable-chips
            label="Requested public IPs (optional, like ONE NIC IP=)"
            :hint="requestHint"
            persistent-hint
            :items="freePublicIPs"
            :value="requestedIPs"
            :rules="[requestRule]"
            @change="setRequested"
          />
        </v-col>

        <v-col cols="6" v-if="tarrifAddons.length > 0">
          <v-autocomplete
            @change="(newVal) => setValue('addons', newVal)"
            label="Addons"
            :value="instance.addons"
            :items="isAddonsLoading ? [] : getAvailableAddons()"
            :loading="isAddonsLoading"
            item-value="uuid"
            item-text="title"
            multiple
          />
        </v-col>
      </v-row>

      <!-- Existing: bind an already running PVE guest (TZ §15) -->
      <v-row align="center">
        <v-col cols="2">
          <v-switch label="Existing" v-model="existing" />
        </v-col>
        <template v-if="existing">
          <v-col cols="6">
            <v-autocomplete
              label="PVE guest (vmid — name @ node)"
              :items="existingVMs"
              :loading="isVMsLoading"
              :value="instance.data?.vmid"
              item-value="vmid"
              item-text="label"
              :rules="requiredRule"
              @change="pickExisting"
            />
          </v-col>
          <v-col cols="2">
            <v-text-field
              label="VMID"
              type="number"
              :value="instance.data?.vmid"
              :rules="requiredRule"
              @change="(v) => setValue('data.vmid', +v)"
            />
          </v-col>
          <v-col cols="2">
            <v-btn icon :loading="isVMsLoading" @click="fetchVMs">
              <v-icon>mdi-refresh</v-icon>
            </v-btn>
          </v-col>
          <v-col cols="12" v-if="vmsError">
            <v-alert dense text type="error">{{ vmsError }}</v-alert>
          </v-col>
        </template>
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store/";
import api from "@/api";
import useInstanceAddons from "@/hooks/useInstanceAddons";
import plansAutocomplete from "@/components/ui/plansAutoComplete.vue";

const props = defineProps(["instance", "planRules", "spUuid", "isEdit"]);
const { instance, isEdit, planRules, spUuid } = toRefs(props);

const emit = defineEmits(["set-instance", "set-value"]);

const store = useStore();
const { tarrifAddons, setTariffAddons, getAvailableAddons, isAddonsLoading } =
  useInstanceAddons(instance, (key, value) => setValue(key, value));

const getDefaultInstance = () => ({
  title: "instance",
  config: {
    template_id: "",
    password: "",
    username: "",
    auto_start: true,
  },
  resources: {
    cpu: 1,
    ram: 1024,
    drive_type: null,
    drive_size: null,
    ips_public: 1,
    ips_private: 0,
  },
  data: {},
  billing_plan: {},
  addons: [],
});

const bilingPlan = ref(null);
const products = ref([]);
const existing = ref(false);
const existingVMs = ref([]);
const isVMsLoading = ref(false);
const vmsError = ref("");
const requiredRule = ref([(val) => !!val || val === 0 || "Field required"]);

onMounted(() => {
  if (!isEdit.value) {
    emit("set-instance", getDefaultInstance());
    existing.value = !!(instance.value.data?.vmid || instance.value.data?.vm_id);
  } else {
    changeBilling(instance.value.billing_plan);
  }
});

const sp = computed(() =>
  store.getters["servicesProviders/all"].find((el) => el.uuid === spUuid.value)
);

const privateBanned = computed(
  () => !!sp.value?.vars?.private_vnet_ban?.value?.default
);

// ---- requested leases (config.ips_public_request), suggestions from the lease table ----
const freePublicIPs = ref([]);
const requestedIPs = computed(() => {
  const v = instance.value.config?.ips_public_request;
  if (Array.isArray(v)) return v;
  if (typeof v === "string" && v) return v.split(",").map((s) => s.trim()).filter(Boolean);
  return [];
});
const requestHint = computed(() => {
  const n = +instance.value.resources?.ips_public || 0;
  const free = sp.value?.publicData?.public_ips?.free;
  return `${requestedIPs.value.length}/${n} requested` + (free !== undefined ? `, ${free} free in the public network` : "");
});
const requestRule = (val) =>
  !val || val.length <= (+instance.value.resources?.ips_public || 0) || "more addresses requested than resources.ips_public";
const setRequested = (list) => {
  const ips = (list || []).map((s) => String(s).trim()).filter(Boolean);
  setValue("config.ips_public_request", ips.length ? ips : undefined);
};
const fetchFreeIPs = async () => {
  if (!spUuid.value) return;
  try {
    const { meta } = await api.servicesProviders.action({
      action: "get_leases",
      uuid: spUuid.value,
      params: { network: "public" },
    });
    freePublicIPs.value = (meta?.networks?.public?.leases || [])
      .filter((l) => l.state === "free")
      .map((l) => l.ip);
  } catch {
    freePublicIPs.value = [];
  }
};
watch(spUuid, fetchFreeIPs, { immediate: true });

const osTemplates = computed(() => {
  if (!sp.value) return {};
  const osTemplates = {};
  Object.keys(sp.value.publicData?.templates || {}).forEach((key) => {
    if (!instance.value?.billing_plan?.meta?.hidedOs?.includes(key)) {
      osTemplates[key] = sp.value.publicData.templates[key];
    }
  });
  return osTemplates;
});

const osNames = computed(() =>
  Object.values(osTemplates.value).map((os) => os.name)
);

const driveTypes = computed(() => {
  return instance.value.billing_plan?.resources
    ?.filter((r) => r.key.includes("drive"))
    .map((k) => k.key.split("_")[1].toUpperCase());
});

const driveSizeConfig = computed(() => {
  let minDisk, maxDisk;
  const dt = instance.value?.resources?.drive_type;
  if (instance.value.billing_plan?.meta?.minDiskSize) {
    minDisk = instance.value.billing_plan.meta.minDiskSize[dt];
  }
  if (instance.value.billing_plan?.meta?.maxDiskSize) {
    maxDisk = instance.value.billing_plan.meta.maxDiskSize[dt];
  }
  if (selectedTemplate.value?.min_size) {
    minDisk = Math.max(selectedTemplate.value.min_size / 1024, minDisk || 0);
  }
  return { minDisk: minDisk || 0, maxDisk: maxDisk || 100000 };
});

const selectedTemplate = computed(
  () => osTemplates.value[instance.value.config.template_id]
);

const driveSizeRule = computed(() => {
  return (val) =>
    existing.value ||
    (+val >= +driveSizeConfig.value.minDisk &&
      +val <= +driveSizeConfig.value.maxDisk) ||
    "Bad drive size";
});
const isDynamicPlan = computed(
  () => instance.value.billing_plan?.kind === "DYNAMIC"
);

const changeOS = (newVal) => {
  let osId = null;
  for (const [key, value] of Object.entries(osTemplates.value)) {
    if (value.name === newVal) {
      osId = key;
      break;
    }
  }
  setValue("config.template_id", +osId);
};

const changeBilling = (val) => {
  bilingPlan.value = val;
  if (bilingPlan.value) {
    products.value = Object.keys(bilingPlan.value.products || {});
  }
  setValue("billing_plan", bilingPlan.value);
};

const setProduct = (newVal) => {
  const product = bilingPlan.value?.products[newVal].resources;
  Object.keys(product || {}).forEach((key) => {
    emit("set-value", { key: "resources." + key, value: product[key] });
  });
  setValue("product", newVal);
  setTariffAddons();
};

const setValue = (key, value) => {
  emit("set-value", { key, value });
};

const fetchVMs = async () => {
  if (!spUuid.value) return;
  isVMsLoading.value = true;
  vmsError.value = "";
  try {
    const { meta } = await api.servicesProviders.action({
      action: "get_users",
      uuid: spUuid.value,
      params: { orphans: false },
    });
    existingVMs.value = (meta?.vms || [])
      .filter((vm) => !vm.nocloud)
      .map((vm) => ({
        ...vm,
        label: `${vm.vmid} — ${vm.name || "unnamed"} @ ${vm.node} (${vm.status})`,
      }));
  } catch (err) {
    vmsError.value = err.response?.data?.message || err.message || "Cannot list VMs";
  } finally {
    isVMsLoading.value = false;
  }
};

// Existing: prefill resources from the real guest so the driver adopts it as is.
const pickExisting = (vmid) => {
  const vm = existingVMs.value.find((v) => v.vmid === vmid);
  setValue("data.vmid", +vmid);
  if (!vm) return;
  setValue("data.vm_name", vm.name);
  if (vm.cpu) setValue("resources.cpu", +vm.cpu);
  if (vm.ram) setValue("resources.ram", +vm.ram);
  if (vm.disk) setValue("resources.drive_size", +vm.disk);
  if (vm.drive_type) setValue("resources.drive_type", vm.drive_type.toUpperCase());
  if (Array.isArray(vm.ips_public)) setValue("resources.ips_public", vm.ips_public.length);
  if (Array.isArray(vm.ips_private)) setValue("resources.ips_private", vm.ips_private.length);
  if (!instance.value.title || instance.value.title === "instance") {
    setValue("title", vm.name || `vm-${vmid}`);
  }
};

watch(existing, (val) => {
  setValue("data.vmid", null);
  setValue("data.vm_name", null);
  if (val && existingVMs.value.length === 0) fetchVMs();
});

watch(driveTypes, (newVal) => {
  if (newVal && newVal.length > 0 && !instance.value.resources.drive_type) {
    setValue("resources.drive_type", newVal[0]);
  }
});
</script>

<script>
export default {
  name: "instance-proxmox-create",
};
</script>

<style scoped></style>

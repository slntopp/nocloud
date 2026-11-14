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
            @change="setValue('title', $event.target.value)"
            label="Name"
            :model-value="instance.title"
            :rules="requiredRule"
          >
          </v-text-field>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <plans-autocomplete
            :value="bilingPlan"
            :custom-params="{
              filters: { type: ['ione'], 'meta.isIndividual': [false] },
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
            :model-value="instance.product"
            :items="products"
            @update:model-value="setProduct"
          />
        </v-col>

        <v-col cols="6">
          <v-autocomplete
            @update:model-value="(newVal) => changeOS(newVal)"
            label="Template"
            :rules="requiredRule"
            :items="osNames"
            :model-value="selectedTemplate?.name"
          >
          </v-autocomplete>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('config.password', $event.target.value)"
            label="Password"
            :rules="requiredRule"
            :model-value="instance.config?.password"
          >
          </v-text-field>
        </v-col>

        <v-col cols="6">
          <v-text-field
            @change="setValue('config.username', $event.target.value)"
            label="Username"
            :model-value="instance.config?.username"
          >
          </v-text-field>
        </v-col>

        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.cpu', +$event.target.value)"
            label="CPU"
            :model-value="instance.resources.cpu"
            type="number"
            :rules="requiredRule"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            :rules="requiredRule"
            @change="setValue('resources.ram', +$event.target.value)"
            label="RAM"
            :model-value="instance.resources.ram"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-select
            :items="driveTypes"
            :rules="requiredRule"
            @update:model-value="(newVal) => setValue('resources.drive_type', newVal)"
            label="Drive type"
            :model-value="instance.resources.drive_type"
          >
          </v-select>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.drive_size', +$event.target.value * 1024)"
            :label="`Drive size (minimum ${driveSizeConfig?.minDisk} GB, maximum ${driveSizeConfig?.maxDisk} GB)`"
            :rules="[driveSizeRule]"
            :model-value="instance.resources.drive_size / 1024"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.ips_public', +$event.target.value)"
            label="IPs public"
            :model-value="instance.resources.ips_public"
            type="number"
          >
          </v-text-field>
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="setValue('resources.ips_private', +$event.target.value)"
            label="IPs private"
            :model-value="instance.resources.ips_private"
            type="number"
          >
          </v-text-field>
        </v-col>

        <v-col cols="6" v-if="tarrifAddons.length > 0">
          <v-autocomplete
            @update:model-value="(newVal) => setValue('addons', newVal)"
            label="Addons"
            :model-value="instance.addons"
            :items="isAddonsLoading ? [] : getAvailableAddons()"
            :loading="isAddonsLoading"
            item-value="uuid"
            item-title="title"
            multiple
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="2">
          <v-switch label="Existing" v-model="existing" />
        </v-col>
        <template v-if="existing">
          <v-col>
            <v-text-field
              label="VM name"
              :model-value="instance.data?.vm_name"
              @change="setValue('data.vm_name', $event.target.value)"
            />
          </v-col>
          <v-col>
            <v-text-field
              label="VM id"
              :model-value="instance.data?.vm_id"
              @change="setValue('data.vm_id', $event.target.value)"
            />
          </v-col>
        </template>
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store/";
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
  },
  resources: {
    cpu: 1,
    ram: 1024,
    drive_type: null,
    drive_size: null,
    ips_public: 0,
    ips_private: 0,
  },
  data: {},
  billing_plan: {},
  addons: [],
});

const bilingPlan = ref(null);
const products = ref([]);
const existing = ref(false);
const requiredRule = ref([(val) => !!val || "Field required"]);

onMounted(() => {
  if (!isEdit.value) {
    emit("set-instance", getDefaultInstance());
    existing.value = !!(
      instance.value.data?.vm_id || instance.value.data?.vm_name
    );
  } else {
    changeBilling(instance.value.billing_plan);
  }
});

const osTemplates = computed(() => {
  const sp = store.getters["servicesProviders/all"].filter(
    (el) => el.uuid === spUuid.value
  )[0];

  if (!sp) return {};

  const osTemplates = {};

  Object.keys(sp.publicData.templates || {}).forEach((key) => {
    if (!instance.value?.billing_plan?.meta?.hidedOs?.includes(key)) {
      osTemplates[key] = sp.publicData.templates[key];
    }
  });

  return osTemplates;
});

const osNames = computed(() => {
  if (!osTemplates.value) return [];

  return Object.values(osTemplates.value).map((os) => os.name);
});

const driveTypes = computed(() => {
  return instance.value.billing_plan?.resources
    ?.filter((r) => r.key.includes("drive"))
    .map((k) => k.key.split("_")[1].toUpperCase());
});

const driveSizeConfig = computed(() => {
  let minDisk, maxDisk;
  if (instance.value.billing_plan?.meta?.minDiskSize) {
    minDisk =
      instance.value.billing_plan.meta.minDiskSize[
        instance.value?.resources?.drive_type
      ];
  }
  if (instance.value.billing_plan?.meta?.maxDiskSize) {
    maxDisk =
      instance.value.billing_plan.meta.maxDiskSize[
        instance.value?.resources?.drive_type
      ];
  }

  if (selectedTemplate.value?.min_size) {
    minDisk = Math.max(selectedTemplate.value?.min_size / 1024, minDisk);
  }

  return {
    minDisk: minDisk || 0,
    maxDisk: maxDisk || 100000,
  };
});

const selectedTemplate = computed(() => {
  return osTemplates.value[instance.value.config.template_id];
});

const driveSizeRule = computed(() => {
  return (val) =>
    (+val >= +driveSizeConfig.value.minDisk &&
      +val <= +driveSizeConfig.value.maxDisk) ||
    "Bad drive size";
});
const isDynamicPlan = computed(() => {
  return instance.value.billing_plan?.kind === "DYNAMIC";
});

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
    products.value = Object.keys(bilingPlan.value.products);
  }
  setValue("billing_plan", bilingPlan.value);
};

const setProduct = (newVal) => {
  const product = bilingPlan.value?.products[newVal].resources;

  Object.keys(product).forEach((key) => {
    emit("set-value", {
      key: "resources." + key,
      value: product[key],
    });
  });
  setValue("product", newVal);

  setTariffAddons();
};

const setValue = (key, value) => {
  emit("set-value", { key, value });
};

watch(existing, () => {
  setValue("data.vm_id", null);
  setValue("data.vm_name", null);
});

watch(driveTypes, (newVal) => {
  if (newVal && newVal.length > 0) {
    setValue("resources.drive_type", newVal[0]);
  }
});
</script>

<script>
export default {
  name: "instance-ione-create",
};
</script>

<style scoped></style>

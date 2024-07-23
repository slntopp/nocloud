<template>
  <div class="module">
    <v-card
      v-if="Object.keys(instance).length > 0"
      class="mb-4 pa-2"
      color="background"
      elevation="0"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            label="Name"
            :value="instance.title"
            :rules="rules.req"
            @change="(value) => setValue('title', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="Type"
            :items="ovhTypes"
            :rules="rules.req"
            v-model="ovhType"
            item-value="value"
            item-text="title"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="Price model"
            item-text="title"
            item-value="uuid"
            :value="instance.billing_plan"
            :items="filtredPlans"
            :rules="planRules"
            return-object
            @change="(value) => setValue('billing_plan', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-select
            :items="durationItems"
            :value="instance.config?.duration"
            item-value="value"
            item-text="title"
            label="Payment:"
            @change="(value) => setValue('config.duration', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Tariff"
            item-text="title"
            item-value="key"
            :value="instance.product"
            :items="filtredTariffs"
            :rules="rules.req"
            @change="(value) => setValue('product', value)"
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="Region"
            :value="instance.config?.configuration[`${ovhType}_datacenter`]"
            :items="datacenters"
            :rules="rules.req"
            :disabled="!datacenters.length"
            @change="
              (value) =>
                setValue(`config.configuration.${ovhType}_datacenter`, value)
            "
          />
        </v-col>
        <v-col cols="6">
          <v-autocomplete
            label="OS"
            :value="instance.config?.configuration[`${ovhType}_os`]"
            :items="images"
            item-text="title"
            item-value="id"
            :rules="rules.req"
            :disabled="!images.length"
            @change="
              (value) => setValue(`config.configuration.${ovhType}_os`, value)
            "
          />
        </v-col>
        <v-col v-if="ovhType === 'cloud'" cols="6">
          <v-text-field
            label="SSH"
            :value="instance?.config?.ssh"
            @change="setValue('config.ssh', $event)"
          />
        </v-col>
        <v-col cols="6" class="d-flex align-center">
          Existing:
          <v-switch
            class="d-inline-block ml-2"
            :input-value="instance.data?.existing"
            @change="(value) => setValue('data.existing', value)"
          />
        </v-col>
        <v-col
          cols="6"
          class="d-flex align-center"
          v-if="instance.data?.existing"
        >
          <v-text-field
            v-if="ovhType"
            :label="`${ovhType} name`"
            :value="instance.data?.[`${ovhType}Name`]"
            :rules="rules.req"
            @change="(value) => setValue(`data.${ovhType}Name`, value)"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6" v-for="type in addonsTypes" :key="type">
          <v-autocomplete
            :label="type.charAt(0).toUpperCase() + type.slice(1)"
            :items="addons.filter((a) => a.type === type)"
            item-text="title"
            item-value="uuid"
            :multiple="type === 'custom'"
            clearable
            return-object
            @change="changeAddons($event, type)"
          />
        </v-col>
      </v-row>
    </v-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import { useStore } from "@/store";

const getDefaultInstance = () => ({
  title: "instance",
  config: {
    auto_renew: false,
    type: "vps",
    planCode: null,
    configuration: {
      vps_datacenter: null,
      vps_os: null,
    },
    duration: "",
    pricingMode: "",
    addons: [],
  },
  data: { existing: false },
  resources: {},
  billing_plan: {},
});

const props = defineProps([
  "plans",
  "instance",
  "planRules",
  "spUuid",
  "meta",
  "isEdit",
  "instanceGroup",
]);
const { instance, planRules, plans } = toRefs(props);

const emit = defineEmits(["set-instance", "set-value"]);

const store = useStore();

const rules = ref({
  req: [(v) => !!v || "required field"],
});

const ovhTypes = [
  { title: "ovh vps", value: "vps" },
  { title: "ovh cloud", value: "cloud" },
  { title: "ovh dedicated", value: "dedicated" },
];
const ovhType = ref("vps");

const fetchedAddons = ref({});
const isAddonsLoading = ref(false);

const currentTariff = ref();

onMounted(() => {
  emit("set-instance", getDefaultInstance());
});

const images = computed(() => {
  if (!currentTariff.value) {
    return [];
  }

  return currentTariff.value.meta.os?.map((os) => ({
    title: os.name ? os.name : os,
    id: os.id ? os.id : os,
  }));
});

const datacenters = computed(() => {
  if (!currentTariff.value) {
    return [];
  }

  return currentTariff.value.meta.datacenter;
});

const addons = computed(() => {
  if (isAddonsLoading.value || !currentTariff.value) {
    return [];
  }

  return [
    ...(currentTariff.value.addons || []),
    ...(instance.value.billing_plan.addons || []),
  ]
    .map((uuid) => {
      const addon = fetchedAddons.value[uuid];
      if (!addon) {
        return;
      }

      const allowed = [
        "snapshot",
        "disk",
        "backup",
        "traffic",
        "ram",
        "softraid",
        "vrack",
        "storage",
        "system-storage",
        "bandwidth",
        "memory",
      ];

      const type =
        allowed.find((key) => addon.meta?.key?.includes(key)) || "custom";

      return {
        ...addon,
        type,
      };
    })
    .filter((addon) => !!addon);
});

const addonsTypes = computed(() => [
  ...new Set(addons.value?.map((item) => item.type)),
]);

const filtredPlans = computed(() =>
  plans.value?.filter((p) => p.type.includes(ovhType.value))
);

const tariffs = computed(() => {
  return Object.keys(instance.value?.billing_plan.products || {}).map(
    (key) => ({
      ...instance.value?.billing_plan.products[key],
      duration: key.split(" ")[0],
      planCode: key.split(" ")[1],
      key,
    })
  );
});

const filtredTariffs = computed(() =>
  tariffs.value.filter((t) => t?.duration === instance.value.config?.duration)
);

const durationItems = computed(() => {
  const annotations = {
    P1M: "Monthly",
    P1Y: "Yearly",
    P1D: "Daily",
    P1H: "Hourly",
  };

  return [...new Set(tariffs.value?.map((item) => item.duration))].map(
    (duration) => ({
      value: duration,
      title: annotations[duration] || duration,
    })
  );
});

const setValue = (path, val) => {
  if (path.includes("product")) {
    setValue("addons", []);

    if (val) {
      const product = instance.value.billing_plan.products[val];
      const resources = product.resources;

      let savedResources = {
        ips_private: 0,
        ips_public: 1,
      };
      switch (ovhType.value) {
        case "vps": {
          savedResources.cpu = +resources.cpu;
          savedResources.ram = resources.ram;
          savedResources.drive_size = resources.disk;
          savedResources.drive_type = "SSD";
          break;
        }
        case "cloud": {
          setValue(
            "config.monthlyBilling",
            instance.value.config?.duration === "P1M"
          );
          savedResources = { ...savedResources, ...resources };
          break;
        }
      }

      emit("set-value", {
        key: "resources",
        value: savedResources,
      });
      currentTariff.value = product;

      emit("set-value", {
        key: "addons",
        value: [],
      });

      setTimeout(() => {
        setAddons();
      }, 0);
    }
  }

  if (path.includes("billing_plan")) {
    setValue("product", undefined);
  }

  if (path.includes("duration")) {
    emit("set-value", {
      value: val === "P1Y" ? "upfront12" : "default",
      key: "config.pricingMode",
    });
    setValue("product", undefined);
  }

  emit("set-value", { value: val, key: path });
};

const setAddons = () => {
  if (!instance.value.billing_plan || !instance.value.product) {
    return;
  }

  isAddonsLoading.value = true;

  [
    ...(currentTariff.value.addons || []),
    ...instance.value.billing_plan.addons,
  ].forEach(async (uuid) => {
    try {
      if (!fetchedAddons.value[uuid]) {
        fetchedAddons.value[uuid] = store.getters["addons/addonsClient"].get({
          uuid,
        });
        fetchedAddons.value[uuid] = (await fetchedAddons.value[uuid]).toJson();
      }
    } catch {
      fetchedAddons.value[uuid] = undefined;
    } finally {
      setTimeout(() => {
        isAddonsLoading.value = Object.values(fetchedAddons.value).some(
          (acc) => acc instanceof Promise
        );
      }, 0);
    }
  });
};

const changeAddons = (value, type) => {
  let newAddons = instance.value.addons.map((uuid) =>
    addons.value.find((a) => a.uuid === uuid)
  );

  newAddons = newAddons.filter((a) => a.type !== type);

  if (value) {
    newAddons.push(...(Array.isArray(value) ? value : [value]));
  }

  if (ovhType.value === "dedicated") {
    newAddons.forEach((a) => {
      const addonKey = a.meta?.key;
      if (!addonKey) {
        return;
      }

      const resources = {};
      if (addonKey.includes("ram")) {
        resources.ram = parseInt(addonKey?.split("-")[1] ?? 0);
      }
      if (addonKey.includes("softraid")) {
        const [count, size] = addonKey?.split("-")[1].split("x") ?? ["0", "0"];

        resources.drive_size = count * parseInt(size) * 1024;
        if (addonKey?.includes("hybrid")) resources.drive_type = "SSD + HDD";
        else if (size.includes("sa")) resources.drive_type = false;
        else resources.drive_type = "SSD";
      }

      setValue("resources", {
        ...instance.value.resources,
        ...resources,
      });
    });
  }

  setValue(
    "addons",
    newAddons.map((a) => a.uuid)
  );
};

watch(ovhType, () => {
  const title = instance.value.title;
  emit("set-instance", getDefaultInstance());
  setValue("config.type", ovhType.value);
  setValue("title", title);
});
</script>

<script>
import { defaultFilterObject } from "@/functions";

export default {
  name: "instance-ovh-create",
};
</script>

<style scoped></style>

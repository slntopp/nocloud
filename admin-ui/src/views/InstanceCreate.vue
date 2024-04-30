<template>
  <v-form v-model="formValid" ref="form">
    <v-container>
      <v-row justify="start">
        <v-col cols="4" md="3" lg="2">
          <accounts-autocomplete
            label="account"
            v-model="account"
            :rules="rules.req"
            fetch-value
            :loading="isLoading"
            return-object
          />
        </v-col>
        <v-col v-if="servicesByAccount.length > 2" cols="4" md="3" lg="2">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="service"
            item-text="title"
            :items="servicesByAccount"
            return-object
            v-model="service"
            :rules="rules.req"
            :loading="isLoading"
          />
        </v-col>
        <v-col cols="4" md="3" lg="2">
          <v-autocomplete
            label="type"
            :items="typeItems"
            v-model="type"
            :rules="rules.req"
          />
        </v-col>
        <v-col cols="4" md="3" lg="2">
          <v-autocomplete
            :filter="defaultFilterObject"
            label="service provider"
            item-value="uuid"
            item-text="title"
            v-model="serviceProviderId"
            :items="servicesProviders"
            :loading="isLoading"
            readonly
            disabled
          />
        </v-col>
        <v-col v-if="type === 'custom'" cols="4" md="3" lg="2">
          <v-text-field
            v-model="customTypeName"
            label="Type name"
            :rules="rules.req"
          />
        </v-col>
        <v-col
          v-if="serviceInstanceGroupTitles.length > 2"
          cols="4"
          md="3"
          lg="2"
        >
          <v-autocomplete
            :items="serviceInstanceGroupTitles"
            v-model="instanceGroupTitle"
            label="instanceGroup"
            :rules="rules.req"
            :loading="isLoading"
          />
        </v-col>
      </v-row>
      <component
        :is-edit="isEdit || route.params.instanceId"
        v-if="isInstanceControlsShowed"
        @set-value="setValue"
        :instance-group="instanceGroup"
        @set-instance-group="instanceGroup = $event"
        @set-instance="instance = $event"
        @set-meta="meta = $event"
        :instance="instance"
        :plans="plans"
        :account-id="account?.uuid"
        :plan-rules="rules.req"
        :sp-uuid="serviceProviderId"
        :is="templates[type] ?? templates.custom"
      />
      <v-row class="mx-5" justify="end">
        <v-btn :loading="isCreateLoading" @click="save">Create</v-btn>
      </v-row>
    </v-container>
  </v-form>
</template>

<script setup>
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";
import api from "@/api";
import { computed, onMounted, watch, ref } from "vue";
import { defaultFilterObject } from "@/functions";
import { useStore } from "@/store";

const store = useStore();
const route = useRoute();
const router = useRouter();

const isEdit = ref(false);
const isCreateLoading = ref(false);
const isLoading = ref(true);
const formValid = ref(false);
const typeItems = ref([]);
const templates = ref([]);
const type = ref("ione");
const customTypeName = ref(null);
const instance = ref({});
const service = ref({});
const account = ref(null);
const namespace = ref(null);
const instanceGroup = ref({});
const instanceGroupTitle = ref("");
const serviceProviderId = ref(null);
const meta = ref({});
const form = ref(null);
const rules = ref({
  req: [(v) => !!v || "required field"],
});

const servicesProviders = computed(() =>
  store.getters["servicesProviders/all"].filter((el) => el.type === type.value)
);
const services = computed(() => store.getters["services/all"]);
const servicesByAccount = computed(() =>
  services.value.filter((s) => s.access.namespace === namespace.value?.uuid)
);

const serviceInstanceGroups = computed(() => {
  if (
    !service.value ||
    !service.value?.instancesGroups ||
    !type.value ||
    !serviceProviderId.value
  ) {
    return [];
  }

  return service.value?.instancesGroups.filter(
    (ig) => ig.type === type.value && ig.sp === serviceProviderId.value
  );
});
const serviceInstanceGroupTitles = computed(() => {
  const igs = serviceInstanceGroups.value.map((i) => i.title);

  return igs;
});
const allPlans = computed(() => {
  return store.getters["plans/all"] || [];
});
const plans = computed(() =>
  allPlans.value
    .map((plan) => ({
      ...plan,
      sp: servicesProviders.value.find((sp) =>
        sp.meta.plans?.includes(plan.uuid)
      )?.uuid,
    }))
    .filter((plan) => !!plan.sp)
);

const isInstanceControlsShowed = computed(
  () =>
    type.value &&
    (!(isEdit.value || route.params.instanceId) ||
      isEdit.value ||
      route.params.instanceId) &&
    !isLoading.value
);

onMounted(async () => {
  const types = require.context(
    "@/components/modules/",
    true,
    /instanceCreate\.vue$/
  );
  types.keys().forEach((key) => {
    const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/instanceCreate\.vue/i);
    if (matched && matched.length > 1) {
      const type = matched[1];
      typeItems.value.push(type);
      templates.value[type] = () =>
        import(`@/components/modules/${type}/instanceCreate.vue`);
    }
  });

  try {
    isLoading.value = true;
    await Promise.all([
      store.dispatch("services/fetch"),
      store.dispatch("servicesProviders/fetch", { anonymously: false }),
      store.dispatch("plans/fetch"),
    ]);
    const instanceId = route.params.instanceId;
    if (instanceId) {
      services.value.forEach((s) => {
        s.instancesGroups.forEach((ig) => {
          isEdit.value = true;
          const instance = ig.instances.find((i) => i.uuid === instanceId);
          if (instance) {
            type.value = ig.type;
            service.value = s;
            serviceProviderId.value = ig.sp;
            instanceGroupTitle.value = ig.title;
            instance.value = instance;
          }
        });
      });
      return;
    }

    let {
      type: newType,
      serviceId,
      accountId,
      serviceProviderId,
    } = route.params;

    if (newType) {
      service.value = services.value.find((s) => s.uuid === serviceId);
      if (!typeItems.value.includes(newType)) {
        customTypeName.value = newType;
        newType = "custom";
      }
      type.value = newType;
      serviceProviderId.value = serviceProviderId;
    }
    if (accountId) {
      account.value = await api.accounts.get(accountId);
    }
  } finally {
    isLoading.value = false;
  }
});

const setValue = ({ key, value }) => {
  const keys = key.split(".");
  switch (keys.length) {
    case 1: {
      return (instance.value[keys[0]] = value);
    }
    case 2: {
      return (instance.value[keys[0]][keys[1]] = value);
    }
    case 3: {
      return (instance.value[keys[0]][keys[1]][keys[2]] = value);
    }
  }
};
const save = async () => {
  if (!form.value.validate()) {
    return;
  }

  instance.value.config.auto_start =
    instance.value.billing_plan.meta.auto_start;
  if (typeof instance.value.billing_plan === "string") {
    instance.value.billing_plan = { uuid: instance.value.billing_plan };
  }

  const fullSp = servicesProviders.value.find(
    (sp) => sp.uuid === serviceProviderId.value
  );

  if (instance.value.type === "ovh") {
    instance.value.config.location = fullSp.locations.find(({ id }) =>
      id.includes(
        instance.value.config.configuration[
          `${instance.value.config.type}_datacenter`
        ]
      )
    )?.title;
  } else {
    instance.value.config.location = fullSp.locations[0]?.title;
  }

  isCreateLoading.value = true;
  try {
    const namespaceUuid = namespace.value?.uuid;

    if (!service.value) {
      service.value = await api.services.create({
        namespace: namespaceUuid,
        service: {
          version: "1",
          title: account.value.title,
          context: {},
          instancesGroups: [],
        },
      });
    }

    if (type.value === "opensrs") {
      await api.plans.update(
        instance.value.billing_plan.uuid,
        instance.value.billing_plan
      );
    }

    let igIndex = service.value.instancesGroups.findIndex(
      (i) => i.title === instanceGroupTitle.value
    );

    if (igIndex === -1) {
      service.value.instancesGroups.push({
        title:
          instanceGroupTitle.value ||
          account.value.title + new Date().getTime(),
        type: customTypeName.value || type.value,
        instances: [],
        sp: serviceProviderId.value,
      });
      igIndex = service.value.instancesGroups.length - 1;
    }

    service.value.instancesGroups[igIndex] = {
      ...service.value.instancesGroups[igIndex],
      ...instanceGroup.value,
    };

    if (isEdit.value) {
      const instanceIndex = service.value.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === instance.value.uuid);
      service.value.instancesGroups[igIndex].instances[instanceIndex] =
        instance.value;
    } else {
      service.value.instancesGroups[igIndex].instances.push(instance.value);
    }

    if (service.value.instancesGroups[igIndex].type === "ione") {
      service.value.instancesGroups[igIndex].resources = {
        ...(service.value.instancesGroups[igIndex].resources || {}),
        ips_public: service.value.instancesGroups[igIndex].instances
          .filter((i) => i.state?.state !== "DELETED")
          .reduce((acc, i) => acc + +i?.resources.ips_public, 0),
      };
    }

    const data = {
      namespace: namespaceUuid,
      service: service.value,
    };

    const response = await api.services._update(data.service);
    if (response.errors) {
      throw response;
    }
    store.commit("snackbar/showSnackbarSuccess", {
      message: isEdit.value
        ? "instance updated successfully"
        : "instance created successfully",
    });

    api.services.up(data.service.uuid);
    router.push({ name: "Instances" });
  } catch (err) {
    const opts = {
      message: err.errors.map((error) => error),
    };
    store.commit("snackbar/showSnackbarError", opts);
  } finally {
    isCreateLoading.value = false;
  }
};

const fetchNamespace = async () => {
  try {
    const { pool } = await store.dispatch("namespaces/fetch", {
      filters: { account: account.value.uuid },
    });
    namespace.value = pool[0];
  } catch (e) {
    console.log(e);
  }
};

watch(serviceProviderId, (sp_uuid) => {
  if (!sp_uuid) return;
  instanceGroupTitle.value = service.value?.instancesGroups.find(
    (ig) => ig.sp === sp_uuid
  )?.title;
});

watch(
  () => instance.value.billing_plan,
  (plan) => {
    if (!plan) {
      return;
    }
    serviceProviderId.value =
      plan.sp || plans.value.find((p) => p.uuid === plan)?.sp;
  }
);
watch(instanceGroupTitle, (newVal) => {
  instanceGroup.value = serviceInstanceGroups.value.find(
    (ig) => ig.title === newVal
  );
});

watch(account, () => {
  service.value = servicesByAccount.value?.[0];

  if (account.value.uuid) {
    fetchNamespace();
  }
});

watch(type, () => {
  if (isEdit.value) {
    isEdit.value = false;
    return;
  }
  if (type.value !== "custom") {
    customTypeName.value = null;
  }
  isEdit.value = false;
});

watch(plans, (newVal) => {
  if (newVal && instance.value?.billingPlan?.uuid) {
    instance.value.billing_plan = instance.value?.billingPlan?.uuid;
    delete instance.value.billingPlan;
  }
});
</script>

<script>
import snackbar from "@/mixins/snackbar.js";
import { useRoute, useRouter } from "vue-router/composables";

export default {
  name: "instance-create",
  mixins: [snackbar],
};
</script>

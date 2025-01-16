<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="template uuid"
          style="display: inline-block; width: 330px"
          :value="provider.uuid"
          :append-icon="copyed == 'rootUUID' ? 'mdi-check' : 'mdi-content-copy'"
          @click:append="addToClipboard(provider.uuid, 'rootUUID')"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="template type"
          style="display: inline-block; width: 150px"
          :value="provider.type"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="proxy"
          style="display: inline-block; width: 250px"
          :value="provider.proxy?.socket"
        />
      </v-col>
      <v-col>
        <v-switch readonly label="public" v-model="provider.public" />
      </v-col>
      <v-col cols="12">
        <div class="d-flex align-start">
          <v-btn
            :to="{
              name: 'ServicesProvider edit',
              params: { uuid: provider.uuid },
            }"
          >
            Edit
          </v-btn>
          <download-template-button
            :name="downloadedFileName"
            :template="template"
            class="mx-2"
            :type="isJson ? 'JSON' : 'YAML'"
          />
          <v-switch
            class="mr-2"
            style="margin-top: 5px; padding-top: 0"
            v-model="isJson"
            :label="!isJson ? 'YAML' : 'JSON'"
          />
        </div>
      </v-col>
    </v-row>

    <component :is="spTypes" :template="provider">
      <!-- Date -->
      <v-row>
        <v-col cols="12" lg="6" class="mt-5 mb-5">
          <v-alert type="info" color="primary">
            <span class="mr-2 text-h6">Last Monitored:</span>
            <template v-if="provider.state && template.state.meta.ts">
              {{
                format(
                  new Date(provider.state.meta.ts * 1000),
                  "dd MMMM yyy  H:mm"
                )
              }}
            </template>
            <template v-else>unknown</template>
          </v-alert>
        </v-col>
      </v-row>

      <!-- Plans -->
      <v-card-title class="px-0 mb-3">Plans:</v-card-title>
      <v-row class="flex-column">
        <v-col>
          <v-dialog max-width="60%" v-model="isDialogVisible">
            <template v-slot:activator="{ on, attrs }">
              <v-btn class="mr-2" v-bind="attrs" v-on="on"> Add </v-btn>
            </template>
            <v-card
              max-width="100%"
              class="ma-auto pa-5"
              color="background-light"
            >
              <plans-table
                no-search
                show-select
                :custom-params="plansParams"
                table-name="plans-sp-table"
                v-model="selectedNewPlans"
                :plans="plans"
                :total="totalPlans"
                :isLoading="isPlansLoading"
                @fetch:plans="fetchPlans"
              />

              <v-card-actions class="d-flex justify-end">
                <v-btn class="mr-5" @click="isDialogVisible = false">
                  Cancel
                </v-btn>
                <v-btn :loading="isLoading" @click="bindPlans">Add</v-btn>
              </v-card-actions>
            </v-card>
          </v-dialog>
          <confirm-dialog
            :disabled="selected.length < 1"
            @confirm="unbindPlans"
          >
            <v-btn :disabled="selected.length < 1" :loading="isDeleteLoading">
              Remove
            </v-btn>
          </confirm-dialog>
        </v-col>
        <v-col>
          <nocloud-table
            table-name="service-providers"
            :items="relatedPlans"
            :headers="headers"
            :footer-error="fetchError"
            v-model="selected"
          />
        </v-col>
      </v-row>
    </component>

    <template
      v-if="provider.extentions && Object.keys(provider.extentions).length > 0"
    >
      <v-card-title class="px-0">Extentions:</v-card-title>
      <component
        v-for="(extention, extName) in provider.extentions"
        :is="extentionsMap[extName].pageComponent"
        :key="extName"
        :data="extention"
      >
      </component>
    </template>

    <v-row> </v-row>
  </v-card>
</template>

<script setup>
import extentionsMap from "@/components/extentions/map.js";
import nocloudTable from "@/components/table.vue";
import plansTable from "@/components/plansTable.vue";
import ConfirmDialog from "@/components/confirmDialog.vue";
import { format } from "date-fns";
import DownloadTemplateButton from "@/components/ui/downloadTemplateButton.vue";
import { computed, onMounted, ref, toRefs } from "vue";
import api from "@/api.js";
import { useStore } from "@/store";

const props = defineProps({
  template: { type: Object, required: true },
});

const { template } = toRefs(props);

const store = useStore();

const copyed = ref(null);
const provider = ref({});
const isJson = ref(true);
const isLoading = ref(false);
const isDeleteLoading = ref(false);
const isDialogVisible = ref(false);

const fetchError = ref("");
const relatedPlans = ref([]);
const selected = ref([]);
const selectedNewPlans = ref([]);

const headers = ref([
  { text: "Title ", value: "title" },
  { text: "UUID ", value: "uuid" },
  { text: "Public ", value: "public" },
  { text: "Type ", value: "type" },
]);

onMounted(async () => {
  provider.value = template.value;
  if (!provider.value.proxy) {
    provider.value.proxy = { socket: "" };
  }

  fetchSelectedPlans();
});

const plans = computed(() => store.getters["plans/all"]);
const totalPlans = computed(() => store.getters["plans/total"]);
const isPlansLoading = computed(() => store.getters["plans/loading"]);

const fetchSelectedPlans = async () => {
  try {
    relatedPlans.value =
      (
        await store.getters["plans/plansClient"].listPlans({
          spUuid: template.value.uuid,
        })
      ).toJson().pool || [];
  } catch (err) {
    console.error(err);

    fetchError.value = "Can't reach the server";
    if (err.response) {
      fetchError.value += `: [ERROR]: ${err.response.data.message}`;
    } else {
      fetchError.value += `: [ERROR]: ${err.message}`;
    }
  }
};

const plansParams = computed(() => {
  const type = [provider.value.type];

  if (provider.value.type === "ione") {
    type.push("ione-vpn");
  }

  if (provider.value.type === "empty") {
    type.push("vpn");
  }

  return {
    showDeleted: false,
    excludeUuids: relatedPlans.value?.map((p) => p.uuid) || [],
    filters: {
      type,
    },
  };
});

const spTypes = computed(() => {
  switch (provider.value.type) {
    case "ione":
      return () => import("@/components/modules/ione/serviceProviderInfo.vue");
    case "ovh":
      return () => import("@/components/modules/ovh/serviceProviderInfo.vue");
    default:
      return () =>
        import("@/components/modules/custom/serviceProviderInfo.vue");
  }
});
const downloadedFileName = computed(() => {
  return template.value.title
    ? template.value.title.replaceAll(" ", "_")
    : "unknown_sp";
});

const addToClipboard = async (text, index) => {
  if (navigator?.clipboard) {
    await navigator.clipboard.writeText(text);
    copyed.value = index;
  } else {
    alert("Clipboard is not supported!");
  }
};

const bindPlans = async () => {
  if (selectedNewPlans.value.length < 1) return;
  isLoading.value = true;

  const plans = selectedNewPlans.value.map((el) => el.uuid);

  try {
    await api.servicesProviders.bindPlan(template.value.uuid, plans);

    const ending = plans.length === 1 ? "" : "s";
    relatedPlans.value.push(...selectedNewPlans.value);
    selectedNewPlans.value = [];
    isDialogVisible.value = false;

    store.commit("snackbar/showSnackbarSuccess", {
      message: `Price model${ending} added successfully.`,
    });
  } catch (err) {
    store.commit("snackbar/showSnackbarError", {
      message: err,
    });
  } finally {
    isLoading.value = false;
  }
};
const unbindPlans = async () => {
  isDeleteLoading.value = true;

  const plans = selected.value.map((el) => el.uuid);

  try {
    await api.servicesProviders.unbindPlan(template.value.uuid, plans);

    const ending = plans.length === 1 ? "" : "s";
    relatedPlans.value = relatedPlans.value.filter(
      (rp) => selected.value.findIndex((s) => s.uuid === rp.uuid) === -1
    );
    selected.value = [];
    store.commit("snackbar/showSnackbarSuccess", {
      message: `Price model${ending} deleted successfully.`,
    });
  } catch (err) {
    store.commit("snackbar/showSnackbarError", {
      message: err,
    });
  } finally {
    isDeleteLoading.value = false;
  }
};

const fetchPlans = (options) => {
  return store.dispatch("plans/fetch", options);
};
</script>

<script>
export default {
  name: "services-provider-info",
};
</script>

<style>
.title_progress {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.v-alert__icon.v-icon {
  margin-top: 5px;
}
.apexcharts-svg {
  background: none !important;
}
.ceil {
  display: inline-block;
  width: 15px;
  height: 15px;
  margin: 5px;
  vertical-align: middle;
  border-radius: 2px;
}
.occupied {
  background: var(--v-success-base);
}
.free {
  background: var(--v-error-base);
}
</style>

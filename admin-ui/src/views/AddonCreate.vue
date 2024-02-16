<template>
  <div class="pa-4">
    <div class="d-flex mb-5" v-if="!isEdit">
      <h1 class="page__title">Create addon group</h1>
    </div>
    <v-form v-model="isValid" ref="addonCreateForm">
      <v-row>
        <v-col cols="3" class="align-center d-flex">
          <v-subheader>Group title</v-subheader>
        </v-col>
        <v-col cols="3" class="align-center d-flex">
          <v-text-field
            label="Title"
            v-model="group.title"
            :rules="[rules.required]"
          />
        </v-col>
      </v-row>
      <v-divider />
      <v-row class="my-3 d-flex align-center">
        <v-subheader>Addons</v-subheader>
        <v-btn @click="addAddon" class="mx-1">Add</v-btn>
        <v-btn
          @click="deleteSelected"
          :disabled="selected.length === 0"
          class="mx-1"
          >Delete</v-btn
        >
      </v-row>
      <nocloud-table
        show-select
        v-model="selected"
        sort-by="uuid"
        :headers="headers"
        :items="group.addons"
      >
        <template v-slot:[`item.public`]="{ item }">
          <v-switch v-model="item.public" />
        </template>
        <template v-slot:[`item.title`]="{ item }">
          <v-text-field :rules="[rules.required]" v-model="item.title" />
        </template>
        <template v-slot:[`item.price`]="{ item }">
          <v-text-field
            :rules="[rules.required]"
            v-model.number="item.price"
            type="number"
          />
        </template>
        <template v-slot:[`item.period`]="{ item }">
          <date-field
            class="mt-3"
            :period="item.period"
            @changeDate="item.period = $event"
          />
        </template>
      </nocloud-table>
      <v-divider />
      <v-row class="mt-3">
        <v-col>
          <v-btn :loading="isSaveLoading" @click="saveGroup" class="mr-2">
            Save
          </v-btn>
        </v-col>
      </v-row>
    </v-form>
  </div>
</template>

<script setup>
import { ref, toRefs, watch } from "vue";
import NocloudTable from "@/components/table.vue";
import dateField from "@/components/date.vue";
import { useStore } from "@/store";
import { getTimestamp } from "@/functions";
import api from "@/api";
import { useRouter } from "vue-router/composables";

const props = defineProps({
  addonGroup: {},
  isEdit: { type: Boolean, default: false },
});
const { addonGroup, isEdit } = toRefs(props);

const store = useStore();
const router = useRouter();

const group = ref({ title: "", addons: [] });
const isSaveLoading = ref(false);
const isValid = ref(false);
const selected = ref([]);
const addonCreateForm = ref(null);

const headers = ref([
  { text: "Title", value: "title", sortable: false },
  { text: "Price", value: "price", sortable: false },
  { text: "Period", value: "period", sortable: false },
  { text: "Public", value: "public", sortable: false },
]);

const rules = ref({ required: (v) => !!v || "This field is required!" });

const addAddon = () => {
  group.value.addons.push({
    title: "",
    price: 0,
    period: null,
    public: true,
    uuid: group.value.addons.length + 1,
  });
};

const deleteSelected = () => {
  group.value.addons = group.value.addons
    .filter((a) => !selected.value.find((s) => s.uuid === a.uuid))
    .map((a, ind) => ({ ...a, uuid: ind }));
  selected.value = [];
};

const saveGroup = async () => {
  if (!(await addonCreateForm.value.validate())) {
    return;
  }
  isSaveLoading.value = true;

  try {
    const addons = group.value.addons.map((a) => ({
      ...a,
      group: group.value.title,
      period: getTimestamp(a.period),
      uuid: undefined,
    }));

    await Promise.all(addons.map((a) => api.put("/addons", a)));
    router.push({ name: "Addons" });
  } catch (e) {
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isSaveLoading.value = false;
  }
};

watch(addonGroup, () => {
  group.value = { ...addonGroup.value };
});
</script>

<style scoped></style>

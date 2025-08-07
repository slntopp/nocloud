<template>
  <div class="pa-4">
    <div class="d-flex mb-5" v-if="!isEdit">
      <h1 class="page__title">Create addon</h1>
    </div>
    <v-form v-model="isValid" ref="addonCreateForm">
      <v-row>
        <v-col cols="1" class="align-center d-flex">
          <v-subheader>Title</v-subheader>
        </v-col>
        <v-col cols="3" class="align-center d-flex">
          <v-text-field
            label="Title"
            v-model="newAddon.title"
            :rules="[rules.required]"
          />
        </v-col>
        <v-col cols="1" class="align-center d-flex">
          <v-subheader>Group</v-subheader>
        </v-col>
        <v-col cols="3" class="align-center d-flex">
          <v-text-field label="Group" v-model="newAddon.group" />
        </v-col>
        <v-col cols="2" class="align-center d-flex">
          <v-switch label="Public" v-model="newAddon.public" />
        </v-col>
      </v-row>
      <nocloud-expansion-panels title="Description" class="mb-5">
        <rich-editor class="pa-5" v-model="newAddon.description" />
      </nocloud-expansion-panels>
      <v-divider />
      <v-row class="my-3 d-flex align-center">
        <v-subheader>Prices</v-subheader>
        <v-btn @click="addPeriod" class="mx-1">Add</v-btn>
        <v-btn
          @click="deleteSelectedPeriods"
          :disabled="selected.length === 0"
          class="mx-1"
          >Delete</v-btn
        >
      </v-row>
      <nocloud-table
        show-select
        v-model="selected"
        sort-by="id"
        item-key="id"
        :headers="headers"
        :items="newAddon.periods"
      >
        <template v-slot:[`item.price`]="{ item }">
          <v-text-field
            :rules="[rules.required]"
            v-model.number="item.price"
            type="number"
            :suffix="defaultCurrency?.code"
          />
        </template>
        <template v-slot:[`item.oneTime`]="{ item }">
          <v-switch
            :input-value="item.period === 0"
            @change="changeOneTime(item, $event)"
          />
        </template>

        <template v-slot:[`item.period`]="{ item }">
          <date-field
            v-if="item.period !== 0"
            class="mt-3"
            :period="item.period"
            @changeDate="item.period = $event"
          />
        </template>
      </nocloud-table>
      <v-divider />
      <v-row class="mt-3">
        <v-col>
          <v-btn :loading="isSaveLoading" @click="saveAddon" class="mr-2">
            Save
          </v-btn>
        </v-col>
      </v-row>
    </v-form>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import NocloudTable from "@/components/table.vue";
import dateField from "@/components/date.vue";
import { useStore } from "@/store";
import { useRouter } from "vue-router/composables";
import RichEditor from "@/components/ui/richEditor.vue";
import NocloudExpansionPanels from "@/components/ui/nocloudExpansionPanels.vue";
import { Addon } from "nocloud-proto/proto/es/billing/addons/addons_pb";

const props = defineProps({
  addon: {},
  isEdit: { type: Boolean, default: false },
});
const { addon, isEdit } = toRefs(props);

const store = useStore();
const router = useRouter();

const newAddon = ref({
  title: "",
  periods: [],
  group: "",
  public: true,
  description: "",
});
const isSaveLoading = ref(false);
const isValid = ref(false);
const selected = ref([]);
const addonCreateForm = ref(null);

const headers = ref([
  { text: "Period", value: "period", sortable: false, width: "400px" },
  { text: "One time", value: "oneTime", sortable: false },
  { text: "Price", value: "price", sortable: false },
]);

const rules = ref({
  required: (v) => !!v || v === 0 || "This field is required!",
});

onMounted(() => {
  if (isEdit.value) {
    setAddon(addon.value);
  } else {
    addPeriod();
  }
});

const defaultCurrency = computed(() => store.getters["currencies/default"]);

const addPeriod = () => {
  newAddon.value.periods.push({
    price: 0,
    period: 2592000,
    id: newAddon.value.periods.length + 1,
  });
};

const deleteSelectedPeriods = () => {
  newAddon.value.periods = newAddon.value.periods
    .filter((a) => !selected.value.find((s) => s.id === a.id))
    .map((a, ind) => ({ ...a, id: ind }));
  selected.value = [];
};

const saveAddon = async () => {
  if (!(await addonCreateForm.value.validate())) {
    return;
  }
  isSaveLoading.value = true;

  try {
    const dto = Addon.fromJson({
      public: !!newAddon.value.public,
      title: newAddon.value.title || "",
      group: newAddon.value.group || "",
      meta: newAddon.value.meta || {},
      system: !!newAddon.value.system,
      periods: newAddon.value.periods.reduce((acc, a) => {
        acc[a.period] = a.price;
        return acc;
      }, {}),
      kind: "PREPAID",
    });

    if (!isEdit.value) {
      const description = await store.dispatch("descriptions/create", {
        text: newAddon.value.description,
      });
      dto.descriptionId = description.uuid;
      await store.getters["addons/addonsClient"].create(dto);
      router.push({ name: "Addons" });
    } else {
      if (addon.value.descriptionId) {
        store.dispatch("descriptions/update", {
          uuid: addon.value.descriptionId,
          text: newAddon.value.description,
        });
        dto.descriptionId = addon.value.descriptionId;
      } else {
        const description = await store.dispatch("descriptions/create", {
          text: newAddon.value.description,
        });
        dto.descriptionId = description.uuid;
      }

      dto.uuid = props.addon.uuid;
      await store.getters["addons/addonsClient"].update(dto);
    }
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", { message: e.message });
  } finally {
    isSaveLoading.value = false;
  }
};

const setAddon = (val) => {
  newAddon.value = {
    ...val,
    periods: Object.keys(val.periods || {}).map((key, ind) => ({
      period: key,
      price: val.periods[key],
      id: ind,
    })),
  };
};

function changeOneTime(item, value) {
  if (value) {
    item.period = 0;
  } else {
    item.period = 2592000;
  }
}

watch(addon, (val) => {
  setAddon(val);
});
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>

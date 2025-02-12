<template>
  <div class="pa-5">
    <div v-if="eventOverrides.length">
      <v-row
        v-for="eventOverride in eventOverrides"
        :key="eventOverride.key + eventOverride.override"
      >
        <v-col cols="5">
          <v-text-field
            dense
            hide-details
            :value="eventOverride.key"
            disabled
            label="Key"
          />
        </v-col>
        <v-col cols="5">
          <v-text-field
            dense
            hide-details
            :value="eventOverride.override"
            disabled
            label="Override"
          />
        </v-col>
        <v-col cols="2">
          <div class="d-flex align-center justify-end">
            <v-btn
              :loading="isDeleteOverrideLoading"
              @click="deleteOverride(eventOverride)"
              icon
            >
              <v-icon>mdi-delete</v-icon>
            </v-btn>
            <v-btn @click="editOverride(eventOverride)" icon>
              <v-icon>mdi-pencil</v-icon>
            </v-btn>
          </div>
        </v-col>
      </v-row>
    </div>

    <div v-else class="d-flex pa-5 justify-center align-center">
      <v-card-title>Event overrides empty</v-card-title>
    </div>

    <div class="d-flex justify-end mt-4">
      <v-btn @click="isAddOverrideOpen = true">Add new</v-btn>
    </div>

    <v-dialog max-width="600px" v-model="isAddOverrideOpen">
      <v-form v-model="isAddOverrideFormValid" ref="addOverrideForm">
        <v-card class="pa-5" color="background-light">
          <v-card-title v-if="isEditOverride">Edit event override</v-card-title>
          <v-card-title v-else>New event override</v-card-title>

          <v-text-field
            :rules="[requiredRule]"
            v-model="newOverride.key"
            label="Key"
          />
          <v-text-field
            :rules="[requiredRule]"
            v-model="newOverride.override"
            label="Override"
          />

          <div class="d-flex justify-center">
            <v-card-title style="color: red" v-if="isNotUniqueOverride">{{
              "Already overrided!"
            }}</v-card-title>
          </div>

          <div class="d-flex justify-end">
            <v-btn @click="isAddOverrideOpen = false">Close</v-btn>
            <v-btn
              @click="addNewEvent"
              :loading="isAddOverrideLoading"
              :disabled="isNotUniqueOverride"
              class="ml-3"
              >{{ isEditOverride ? "Edit" : "Add" }}</v-btn
            >
          </div>
        </v-card>
      </v-form>
    </v-dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs, watch } from "vue";
import api from "@/api";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const eventOverrides = ref([]);

const addOverrideForm = ref();
const isAddOverrideFormValid = ref();
const isAddOverrideOpen = ref(false);
const isAddOverrideLoading = ref(false);
const newOverride = ref({ key: "", override: "" });

const isDeleteOverrideLoading = ref(false);

const isEditOverride = ref(false);
const editedOverride = ref(false);

onMounted(() => setEventOverrides());

const requiredRule = (val) => !!val || "Field required";
const isNotUniqueOverride = computed(
  () =>
    !isEditOverride.value &&
    eventOverrides.value.find(
      (override) =>
        override.override === newOverride.value.override &&
        override.key === newOverride.value.key
    )
);

const setEventOverrides = () => {
  eventOverrides.value = [...template.value.customEvents];
};

const saveEventOverrides = (data) => {
  return api.plans.update(template.value.uuid, {
    ...template.value,
    customEvents: data,
  });
};

const addNewEvent = async () => {
  if (!addOverrideForm.value.validate()) {
    return;
  }

  try {
    isAddOverrideLoading.value = true;

    let data = [];
    if (isEditOverride.value) {
      data = eventOverrides.value.map((override) =>
        editedOverride.value.key === override.key &&
        editedOverride.value.override === override.override
          ? newOverride.value
          : override
      );
    } else {
      data = [...eventOverrides.value, { ...newOverride.value }];
    }

    await saveEventOverrides(data);
    eventOverrides.value = data;

    newOverride.value = { key: "", override: "" };
    isAddOverrideOpen.value = false;
    isEditOverride.value = false;
    editedOverride.value = {};
  } finally {
    isAddOverrideLoading.value = false;
  }
};

const deleteOverride = async (override) => {
  try {
    isDeleteOverrideLoading.value = true;

    const data = eventOverrides.value.filter(
      (current) =>
        current.key !== override.key && current.override !== override.override
    );

    await saveEventOverrides(data);
    eventOverrides.value = data;
  } finally {
    isDeleteOverrideLoading.value = false;
  }
};

const editOverride = (override) => {
  isEditOverride.value = true;
  editedOverride.value = { ...override };
  newOverride.value = { ...override };
  isAddOverrideOpen.value = true;
};

watch(template, setEventOverrides);
</script>

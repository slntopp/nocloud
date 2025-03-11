<template>
  <div class="pa-5">
    <div class="d-flex align-center">
      <v-card-title>Rules enabled</v-card-title>
      <v-switch v-model="isRulesEnabled"></v-switch>
    </div>

    <v-form ref="suspendRulesForm">
      <div v-for="day in dayOfWeeks" :key="day">
        <v-card-title
          >{{ day }}

          <v-switch
            class="ml-3"
            @change="suspendRules[day].enabled = $event"
            :input-value="suspendRules[day]?.enabled"
            label="Rules enabled"
          />

          <v-switch
            class="ml-3"
            @change="changeEnableAll(day)"
            :input-value="suspendRules[day]?.rules?.length === 0"
            label="Enable all day"
          />
          <v-btn @click="addNewIn(day)" class="ml-2" icon>
            <v-icon>mdi-plus</v-icon>
          </v-btn></v-card-title
        >

        <v-row
          v-for="(rule, index) in suspendRules[day]?.rules || []"
          :key="index"
          class="rule-row"
        >
          <div class="field">
            <span>UTC Start time:</span>
            <v-text-field :rules="timeRule" v-model="rule.startTime" />
          </div>
          <div class="field">
            <span>Local Start time:</span>
            <v-text-field
              disabled
              readonly
              :value="formatTimeFromUtc(rule.startTime)"
            />
          </div>
          <div class="field">
            <span>UTC End time:</span>
            <v-text-field :rules="timeRule" v-model="rule.endTime" />
          </div>
          <div class="field">
            <span>Local End time:</span>
            <v-text-field
              disabled
              readonly
              :value="formatTimeFromUtc(rule.endTime)"
            />
          </div>

          <div class="d-flex justify-center align-center">
            <v-btn @click="deleteRangeFromDay(day, index)" class="mr-2" icon>
              <v-icon>mdi-delete</v-icon>
            </v-btn>
          </div>
        </v-row>
      </div>
    </v-form>

    <div class="d-flex justify-end">
      <v-btn @click="save" :loading="isSaveLoading">Save</v-btn>
    </div>
  </div>
</template>

<script setup>
import { DayOfWeek } from "nocloud-proto/proto/es/services_providers/services_providers_pb";
import { onMounted, ref, toRefs } from "vue";
import { useStore } from "@/store";
import api from "@/api";

const dayOfWeeks = Object.keys(DayOfWeek).filter((v) => !+v && +v !== 0);

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const suspendRules = ref({});
const isRulesEnabled = ref(false);
const isSaveLoading = ref(false);
const suspendRulesForm = ref();

const timeRule = [(v) => !!isTime(v) || "Not valid time"];

onMounted(() => {
  dayOfWeeks.forEach(
    (day) => (suspendRules.value[day] = { rules: [], enabled: true })
  );

  setTemplateRules();
  isRulesEnabled.value = !!template.value?.suspendRules?.enabled;
});

function formatTimeFromUtc(time) {
  if (!isTime(time)) {
    return "-";
  }
  const nowDate = new Date().toUTCString();
  const oldTime = /[0-9]{2}:[0-9]{2}/.exec(nowDate)[0];
  var newDate = new Date(Date.parse(nowDate.replace(oldTime, time)));
  return /[0-9]{2}:[0-9]{2}/.exec(newDate)[0];
}

function isTime(value) {
  return /^([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$/.test(value);
}

const setTemplateRules = () => {
  template.value.suspendRules.schedules.map((shedule) => {
    suspendRules.value[shedule.day] = {
      enabled: shedule.enabled,
      rules: shedule.allowedSuspendTime.map((range) => ({
        startTime: range.startTime,
        endTime: range.endTime,
      })),
    };
  });

  suspendRules.value = { ...suspendRules.value };
};

const addNewIn = (day) => {
  if (!suspendRules.value[day].rules) {
    suspendRules.value[day].rules = [];
  }
  suspendRules.value[day].rules.push({
    endTime: "00:00",
    startTime: "00:00",
  });
  suspendRules.value = { ...suspendRules.value };
};

const deleteRangeFromDay = (day, index) => {
  suspendRules.value[day].rules = suspendRules.value[day].rules.filter(
    (_, i) => i !== index
  );
  suspendRules.value = { ...suspendRules.value };
};

const changeEnableAll = (day) => {
  if (suspendRules.value[day]?.rules.length === 0) {
    addNewIn(day);
  } else {
    suspendRules.value[day].rules = [];
  }
};

const save = async () => {
  if (!suspendRulesForm.value.validate()) {
    store.commit("snackbar/showSnackbarError", {
      message: "Rules not valid!!!",
    });
    return;
  }

  isSaveLoading.value = true;
  try {
    const newSuspendRules = {};
    newSuspendRules.enabled = isRulesEnabled.value;
    newSuspendRules.schedules = [];

    Object.keys(suspendRules.value).map((day) => {
      newSuspendRules.schedules.push({
        day,
        allowedSuspendTime: suspendRules.value[day].rules,
        enabled: suspendRules.value[day].enabled,
      });
    });

    await api.servicesProviders.update(template.value.uuid, {
      ...template.value,
      suspendRules: newSuspendRules,
    });

    store.commit("snackbar/showSnackbarSuccess", {
      message: "Done",
    });
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error while try update suspend rules",
    });
  } finally {
    isSaveLoading.value = false;
  }
};
</script>

<style scoped lang="scss">
.rule-row {
  .field {
    margin: 0px 20px;
    display: flex;
    align-items: center;

    span {
      margin-right: 10px;
    }

    div {
      max-width: 75px;
    }
  }
}
</style>

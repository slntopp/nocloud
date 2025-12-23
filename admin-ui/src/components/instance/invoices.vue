<template>
  <div class="pa-5">
    <v-row align="center">
      <v-col cols="3">
        <date-picker
          label="Next forced renew invoice"
          :value="nextForcedRenewInvoice"
          :clearable="false"
          @input="nextForcedRenewInvoice = $event"
          :disabled="isSaveLoading"
        />
      </v-col>
      <v-col cols="1"
        ><v-btn :loading="isSaveLoading" @click="saveNextForcedRenewInvoice"
          >Save</v-btn
        ></v-col
      >
    </v-row>
    <invoices-table
      table-name="instance-invoices"
      no-search
      hide-select
      :custom-filter="{ instances: [template.uuid] }"
    />
  </div>
</template>

<script setup>
import { onMounted, ref, toRefs } from "vue";
import InvoicesTable from "../invoicesTable.vue";
import DatePicker from "../ui/dateTimePicker.vue";
import {
  formatDateToTimestamp,
  timestampToDateTimeLocal,
} from "../../functions";
import { useStore } from "@/store";
import { UpdateRequest } from "nocloud-proto/proto/es/instances/instances_pb";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store = useStore();

const nextForcedRenewInvoice = ref();
const isSaveLoading = ref(false);

onMounted(() => {
  if (!+template.value?.meta?.nextForcedRenewInvoice) {
    return;
  }
  const nextDate = timestampToDateTimeLocal(
    template.value?.meta?.nextForcedRenewInvoice || 0
  );
  nextForcedRenewInvoice.value = nextDate;
});

const saveNextForcedRenewInvoice = async () => {
  var date = formatDateToTimestamp(new Date(nextForcedRenewInvoice.value)) || 0;
  try {
    isSaveLoading.value = true;

    const data = JSON.parse(JSON.stringify(template.value));

    if (!data.meta) {
      data.meta = {};
    }
    data.meta.next_forced_renew_invoice = date;
    console.log(data.meta, date);

    await store.getters["instances/instancesClient"].update(
      UpdateRequest.fromJson({ instance: data }, { ignoreUnknownFields: true })
    );
  } catch (e) {
    console.log(e);
    store.commit("snackbar/showSnackbarError", {
      message: e.response?.data?.message || "Error during saving date",
    });
  } finally {
    isSaveLoading.value = false;
  }
};
</script>

<script>
export default {
  name: "instanses-invoices",
};
</script>

<style scoped lang="scss"></style>

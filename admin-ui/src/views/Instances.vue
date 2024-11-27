<template>
  <div class="pa-4">
    <v-menu offset-y :close-on-content-click="false">
      <template v-slot:activator="{ on, attrs }">
        <v-btn class="mr-2" color="background-light" v-bind="attrs" v-on="on">
          Create
        </v-btn>
      </template>

      <v-card class="pa-4">
        <v-form ref="form" v-model="newInstance.isValid">
          <accounts-autocomplete
            dense
            label="account"
            style="width: 300px"
            v-model="newInstance.account"
            :rules="rules.req"
          />
          <v-autocomplete
            dense
            label="type"
            style="width: 300px"
            v-model="newInstance.type"
            :items="allTypes"
            :rules="rules.req"
          />
          <v-text-field
            dense
            label="type name"
            v-if="newInstance.type === 'custom'"
            style="width: 300px"
            v-model="newInstance.customTitle"
            :rules="rules.req"
          />

          <v-btn
            :to="{
              name: 'Instance create',
              query: {
                accountId: newInstance.account,
                type:
                  newInstance.type === 'custom'
                    ? newInstance.customTitle
                    : newInstance.type,
              },
            }"
            :disabled="!newInstance.isValid"
          >
            OK
          </v-btn>
        </v-form>
      </v-card>
    </v-menu>

    <confirm-dialog
      :disabled="selected.length < 1"
      @confirm="deleteSelectedInstances"
    >
      <v-btn
        class="mr-2"
        color="background-light"
        :disabled="selected.length < 1"
        :loading="isDeleteLoading"
      >
        Terminate
      </v-btn>
    </confirm-dialog>

    <instances-table v-model="selected" :refetch="refetch" />
  </div>
</template>

<script setup>
import api from "@/api.js";
import confirmDialog from "@/components/confirmDialog.vue";
import instancesTable from "@/components/instancesTable.vue";
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";
import { useStore } from "@/store";
import { ref } from "vue";

const allTypes = ref([]);
const newInstance = ref({
  isValid: false,
  type: "",
  customTitle: "",
  account: "",
});
const isDeleteLoading = ref(false);
const refetch = ref(false);
const rules = ref({
  req: [(v) => !!v || "This field is required!"],
});
const selected = ref([]);

const store = useStore();

const deleteSelectedInstances = async () => {
  const deletePromises = selected.value.map((el) =>
    api.delete(`/instances/${el.uuid}`)
  );
  isDeleteLoading.value = true;

  try {
    const res = await Promise.all(deletePromises);
    if (res.every(({ result }) => result)) {
      const ending = deletePromises.length === 1 ? "" : "s";

      refetch.value = !refetch.value;
      selected.value = [];

      store.commit("snackbar/showSnackbarSuccess", {
        message: `Instance${ending} deleted successfully.`,
      });
    } else {
      store.commit("snackbar/showSnackbarError", {
        message: `Error: ${
          res.response?.data?.message ?? res.message ?? "Unknown"
        }.`,
      });
    }
  } catch (err) {
    if (err.response.status >= 500 || err.response.status < 600) {
      const opts = {
        message: `Service Unavailable: ${
          err.response?.data?.message ?? err.message ?? "Unknown"
        }.`,
        timeout: 0,
      };
      store.commit("snackbar/showSnackbarError", opts);
    } else {
      const opts = {
        message: `Error: ${err.response?.data?.message ?? "Unknown"}.`,
      };
      store.commit("snackbar/showSnackbarError", opts);
    }
  } finally {
    isDeleteLoading.value = false;
  }
};

const types = require.context(
  "@/components/modules/",
  true,
  /instanceCreate\.vue$/
);

types.keys().forEach((key) => {
  const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/instanceCreate\.vue/i);
  if (matched && matched.length > 1) {
    allTypes.value.push(matched[1]);
  }
});
</script>

<script>
export default {
  name: "instances-view",
};
</script>

<style>
.pa-4 .v-icon.group-icon {
  display: none;
  margin: 0 0 2px 4px;
  font-size: 18px;
  opacity: 0.5;
  cursor: pointer;
}
</style>

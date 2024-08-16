<template>
  <div>
    <confirm-dialog
      :disabled="isDeleteDisabled"
      @confirm="tryDeleteSelectedPlans"
    >
      <v-btn
        class="mr-2"
        :disabled="isDeleteDisabled"
        :loading="isDeleteLoading"
      >
        Delete
      </v-btn>
    </confirm-dialog>

    <v-dialog :value="isLinkedInstances" width="60vw">
      <v-card class="confirm-card">
        <v-card-title>
          You can't delete a price model while there are instances using it!
        </v-card-title>
        <v-card-subtitle>
          To delete price model, select the price model that these instances
          will use.
        </v-card-subtitle>
        <nocloud-table
          table-name="linked-plans"
          :show-select="false"
          :items="linked"
          :headers="linkedHeaders"
        >
          <template v-slot:[`item.title`]="{ item }">
            <router-link
              :to="{ name: 'Instance', params: { instanceId: item.uuid } }"
            >
              {{ item.title }}
            </router-link>
          </template>
          <template v-slot:[`item.plan`]="{ item }">
            <plans-auto-complete
              v-model="item.newPlan"
              return-object
              :custom-params="{
                filters: { type: [item.type] },
                showDeleted: false,
                anonymously: true,
              }"
            />
          </template>
        </nocloud-table>

        <div class="d-flex justify-end">
          <v-btn class="mr-3" @click="cancelMoveInstances">Cancel</v-btn>
          <v-btn
            :disabled="linked.some((instance) => !instance.newPlan)"
            @click="deleteSelectedPlansAndMoveLinked"
            >Delete</v-btn
          >
        </div>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import confirmDialog from "@/components/confirmDialog.vue";
import { computed, ref, toRefs } from "vue";
import { useStore } from "@/store";
import nocloudTable from "@/components/table.vue";
import plansAutoComplete from "@/components/ui/plansAutoComplete.vue";

const props = defineProps(["plans"]);
const { plans } = toRefs(props);

const emit = defineEmits(["delete"]);

const store = useStore();

const linkedHeaders = ref([
  { text: "Instance", value: "title" },
  { text: "Price model", value: "plan" },
]);

const linked = ref([]);
const isDeleteLoading = ref(false);

const isDeleteDisabled = computed(() => plans.value.length === 0);

const isLinkedInstances = computed(() => linked.value.length > 0);

const fetchLinked = async () => {
  linked.value = [];

  const instances = await store.dispatch("instances/fetch", {
    filters: {
      billing_plan: plans.value.map((p) => p.uuid),
    },
  });

  linked.value = instances;
};

const tryDeleteSelectedPlans = async () => {
  try {
    isDeleteLoading.value = true;

    await fetchLinked();

    if (linked.value.length > 0) {
      return;
    }

    await deletePlans(plans.value);

    emit("delete");
  } catch (e) {
    console.log(e);

    store.commit("snackbar/showSnackbarError", {
      message: "Error during delete plans",
    });
  } finally {
    isDeleteLoading.value = false;
  }
};

const deleteSelectedPlansAndMoveLinked = async () => {
  try {
    isDeleteLoading.value = true;

    await Promise.all(
      linked.value.map((instance) => {
        return store.getters["instances/instancesClient"].update({
          instance: { uuid: instance.uuid, billingPlan: instance.newPlan },
        });
      })
    );

    await deletePlans(plans.value);
    cancelMoveInstances();
    emit("delete");
  } catch {
    store.commit("snackbar/showSnackbarError", {
      message: "Error during delete plans",
    });
  } finally {
    isDeleteLoading.value = false;
  }
};

const cancelMoveInstances = () => {
  linked.value = [];
};

const deletePlans = (plans) => {
  return Promise.all(
    plans.map((p) =>
      store.getters["plans/plansClient"].deletePlan({ uuid: p.uuid })
    )
  );
};
</script>

<style>
.confirm-card div {
  background-color: var(--v-background-base);
}
.confirm-card {
  background-color: var(--v-background-base) !important;
}
</style>

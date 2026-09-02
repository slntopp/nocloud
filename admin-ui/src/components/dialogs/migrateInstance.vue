<template>
  <instance-control-btn hint="Migrate (PVE)">
    <v-dialog v-model="isModalOpen" max-width="500">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          :disabled="disabled"
          :loading="loading"
          class="ma-1"
          v-bind="attrs"
          v-on="on"
        >
          <v-icon>mdi-server-network</v-icon>
        </v-btn>
      </template>
      <v-card color="background-light">
        <v-card-title>Migrate VM to another node</v-card-title>
        <v-card-text>
          <v-select
            v-model="target"
            :items="nodes"
            label="Target node"
            :hint="`current: ${currentNode || '-'}`"
            persistent-hint
          />
          <v-switch
            v-model="online"
            label="Online (live) migration"
            :disabled="!isRunning"
            :hint="isRunning ? '' : 'VM is not running: offline migration'"
            persistent-hint
          />
        </v-card-text>
        <v-card-actions class="d-flex justify-end">
          <v-btn class="mr-2" @click="isModalOpen = false">Cancel</v-btn>
          <v-btn :disabled="!target" @click="migrate">Migrate</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </instance-control-btn>
</template>

<script setup>
import { toRefs, ref, computed } from "vue";
import { useStore } from "@/store";
import InstanceControlBtn from "@/components/ui/hintBtn.vue";

const props = defineProps(["disabled", "loading", "template"]);
const emit = defineEmits(["click"]);

const store = useStore();
const { template } = toRefs(props);

const isModalOpen = ref(false);
const target = ref("");
const online = ref(true);

const sp = computed(() =>
  store.getters["servicesProviders/all"].find((s) => s.uuid === template.value.sp)
);
const currentNode = computed(
  () => template.value.data?.node || template.value.state?.meta?.node
);
const nodes = computed(() =>
  Object.keys(sp.value?.publicData?.nodes || {}).filter(
    (n) => n !== currentNode.value && sp.value.publicData.nodes[n].state === "ONLINE"
  )
);
const isRunning = computed(() => template.value.state?.state === "RUNNING");

const migrate = () => {
  isModalOpen.value = false;
  emit("click", { target: target.value, online: online.value && isRunning.value });
};
</script>

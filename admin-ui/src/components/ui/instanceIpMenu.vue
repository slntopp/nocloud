<template>
  <v-menu
    bottom
    nudge-top="20"
    nudge-left="15"
    transition="slide-y-transition"
    :close-on-content-click="false"
  >
    <template v-slot:activator="{ on, attrs }">
      <v-text-field
        :label="`IPs_${type}`"
        v-if="props.ui !== 'span'"
        v-bind="attrs"
        v-on="on"
        readonly
        :value="ip"
      >
        <template v-slot:append>
          <v-icon v-if="!edit && ip" @click="addToClipboard(ip)"
            >mdi-content-copy</v-icon
          >
          <v-icon v-else>mdi-swap-horizontal</v-icon>
        </template>
      </v-text-field>
      <span v-else v-bind="attrs" v-on="on">{{ ip }}</span>
    </template>

    <v-list v-if="!edit && item.state.meta?.networking?.[type]?.length" dense>
      <v-list-item v-for="net of item.state.meta.networking?.[type]" :key="net">
        <v-list-item-title>{{ net }}</v-list-item-title>
        <v-icon @click="addToClipboard(net)"> mdi-content-copy</v-icon>
      </v-list-item>
    </v-list>

    <v-list v-else>
      <v-list-item v-for="(net, index) of newIps" :key="net + index">
        <v-text-field :value="net" @input="newIps[index] = $event" dense>
          <template v-slot:append>
            <v-btn icon @click="deleteFromNewIps(index)">
              <v-icon> mdi-delete</v-icon>
            </v-btn>
            <v-btn icon @click="addToClipboard(net)">
              <v-icon> mdi-content-copy</v-icon>
            </v-btn>
          </template>
        </v-text-field>
      </v-list-item>

      <v-list-item
        v-if="newIps.length === 0"
        class="d-flex justify-center align-center"
      >
        <v-list-item-title>Empty</v-list-item-title>
      </v-list-item>

      <v-list-item class="d-flex justify-end">
        <v-btn icon @click="addToNewIps()">
          <v-icon> mdi-plus</v-icon>
        </v-btn>
        <v-btn :loading="isSaveNewIpLoading" icon @click="saveNewIps()">
          <v-icon> mdi-content-save</v-icon>
        </v-btn>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup>
import { computed, defineProps, ref, watch, toRefs } from "vue";
import { addToClipboard } from "@/functions";
import { useStore } from "@/store";

const props = defineProps({
  item: {},
  ui: {},
  type: { type: String, default: "public" },
  edit: { type: Boolean, default: false },
});
const { type, item } = toRefs(props);

const store = useStore();

const newIps = ref([]);
const isSaveNewIpLoading = ref(false);

const ip = computed(
  () =>
    item.value.state.meta?.networking?.[type.value]?.find(
      (ip) =>
        /^(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]\d|\d)){3}$/gm.exec(
          ip
        ) || /\/32$/.exec(ip)
    ) || item.value.state.meta?.networking?.[type.value]?.[0]
);

const setNewIps = () => {
  newIps.value = [];

  item.value.state.meta?.networking?.[type.value]?.forEach((ip) => {
    newIps.value.push(ip);
  });
};

const deleteFromNewIps = (index) => {
  newIps.value = newIps.value.filter((_, ind) => ind !== index);
};

const addToNewIps = () => {
  newIps.value.push("0.0.0.0");
};

const saveNewIps = async () => {
  const data = {};
  const anotherType = type.value === "public" ? "private" : "public";

  data[type.value] = [...newIps.value];
  data[anotherType] = item.value.state.meta?.networking?.[anotherType] || [];

  try {
    isSaveNewIpLoading.value = true;

    await store.dispatch("actions/sendVmAction", {
      action: "change_ip",
      params: data,
      template: item.value,
    });
  } finally {
    isSaveNewIpLoading.value = false;
  }
};

watch(type, setNewIps);

setNewIps();
</script>

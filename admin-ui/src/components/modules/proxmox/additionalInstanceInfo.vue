<template>
  <div>
    <v-row>
      <v-col>
        <instance-ip-menu edit :item="template" />
      </v-col>
      <v-col>
        <instance-ip-menu edit type="private" :item="template" />
      </v-col>
      <v-col>
        <v-select
          append-icon="mdi-pencil"
          :value="template.config.template_id"
          label="OS (template)"
          :items="allOs"
          @input="emit('update', { key: 'config.template_id', value: $event })"
        />
      </v-col>
      <v-col>
        <v-text-field
          @input="emit('update', { key: 'config.username', value: $event })"
          :value="template.config.username"
          label="Username"
          append-icon="mdi-pencil"
        />
      </v-col>
      <v-col>
        <password-text-field
          :readonly="false"
          @input="emit('update', { key: 'config.password', value: $event })"
          :value="template.config.password"
          copy
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-text-field readonly label="PVE state" :value="pveState" />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="VMID"
          :value="template.data?.vmid || template.data?.vm_id"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Node"
          :value="template.data?.node || template.state?.meta?.node"
        />
      </v-col>
      <v-col>
        <v-text-field
          append-icon="mdi-pencil"
          @input="emit('update', { key: 'resources.cpu', value: +$event })"
          :value="template.resources.cpu"
          label="CPU"
        />
      </v-col>
      <v-col>
        <v-text-field
          @input="emit('update', { key: 'resources.ram', value: toGB($event) })"
          append-icon="mdi-pencil"
          :value="getGB(template.resources.ram)"
          label="RAM (GB)"
        />
      </v-col>
      <v-col>
        <v-select
          @input="emit('update', { key: 'resources.drive_type', value: $event })"
          append-icon="mdi-pencil"
          :items="driveTypes"
          :value="template.resources.drive_type"
          label="Disk type"
        />
      </v-col>
      <v-col>
        <v-text-field
          @input="emit('update', { key: 'resources.drive_size', value: toGB($event) })"
          append-icon="mdi-pencil"
          :value="getGB(template.resources.drive_size)"
          label="Disk size (GB)"
        />
      </v-col>
    </v-row>
    <v-row v-if="template.data?.error || template.data?.last_task">
      <v-col v-if="template.data?.error" cols="12">
        <v-alert dense text type="error">{{ template.data.error }}</v-alert>
      </v-col>
      <v-col v-if="template.data?.last_task" cols="12">
        <v-text-field readonly label="Running PVE task" :value="template.data.last_task" />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { toRefs, defineProps, computed } from "vue";
import InstanceIpMenu from "@/components/ui/instanceIpMenu.vue";
import PasswordTextField from "@/components/ui/passwordTextField.vue";

const props = defineProps(["template", "sp"]);
const { template } = toRefs(props);

const emit = defineEmits(["update"]);

const allOs = computed(() => {
  const os = [];
  Object.keys(props.sp?.publicData?.templates || {}).forEach((key) => {
    os.push({ text: props.sp.publicData.templates[key].name, value: +key });
  });
  return os;
});

const pveState = computed(() => {
  const m = template.value.state?.meta || {};
  const parts = [m.state_str || template.value.state?.state];
  if (m.qmpstatus && m.qmpstatus !== m.status) parts.push(`qmp: ${m.qmpstatus}`);
  if (m.lock) parts.push(`lock: ${m.lock}`);
  return parts.filter(Boolean).join(", ");
});

const driveTypes = computed(() => {
  return [
    ...new Set([
      template.value.resources?.drive_type,
      ...(template.value.billingPlan?.resources || [])
        .filter((r) => r.key.startsWith("drive_"))
        .map((r) => r.key.replace("drive_", "").toUpperCase()),
    ]),
  ];
});

const getGB = (value) => ((+value || 0) / 1024).toFixed();
const toGB = (value) => (+value || 0) * 1024;
</script>

<style scoped></style>

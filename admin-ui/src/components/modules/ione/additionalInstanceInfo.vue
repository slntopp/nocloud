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
          label="OS"
          :items="allOs"
          @input="
            emit('update', {
              key: 'config.template_id',
              value: $event,
            })
          "
        />
      </v-col>
      <v-col>
        <v-text-field
          @input="
            emit('update', {
              key: 'config.username',
              value: $event,
            })
          "
          :value="template.config.username"
          label="Username"
          append-icon="mdi-pencil"
        />
      </v-col>
      <v-col>
        <password-text-field
          :readonly="false"
          @input="
            emit('update', {
              key: 'config.password',
              value: $event,
            })
          "
          :value="template.config.password"
          copy
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          label="OpenNebula State"
          :value="template.state.meta?.state_str || template.state.state"
        />
      </v-col>
      <v-col>
        <v-text-field
          append-icon="mdi-pencil"
          @input="
            emit('update', {
              key: 'resources.cpu',
              value: $event,
            })
          "
          :value="template.resources.cpu"
          label="CPU"
        />
      </v-col>
      <v-col>
        <v-text-field
          @input="
            emit('update', {
              key: 'resources.ram',
              value: toGB($event),
            })
          "
          append-icon="mdi-pencil"
          :value="getGB(template.resources.ram)"
          label="RAM (GB)"
        />
      </v-col>
      <v-col>
        <v-select
          @input="
            emit('update', {
              key: 'resources.drive_type',
              value: $event,
            })
          "
          append-icon="mdi-pencil"
          :items="driveTypes"
          :value="template.resources.drive_type"
          label="Disk type"
        />
      </v-col>
      <v-col>
        <v-text-field
          @input="
            emit('update', {
              key: 'resources.drive_size',
              value: toGB($event),
            })
          "
          append-icon="mdi-pencil"
          :value="getGB(template.resources.drive_size)"
          label="Disk size (GB)"
        />
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

const emit = defineEmits("update");

const allOs = computed(() => {
  const os = [];

  Object.keys(props.sp?.publicData.templates || {}).forEach((key) => {
    os.push({ text: props.sp?.publicData.templates[key].name, value: +key });
  });

  return os;
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

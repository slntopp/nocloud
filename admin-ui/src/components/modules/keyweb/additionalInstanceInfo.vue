<template>
  <div>
    <v-row>
      <v-col cols="2">
        <v-text-field
          readonly
          :value="template.data.display_hostname"
          label="Provider API vpsName"
        />
      </v-col>

      <v-col cols="2">
        <instance-ip-menu
          v-if="template.state.meta?.networking?.['public']?.length"
          edit
          :item="template"
        />
        <instance-ip-menu v-else edit :item="template" type="display_ips" />
      </v-col>
      <v-col cols="2">
        <instance-ip-menu edit type="private" :item="template" />
      </v-col>

      <v-col cols="2">
        <v-text-field
          append-icon="mdi-pencil"
          @input="emit('update', { key: 'config.username', value: $event })"
          :value="template.config.username"
          label="Username"
        />
      </v-col>

      <v-col cols="2">
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

      <v-col cols="2">
        <v-text-field label="OS" readonly :value="os" />
      </v-col>

      <v-col cols="2">
        <v-text-field
          append-icon="mdi-pencil"
          @input="emit('update', { key: 'config.hostname', value: $event })"
          :value="template.config.hostname"
          label="Hostname"
        />
      </v-col>

      <v-col cols="2">
        <v-text-field :value="cpu" readonly label="CPU" />
      </v-col>

      <v-col cols="2">
        <v-text-field :value="ram" readonly label="RAM" />
      </v-col>

       <v-col cols="2">
        <v-text-field value="SSD" readonly label="Disk type" />
      </v-col>

      <v-col cols="2">
        <v-text-field :value="disk" readonly label="Disk size" />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import PasswordTextField from "@/components/ui/passwordTextField.vue";
import { toRefs, defineProps, computed } from "vue";
import InstanceIpMenu from "@/components/ui/instanceIpMenu.vue";

const props = defineProps(["template", "sp", "addons"]);
const { template, addons } = toRefs(props);

const emit = defineEmits(["update"]);

const product = computed(() => {
  return (
    template.value.billingPlan.products[template.value.product] || {
      resources: {},
    }
  );
});

const cpu = computed(() => {
  return product.value.resources.cpu;
});

const ram = computed(() => {
  return product.value.resources.ram;
});

const disk = computed(() => {
  return product.value.resources.disk;
});

const os = computed(() => {
  console.log(addons.value);

  return (
    addons.value?.find((a) => a?.meta?.type?.toLowerCase().includes("os"))
      ?.title || "unknown"
  );
});
</script>

<style scoped></style>

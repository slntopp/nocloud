<template>
  <div>
    <v-row>
      <v-col cols="2">
        <instance-ip-menu edit :item="template" />
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
        <v-text-field
          append-icon="mdi-pencil"
          @input="emit('update', { key: 'config.hostname', value: $event })"
          :value="template.config.hostname"
          label="Hostname"
        />
      </v-col>

      <v-col cols="2">
        <v-text-field
          readonly
          :value="template.data.display_hostname"
          label="Display hostname"
        />
      </v-col>

      <v-col cols="2">
        <v-text-field
          readonly
          :value="template.data.display_ips.join(', ')"
          label="Display IPs"
        />
      </v-col>

      <v-col cols="2">
        <v-text-field :value="cpu" readonly label="CPU" />
      </v-col>

      <v-col cols="2">
        <v-text-field :value="ram" readonly label="RAM" />
      </v-col>

      <v-col cols="2">
        <v-text-field :value="disk" readonly label="Disk" />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import PasswordTextField from "@/components/ui/passwordTextField.vue";
import { toRefs, defineProps, computed } from "vue";
const props = defineProps(["template", "sp"]);
const { template } = toRefs(props);
import InstanceIpMenu from "@/components/ui/instanceIpMenu.vue";

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
</script>

<style scoped></style>

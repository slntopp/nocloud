<template>
  <div>
    <v-row>
      <v-col>
        <instance-ip-menu edit :item="template" />
      </v-col>
      <v-col>
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
    </v-row>
  </div>
</template>

<script setup>
import PasswordTextField from "@/components/ui/passwordTextField.vue";
import { toRefs, defineProps } from "vue";
const props = defineProps(["template", "sp"]);
const { template } = toRefs(props);
import InstanceIpMenu from "@/components/ui/instanceIpMenu.vue";

const emit = defineEmits(["update"]);
</script>

<style scoped></style>

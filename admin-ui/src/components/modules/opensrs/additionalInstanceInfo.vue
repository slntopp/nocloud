<template>
  <div>
    <v-row>
      <v-col cols="2">
        <v-text-field
          append-icon="mdi-pencil"
          @input="
            emit('update', { key: 'resources.reg_username', value: $event })
          "
          :value="template.resources.reg_username"
          label="Username"
        />
      </v-col>

      <v-col cols="2">
        <password-text-field
          :readonly="false"
          @input="
            emit('update', {
              key: 'resources.reg_password',
              value: $event,
            })
          "
          :value="template.resources.reg_password"
          copy
        />
      </v-col>

      <v-col cols="2">
        <v-text-field
          append-icon="mdi-pencil"
          @input="emit('update', { key: 'resources.domain', value: $event })"
          :value="template.resources.domain"
          label="Domain"
        />
      </v-col>

      <v-col cols="2">
        <v-text-field
          append-icon="mdi-pencil"
          @input="
            emit('update', { key: 'resources.user.email', value: $event })
          "
          :value="template.resources.user?.email"
          label="Email"
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

const emit = defineEmits(["update"]);
</script>

<style scoped></style>

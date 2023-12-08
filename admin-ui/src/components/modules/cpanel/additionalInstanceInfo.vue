<template>
  <v-row>
    <v-col cols="2">
      <v-text-field
        readonly
        :value="template.config.domain"
        label="Domain"
        @click:append="addToClipboard(template.config.domain)"
        append-icon="mdi-content-copy"
      />
    </v-col>
    <v-col cols="2">
      <v-text-field
        readonly
        :value="template.config.email"
        label="Email"
        @click:append="addToClipboard(template.config.email)"
        append-icon="mdi-content-copy"
      />
    </v-col>
    <v-col cols="2">
      <password-text-field :value="template.config.password" copy />
    </v-col>
  </v-row>
</template>

<script setup>
import { toRefs, defineProps } from "vue";
import PasswordTextField from "@/components/ui/passwordTextField.vue";
import { addToClipboard } from "@/functions";

const props = defineProps(["template", "sp"]);

const { template } = toRefs(props);
</script>

<style scoped></style>

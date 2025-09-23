<template>
  <v-row>
    <v-col cols="3">
      <v-text-field
        :value="template.config.domain"
        label="Domain"
        @input="emit('update', { key: 'config.domain', value: $event })"
      >
        <template v-slot:append>
          <v-icon>mdi-pencil</v-icon>
          <v-icon class="ml-1" @click="addToClipboard(template.config.domain)"
            >mdi-content-copy</v-icon
          >
        </template>
      </v-text-field>
    </v-col>
    <v-col cols="3">
      <v-text-field
        @input="emit('update', { key: 'config.email', value: $event })"
        :value="template.config.email"
        label="Email"
      >
        <template v-slot:append>
          <v-icon>mdi-pencil</v-icon>
          <v-icon class="ml-1" @click="addToClipboard(template.config.email)"
            >mdi-content-copy</v-icon
          >
        </template>
      </v-text-field>
    </v-col>
    <v-col cols="3">
      <password-text-field
        @input="emit('update', { key: 'data.password', value: $event })"
        :value="template.config?.password || template.data?.password"
        copy
        :readonly="false"
      />
    </v-col>

    <v-col cols="3">
      <v-text-field
        append-icon="mdi-pencil"
        @input="emit('update', { key: 'data.username', value: $event })"
        :value="template.config?.username || template.data?.username"
        label="Username"
      />
    </v-col>

    <v-col cols="2">
      <v-text-field
        append-icon="mdi-pencil"
        @input="emit('update', { key: 'resources.bandwidth', value: $event })"
        :value="template.resources.bandwidth"
        label="Bandwidth"
      />
    </v-col>

    <v-col cols="2">
      <v-text-field
        append-icon="mdi-pencil"
        @input="emit('update', { key: 'resources.ssd', value: $event })"
        :value="template.resources.ssd"
        label="SSD (GB)"
      />
    </v-col>

    <v-col cols="2">
      <v-text-field
        append-icon="mdi-pencil"
        @input="emit('update', { key: 'resources.email', value: $event })"
        :value="template.resources.email"
        label="Email"
      />
    </v-col>

    <v-col cols="2">
      <v-text-field
        append-icon="mdi-pencil"
        @input="emit('update', { key: 'resources.mysql', value: $event })"
        :value="template.resources.mysql"
        label="MySQL"
      />
    </v-col>

    <v-col cols="2">
      <v-text-field
        append-icon="mdi-pencil"
        @input="emit('update', { key: 'resources.websites', value: $event })"
        :value="template.resources.websites"
        label="Websites"
      />
    </v-col>
  </v-row>
</template>

<script setup>
import { toRefs, defineProps } from "vue";
import PasswordTextField from "@/components/ui/passwordTextField.vue";
import { addToClipboard } from "@/functions";

const props = defineProps(["template", "sp"]);
const { template } = toRefs(props);

const emit = defineEmits(["update"]);
</script>

<style scoped></style>

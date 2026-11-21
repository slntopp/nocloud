<template>
  <v-expansion-panels v-model="open" variant="accordion" flat class="login">
    <v-expansion-panel value="login">
      <v-expansion-panel-title>Use a different SSH login</v-expansion-panel-title>
      <v-expansion-panel-text>
        <p class="text-caption text-medium-emphasis mb-3">
          For a machine the scan runner's own key no longer opens. The login is encrypted,
          used for this run only and not saved.
        </p>
        <v-text-field
          :model-value="modelValue.user"
          label="user"
          placeholder="root"
          density="compact"
          autocomplete="off"
          @update:model-value="set('user', $event)"
        />
        <v-text-field
          :model-value="modelValue.password"
          label="password"
          type="password"
          density="compact"
          autocomplete="new-password"
          @update:model-value="set('password', $event)"
        />
        <v-textarea
          :model-value="modelValue.key"
          label="private key"
          placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
          rows="3"
          density="compact"
          spellcheck="false"
          hide-details
          @update:model-value="set('key', $event)"
        />
      </v-expansion-panel-text>
    </v-expansion-panel>
  </v-expansion-panels>
</template>

<script setup>
import { ref } from "vue";

// An SSH login for one scan or fix.
const props = defineProps({
  modelValue: { type: Object, required: true },
  expanded: { type: Boolean, default: false },
});
const emit = defineEmits(["update:modelValue"]);

const open = ref(props.expanded ? "login" : undefined);
const set = (field, value) => emit("update:modelValue", { ...props.modelValue, [field]: value });
</script>

<script>
export const emptyLogin = () => ({ user: "", password: "", key: "" });

// The login typed into the form, as the driver takes it; nothing when the
// form was left empty.
export function loginParams(login) {
  const password = login?.password ?? "";
  const key = (login?.key ?? "").trim();
  if (!password && !key) return {};
  return {
    login: {
      user: (login.user ?? "").trim() || "root",
      ...(password && { password }),
      ...(key && { key }),
    },
  };
}

export default { name: "MondooLoginForm" };
</script>

<style scoped>
.login {
  border: 1px solid rgba(128, 128, 128, 0.3);
  border-radius: 6px;
}
</style>

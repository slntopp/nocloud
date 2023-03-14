<template>
  <v-card
    :id="instance.uuid"
    class="mb-4 pa-2"
    elevation="0"
    color="background"
  >
    <v-row>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) => $emit('set-value', { key: 'title', value: newVal })
          "
          label="title"
          :value="instance.title"
        >
        </v-text-field>
      </v-col>
      <v-col v-if="showRemove" class="d-flex justify-end">
        <v-btn @click="() => $emit('remove')"> remove </v-btn>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="6">
        <v-select
          @change="(newVal) => $emit('change-os', newVal)"
          label="template"
          :items="osNames"
          :value="osName"
        >
        </v-select>
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'config.password',
                value: newVal,
              })
          "
          label="password"
          :value="instance.config.password"
        >
        </v-text-field>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.cpu',
                value: +newVal,
              })
          "
          label="cpu"
          :value="instance.resources.cpu"
          type="number"
        >
        </v-text-field>
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.ram',
                value: +newVal,
              })
          "
          label="ram"
          :value="instance.resources.ram"
          type="number"
        >
        </v-text-field>
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', { key: 'resources.drive_type', value: newVal })
          "
          label="drive type"
          :value="instance.resources.drive_type"
        >
        </v-text-field>
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.drive_size',
                value: +newVal,
              })
          "
          label="drive size"
          :value="instance.resources.drive_size"
          type="number"
        >
        </v-text-field>
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.ips_public',
                value: +newVal,
              })
          "
          label="ips public"
          :value="instance.resources.ips_public"
          type="number"
        >
        </v-text-field>
      </v-col>
      <v-col cols="6">
        <v-text-field
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'resources.ips_private',
                value: +newVal,
              })
          "
          label="ips private"
          :value="instance.resources.ips_private"
          type="number"
        >
        </v-text-field>
      </v-col>
      <v-col cols="6">
        <v-select
          label="price model"
          item-text="title"
          item-value="uuid"
          :value="instance.plan"
          :items="plans.list"
          :rules="planRules"
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'billing_plan',
                value: newVal,
              })
          "
        />
      </v-col>
      <v-col cols="6">
        <v-select
          label="product"
          :value="instance.productTitle"
          v-if="plansProducts.length > 0"
          :items="plansProducts"
          @change="
            (newVal) => $emit('set-value', { key: 'product', value: newVal })
          "
        />
      </v-col>
    </v-row>
  </v-card>
</template>

<script>
export default {
  name: "instanceCreateCard",
  props: {
    instance: {},
    "os-names": {},
    "os-name": {},
    "plan-rules": {},
    plans: {},
    "plans-products": {},
    "show-remove": { type: Boolean, default: true },
  },
  emits: ["set-value", "change-os", "remove"],
};
</script>

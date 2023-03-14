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
          label="price model"
          item-text="title"
          item-value="uuid"
          :value="instance.plan"
          :items="plans.list"
          :rules="planRules"
          @change="
            (newVal) =>
              $emit('set-value', { key: 'billing_plan', value: newVal })
          "
        />
      </v-col>
      <v-col cols="6" v-if="instance.products?.length > 0">
        <v-select
          label="product"
          :value="instance.productTitle"
          :items="instance.products"
          @change="
            (newVal) => $emit('set-value', { key: 'product', value: newVal })
          "
        />
      </v-col>
      <v-col cols="6">
        <v-select
          label="tariff"
          item-text="title"
          item-value="code"
          :value="instance.config.planCode"
          :items="flavors[instance.billing_plan.uuid]"
          :rules="rules.req"
          :loading="isFlavorsLoading"
          @change="
            (newVal) =>
              $emit('set-value', { key: 'config.planCode', value: newVal })
          "
        />
      </v-col>
      <v-col cols="6">
        <v-select
          label="region"
          :value="instance.config.configuration.vps_datacenter"
          :items="regions[instance.config.planCode]"
          :rules="rules.req"
          :disabled="!instance.config.planCode"
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'config.configuration.vps_datacenter',
                value: newVal,
              })
          "
        />
      </v-col>
      <v-col cols="6">
        <v-select
          label="OS"
          :value="instance.config.configuration.vps_os"
          :items="images[instance.config.planCode]"
          :rules="rules.req"
          :disabled="!instance.config.planCode"
          @change="
            (newVal) =>
              $emit('set-value', {
                key: 'config.configuration.vps_os',
                value: newVal,
              })
          "
        />
      </v-col>
      <v-col cols="6" class="d-flex align-center">
        Payment:
        <v-switch
          class="d-inline-block ml-2"
          true-value="P1Y"
          false-value="P1M"
          :value="instance.config.duration"
          :label="instance.config.duration === 'P1Y' ? 'yearly' : 'monthly'"
          @change="
            (newVal) =>
              $emit('set-value', { key: 'config.duration', value: newVal })
          "
        />
      </v-col>
      <v-col cols="6" class="d-flex align-center">
        Existing:
        <v-switch
          class="d-inline-block ml-2"
          :value="instance.data.existing"
          @change="
            (newVal) =>
              $emit('set-value', { key: 'data.existing', value: newVal })
          "
        />
      </v-col>
      <v-col cols="6" class="d-flex align-center" v-if="instance.data.existing">
        <v-text-field
          label="VPS name"
          :value="instance.data.vpsName"
          @change="
            (newVal) =>
              $emit('set-value', { key: 'data.vpsName', value: newVal })
          "
          :rules="rules.req"
        />
      </v-col>
    </v-row>
    <template
      v-if="Object.values(addons[instance.config.planCode] || {}).length > 0"
    >
      <v-card-title class="px-0 pb-0">Addons:</v-card-title>
      <v-row>
        <v-col
          cols="6"
          v-for="(addon, key) in addons[instance.config.planCode]"
          :key="key"
        >
          <v-select
            :label="key"
            :items="addon"
            @change="
              (value) => $emit('set-value', { key: 'config.addons', value })
            "
          />
        </v-col>
      </v-row>
    </template>
  </v-card>
</template>

<script>
export default {
  name: "instanceCreateCard",
  props: {
    instance: {},
    "show-remove": { type: Boolean, default: true },
    "plan-rules": {},
    plans: {},
    flavors: {},
    images: {},
    regions: {},
    addons: {},
    "is-flavors-loading": {},
  },
  emits: ["set-value", "remove"],
  data: () => ({
    rules: {
      req: [(v) => !!v || "required field"],
    },
  }),
};
</script>

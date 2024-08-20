<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          :label="`Provider API ${ovhType}Name`"
          :value="template.data?.[ovhType + 'Name']"
          @click:append="addToClipboard(template.data?.[ovhType + 'Name'])"
          append-icon="mdi-content-copy"
        />
      </v-col>

      <v-col>
        <instance-ip-menu :item="template" />
      </v-col>

      <v-col>
        <v-text-field
          readonly
          label="Login"
          :value="template.state.meta?.login"
          @click:append="addToClipboard(template.state.meta?.login)"
          append-icon="mdi-content-copy"
        />
      </v-col>

      <v-col>
        <password-text-field
          readonly
          :value="template.state.meta?.password"
          copy
        />
      </v-col>

      <v-col>
        <v-text-field
          append-icon="mdi-pencil"
          @input="
            emit('update', {
              key: `config.configuration.${ovhType}_os`,
              value: $event,
            })
          "
          :value="os"
          label="OS"
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-text-field
          append-icon="mdi-pencil"
          @input="
            emit('update', {
              key: `config.configuration.${ovhType}_datacenter`,
              value: $event,
            })
          "
          :value="template.config.configuration?.[`${ovhType}_datacenter`]"
          label="Datacenter"
        />
      </v-col>

      <v-col>
        <v-text-field
          append-icon="mdi-pencil"
          @input="
            emit('update', {
              key: `resources.cpu`,
              value: $event,
            })
          "
          :value="cpu"
          label="CPU"
        />
      </v-col>

      <v-col>
        <v-text-field
          append-icon="mdi-pencil"
          @input="
            emit('update', {
              key: `resources.ram`,
              value: $event,
            })
          "
          :value="template.resources.ram"
          label="RAM"
        />
      </v-col>

      <v-col>
        <v-text-field
          append-icon="mdi-pencil"
          @input="
            emit('update', {
              key: `resources.drive_type`,
              value: $event,
            })
          "
          :value="template.resources.drive_type"
          label="Disk type"
        />
      </v-col>

      <v-col>
        <v-text-field
          append-icon="mdi-pencil"
          @input="
            emit('update', {
              key: `resources.drive_size`,
              value: $event,
            })
          "
          :value="template.resources.drive_size"
          label="Disk size"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { toRefs, defineProps, computed } from "vue";
import InstanceIpMenu from "@/components/ui/instanceIpMenu.vue";
import { addToClipboard } from "@/functions";
import PasswordTextField from "@/components/ui/passwordTextField.vue";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const emit = defineEmits(["update"]);

const os = computed(() => {
  const os =
    template.value.config.configuration[
      Object.keys(template.value.config.configuration).find((k) =>
        k.includes("os")
      )
    ];

  if (template.value.config.type === "cloud") {
    return template.value.billingPlan.products[
      template.value.product
    ].meta.os.find(({ id }) => id === os).name;
  }

  return os;
});

const cpu = computed(() => {
  if (ovhType.value === "dedicated") {
    return template.value.billingPlan.products[template.value.product]?.meta
      .cpu;
  }

  return template.value.resources.cpu;
});

const ovhType = computed(() => {
  return template.value.config.type;
});
</script>

<style scoped></style>

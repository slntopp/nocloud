<template>
  <div>
    <v-row>
      <v-col>
        <v-text-field
          readonly
          :label="`Provider API ${ovhType}Name`"
          :value="template.data[ovhType + 'Name']"
          @click:append="addToClipboard(template.data[ovhType + 'Name'])" append-icon="mdi-content-copy"
        />
      </v-col>
      <v-col>
        <instance-ip-menu :item="template" />
      </v-col>
      <v-col>
        <v-text-field readonly :value="os" label="OS"/>
      </v-col>
      <v-col>
        <v-text-field
          readonly
          label="Login"
          :value="template.state.meta.login"
          @click:append="addToClipboard(template.data[ovhType + 'Name'])" append-icon="mdi-content-copy"
        />
      </v-col>
      <v-col>
        <password-text-field
          readonly
          :value="template.state.meta.password"
          copy
        />
      </v-col>
    </v-row>
    <v-row>
      <v-col>
        <v-text-field readonly :value="cpu" label="CPU" />
      </v-col>
      <v-col>
        <v-text-field readonly :value="template.resources.ram" label="RAM" />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.resources.drive_type"
          label="Disk type"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.resources.drive_size"
          label="Disk size"
        />
      </v-col>
      <v-col>
        <v-text-field
          readonly
          :value="template.resources.cpu"
          label="Software"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script setup>
import { toRefs, defineProps, ref, computed, onMounted } from "vue";
import InstanceIpMenu from "@/components/ui/instanceIpMenu.vue";
import api from "@/api";
import {addToClipboard} from "@/functions";
import PasswordTextField from "@/components/ui/passwordTextField.vue";

const props = defineProps(["template"]);

const { template } = toRefs(props);

const os = computed(() => {
  return template.value.config.configuration[
    Object.keys(template.value.config.configuration).find((k) =>
      k.includes("os")
    )
  ];
});

const ovhType = computed(() => {
  return template.value.config.type;
});

const cpu = ref("");

onMounted(async () => {
  cpu.value = template.value.resources.cpu;
  if (template.value.config.type === "dedicated") {
    const { meta } = await api.servicesProviders.action({
      uuid: template.value.sp,
      action: "checkout_baremetal",
      params: JSON.parse(JSON.stringify(template.value.config)),
    });

    meta.order.details.forEach((d) => {
      if (
        d.description.toLowerCase().includes("intel") ||
        d.description.toLowerCase().includes("amd")
      ) {
        cpu.value = d.description;
      }
    });
  }
});
</script>

<style scoped></style>

<template>
    <v-row>
      <v-col cols="3">
        <route-text-field
            :to="{ name: 'Account', params: { accountId: user?.uuid } }"
            label="User"
            :value="user?.title"
        />
      </v-col>
      <v-col>
      </v-col>
    </v-row>
</template>

<script setup>
import { toRefs, defineProps, computed } from "vue";
import {useStore} from "@/store";
import RouteTextField from "@/components/ui/routeTextField.vue";

const props = defineProps(["template"]);
const { template } = toRefs(props);

const store=useStore()

const user=computed(()=>store.getters['accounts/all'].find(a=>a.uuid===template.value.config.user))
</script>

<style scoped></style>

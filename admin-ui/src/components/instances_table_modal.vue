<template>
  <v-dialog max-width="1280" :value="visible" @input="$emit('close')">
    <template v-slot:activator="{ on, attrs }">
      <slot v-on="on" v-bind="attrs" name="activator" />
    </template>
    <v-card elevation="0" color="background" class="pa-4">
      <instances-table
        class="mt-0"
        :value="null"
        :headers="headers"
        :items="accountInstances"
        :show-select="false"
      />
    </v-card>
  </v-dialog>
</template>

<script setup>
import { computed } from 'vue'
import store from '@/store'
import instancesTable from '@/components/instances_table.vue'

const props = defineProps({
  uuid: { type: String, required: true },
  visible: { type: Boolean, required: true }
})

const headers = [
  { text: "Title", value: "title" },
  { text: "Due date", value: "dueDate" },
  { text: "Status", value: "state" },
  { text: "Tariff", value: "product" },
  { text: "Price", value: "accountPrice" },
]

const namespaces = computed(() =>
  store.getters['namespaces/all'] ?? []
)

const instances = computed(() =>
  store.getters['services/getInstances'] ?? []
)

const accountInstances = computed(() => {
  const namespace = namespaces.value.find(({ access }) => 
    access.namespace === props.uuid
  )

  if (!namespace) return []
  return instances.value.filter(({ access }) =>
    access.namespace === namespace.uuid
  )
})

store.dispatch('services/fetch')
store.dispatch('namespaces/fetch')
</script>

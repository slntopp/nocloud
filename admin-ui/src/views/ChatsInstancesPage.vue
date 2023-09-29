<template>
  <v-card elevation="0" color="background" class="pa-4">
    <instances-table
      :value="null"
      :items="accountInstances"
      :show-select="false"
    />
  </v-card>
</template>

<script setup>
import { computed } from 'vue'
import store from '@/store'
import router from '@/router'
import instancesTable from '@/components/instances_table.vue'

const namespaces = computed(() =>
  store.getters['namespaces/all'] ?? []
)

const instances = computed(() =>
  store.getters['services/getInstances'] ?? []
)

const accountInstances = computed(() => {
  const namespace = namespaces.value.find(({ access }) => 
    access.namespace === router.currentRoute.params.uuid
  )

  if (!namespace) return []
  return instances.value.filter(({ access }) =>
    access.namespace === namespace.uuid
  )
})

store.dispatch('services/fetch')
store.dispatch('namespaces/fetch')
</script>

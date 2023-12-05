<template>
  <component
    offset-y
    max-width="1280"
    :is="(type === 'dialog') ? VDialog : VMenu"
    :value="visible"
    :close-on-content-click="false"
    @input="emits('close')"
  >
    <template v-slot:activator="{ on, attrs }">
      <slot v-on="on" v-bind="attrs" name="activator" />
    </template>

    <v-card
      class="pa-4"
      elevation="0"
      color="background"
      @mouseenter="emits('hover')"
      @mouseleave="emits('close')"
    >
      <instances-table
        open-in-new-tab
        style="margin-top: 0 !important"
        :value="[]"
        :headers="headers"
        :items="accountInstances"
        :show-select="false"
      />
    </v-card>
  </component>
</template>

<script setup>
import { computed } from 'vue'
import { VDialog, VMenu } from 'vuetify/lib/components'
import store from '@/store'
import instancesTable from '@/components/instances_table.vue'

const props = defineProps({
  uuid: { type: String, required: true },
  visible: { type: Boolean, required: true },
  type: { type: String, default: 'dialog' }
})
const emits = defineEmits(['hover', 'close'])

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

store.dispatch('services/fetch',{showDeleted:true})
store.dispatch('namespaces/fetch')
</script>

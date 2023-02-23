<template>
    <v-card :loading="isFetchLoading" elevation="0" color="background-light" class="pa-4">
        <v-text-field class="mt-5" :value="namespace?.title" @input="$emit('input:title', $event)" label="title"
            style="width: 330px" />
        <v-card-title>Access</v-card-title>
        <v-text-field readonly :value="namespace?.access?.level" label="level" style="width: 330px" />
        <v-text-field readonly :value="namespace?.access?.role" label="role" style="width: 330px" />
        <v-text-field readonly :value="namespace?.access?.namespace" label="namespace" style="width: 330px" />

        <div class="pt-4">
            <v-btn class="mt-4 mr-2" :loading="isEditLoading" @click="editNamespace">
                Submit
            </v-btn>
        </div>
    </v-card>
</template>
  
<script>
import api from '@/api.js'

export default {
    name: "namespace-info",
    data: () => ({ namespaceTitle: '...', isFetchLoading: false, isEditLoading: false }),
    props: ['namespace'],
    methods: {
        editNamespace() {
            this.isEditLoading = true

            api.namespaces.edit(this.namespace).finally(() => {
                this.isEditLoading = false
            })
        }
    },
    computed: {
        all() {
            return this.$store.getters['namespaces/all']
        },
    },
};
</script>
  

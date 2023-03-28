<template>
    <v-card :loading="loading" elevation="0" color="background-light" class="pa-4">
        <v-text-field class="mt-5" v-model="newTitle" label="title" style="width: 330px" />
        <v-card-title>Access</v-card-title>
        <v-text-field readonly :value="namespace?.access?.level" label="level" style="width: 330px" />
        <v-text-field readonly :value="namespace?.access?.role" label="role" style="width: 330px" />
        <v-text-field readonly :value="namespace?.access?.namespace" label="namespace" style="width: 330px" />

        <div class="pt-4">
            <v-btn class="mt-4 mr-2" :loading="isEditLoading" @click="editNamespace">
                Save
            </v-btn>
        </div>
    </v-card>
</template>
  
<script>
import api from '@/api.js'

export default {
    name: "namespace-info",
    data: () => ({ isEditLoading: false, newTitle: '' }),
    props: ['namespace', 'loading'],
    methods: {
        editNamespace() {
            this.isEditLoading = true

            api.namespaces.edit({ ...this.namespace, title: this.newTitle }).then(() => {
                this.$emit('input:title', this.newTitle)
            }).catch(e => console.log(e)).finally(() => {
                this.isEditLoading = false
            })
        }
    },
    computed: {
        all() {
            return this.$store.getters['namespaces/all']
        },
    },
    watch: {
        'namespace.title'(newVal) {
            this.newTitle = newVal
        }
    }
};
</script>
  

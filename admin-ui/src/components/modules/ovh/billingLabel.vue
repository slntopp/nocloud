<template>
<v-container>
    <v-row>
        <v-col>
            <v-chip outlined color="primary" >Due to date: {{dueDate}}</v-chip>
        </v-col>
        <v-col>
            <v-chip color="primary" outlined>Price: {{price}}</v-chip>
        </v-col>
    </v-row>
</v-container>
</template>

<script setup>
import {computed, defineProps} from 'vue'
import {getOvhPrice} from "@/functions";

const props=defineProps(['template'])

const dueDate=computed(()=>{
    return props.template.data.expiration
})

const price=computed(()=>{
    return getOvhPrice(props.template)
})
</script>

<style scoped>

</style>
<template>
    <div class="settings pa-10">
        <v-row justify="center">
            <v-card-title>Map settings</v-card-title>
        </v-row>
        <v-card-title>Rewards:</v-card-title>
        <v-row align="center">
            <v-col cols="2" v-for="reward in rewards" :key="reward.id">
                <v-card class="mx-auto my-12" color="background-light">
                    <v-img cover :src="reward.img" />
                    <v-card-title><span>{{ reward.description }}</span></v-card-title>
                    <div class="d-flex pa-5 flex-row-reverse">
                        <v-btn color="primary">Delete</v-btn>
                    </div>
                </v-card>
            </v-col>
            <v-col cols="1" class="ma">
                <v-card color="background-light" class="d-flex justify-center align-center" outlined>
                    <v-btn flat color="background-light" class="pa-5">
                        <v-icon size="64">mdi-plus</v-icon>
                    </v-btn>
                </v-card>
            </v-col>
        </v-row>
    </div>
</template>

<script>
export default {
    data: () => ({
        rewards: [{ id: '12', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description" }]
    })
}
</script>

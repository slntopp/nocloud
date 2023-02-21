<template>
    <div class="settings pa-10">
        <v-row justify="center">
            <v-card-title>Map settings</v-card-title>
        </v-row>
        <v-card-title>Rewards:</v-card-title>
        <v-row>
            <v-col heigh="100%" xs="4" md="3" lg="2" xl="2" v-for="reward in rewards" :key="reward.id">
                <v-card height="100%" color="background-light">
                    <v-img cover :src="reward.img" />
                    <v-divider />
                    <v-card-title><span>{{ reward.description }}</span></v-card-title>
                    <div class="d-flex pa-5  flex-row-reverse">
                        <v-btn color="primary" @click="deleteReward(reward.id)">Delete</v-btn>
                    </div>
                </v-card>
            </v-col>
            <v-col class="d-flex justify-center align-center" cols="1">
                <v-dialog v-model="addRewardDialog" persistent>
                    <template v-slot:activator="{ on, attrs }">
                        <v-btn v-on="on" v-bind="attrs" block width="75" height="75" color="background-light">
                            <v-icon size="64">mdi-plus</v-icon>
                        </v-btn>
                    </template>
                    <v-card color="background-light" class="pa-5 ma-auto" max-width="600">
                        <v-card-title class="text-h5">
                            Add new reward:
                        </v-card-title>
                        <v-file-input accept="image/*" v-model="newReward.file" clearable label="File input" underlined />
                        <v-text-field class="ml-8" v-model="newReward.description" />
                        <v-card-actions class="flex-row-reverse">
                            <v-btn class="mx-5" color="red" @click="closeAddReward">
                                close
                            </v-btn>
                            <v-btn class="mx-5" color="primary" @click="addReward">
                                add
                            </v-btn>
                        </v-card-actions>
                    </v-card>
                </v-dialog>
            </v-col>
        </v-row>
    </div>
</template>

<script>
export default {
    data: () => ({

        addRewardDialog: false,
        newReward: { file: null, description: '' },

        rewards: [
            { id: '1', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description1" },
            { id: '2', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description2" },
            { id: '3', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description3" },
            { id: '4', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description4" },
            { id: '5', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description5" },
            { id: '6', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description6" },
            { id: '7', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description7" },
            { id: '8', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", },
            { id: '912', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description9" },
            { id: '212', img: "https://cdn.vuetifyjs.com/images/cards/cooking.png", description: "some description10" },
        ]
    }),
    methods: {
        deleteReward(id) {
            console.log(id);
        },
        addReward() {
            console.log(this.newReward);
        },
        closeAddReward() {
            this.newReward = { file: null, description: '' }
            this.addRewardDialog = false
        }
    }
}
</script>

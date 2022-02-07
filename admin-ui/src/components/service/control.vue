<template>
	<v-card
		elevation="0"
		color="background-light"
		class="pa-4"
	>
		service state: <v-chip
			x-small
			:color="chipColor"
		>
			{{this.service.status}}
		</v-chip>
		<div class="mt-4 mb-2">control:</div>

		<div class="control">
			<template v-if="service.status == 'up' || service.status == 'del'">						
				<v-btn
					:loading="loading.action"
					@click="down"
				>
					down service
				</v-btn>
			</template>
			<template v-else>
				deploy:
				<v-form ref="deployForm">
					<v-row>
						<v-col>
							<v-select
								label="instance group"
								:items="instancesGroups"
								:rules="[v=>!!v || 'required']"
								v-model="deployInstancesGroup"
							>
							</v-select>
						</v-col>
					</v-row>
					<v-row>
						<v-col>
							<v-select
								label="services provider"
								:items="servicesProviders"
								item-value="uuid"
								item-text="title"
								:rules="[v=>!!v || 'required']"
								v-model="deployServiceProvider"
							>
							</v-select>
						</v-col>
					</v-row>
					<v-row>
						<v-col>
							<v-btn
								:loading="loading.action"
								:disabled="!deployServiceProvider || !deployInstancesGroup"
								@click="deploy"
							>
								deploy
							</v-btn>
						</v-col>
					</v-row>
				</v-form>
			</template>
		</div>
	</v-card>
</template>

<script>
import api from "@/api"

export default {
	name: "service-control",
	data: () => ({
		deployServiceProvider: '',
		deployInstancesGroup: '',
		loading: {
			action: false
		}
	}),
	props: {
		service: {
			type: Object,
			required: true
		},
		'chip-color': {
			type: String,
			required: true
		}
	},
	methods: {
		deploy(){
			if(!this.$refs.deployForm.validate())
				return
			this.loading.action = true;				
				
			api.services.up(this.serviceId, this.deployInstancesGroup, this.deployServiceProvider)
			.then(() => {
				this.$store.dispatch('services/fetch')
			})
			.finally(() => {
				this.loading.action = false
			})
		},
		down(){
			this.loading.action = true;			
			api.post(`/services/${this.serviceId}/down`, {})
			.then(() => {
				this.$store.dispatch('services/fetch')
			})
			.finally(() => {
				this.loading.action = false
			})
		}
	},
	computed: {
		instancesGroups(){
			const result = []
			this.service.instancesGroups

			for (const group in this.service.instancesGroups) {
				result.push({text: group, value: this.service.instancesGroups[group].uuid})
			}
			return result
		},
		servicesProviders(){
			return this.$store.getters['servicesProviders/all'];
		},
		serviceId(){
			return this.$route.params.serviceId;
		},
	},
	mounted(){
		if (this.service.status != 'up' && this.service.status != 'del') {
			this.$store.dispatch('servicesProviders/fetch')
		}
	}
}
</script>

<style>

</style>
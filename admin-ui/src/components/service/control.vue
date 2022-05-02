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

		<div v-if="this.service.status == 'up'">VM control:</div>
		<v-row>
			<v-col>
				<v-btn
					v-for="(btn, index) in vmControlBtns" :key="btn.action"
					@click="sendVmAction(btn.action)"
					:class="{ 'mr-2': index !== vmControlBtns.lenght - 1 }"
					:disabled="actionLoading && actualAction != btn.action"
					:loading="actionLoading && actualAction == btn.action"
				>
					{{btn.title || btn.action}}
				</v-btn>
			</v-col>
		</v-row>

		<div class="mt-4 mb-2">service control:</div>

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

		<v-snackbar
			v-model="snackbar.visibility"
			:timeout="snackbar.timeout"
			:color="snackbar.color"
		>
			{{snackbar.message}}
			<template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
				<router-link :to="snackbar.route">
					Look up.
				</router-link>
			</template>
			

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
		</v-snackbar>
	</v-card>
</template>

<script>
import api from "@/api"
import snackbar from "@/mixins/snackbar.js"

export default {
	name: "service-control",
	mixins: [snackbar],
	data: () => ({
		deployServiceProvider: '',
		deployInstancesGroup: '',
		loading: {
			action: false
		},
		vmState: null,
		actualAction: '',
		actionLoading: false,

		vmControlBtns: [
			{
				action: "poweroff",
				title: "poweroff", //not reqired, use 'action' for a name if not found
			},
			{
				action: "resume",
			},
			{
				action: "suspend",
			},
			{
				action: "reboot",
			},
		]
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
			api.services.down(this.serviceId)
			.then(() => {
				this.$store.dispatch('services/fetch')
			})
			.finally(() => {
				this.loading.action = false
			})
		},
		sendVmAction(action){
			// console.log(action, this.service)
			const groupName = Object.keys(this.service.instancesGroups)[0];
			let vminfo = {
				service: this.service.uuid,
				group: groupName,
				instance: this.service.instancesGroups[groupName].instances[0].uuid,
			}
			// console.log(vminfo)
			this.actualAction = action;
			this.actionLoading = true;
			api.services.action(vminfo, action)
			.then(() => {
				this.showSnackbarSuccess({message: `Done!`})
			})
			.catch(err => {
				const opts = {
					message: `Error: ${err?.response?.data?.message ?? 'Unknown'}.`,
				}
				this.showSnackbarError(opts);
			})
			.finally(() => {
				this.actualAction = "";
				this.actionLoading = false;
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
		api.get(`/services/${this.serviceId}/states`)
		.then(res => {
			console.log(res)
		})
		api.services.getStates(this.serviceId)
		.then(res => {
			console.log(res)
		})
	}
}
</script>

<style>

</style>
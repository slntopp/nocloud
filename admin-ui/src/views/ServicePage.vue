<template>
	<div class="service pa-4">
		<div class="page__title">
			Service: {{ service.title }}
			<v-chip
				small
				:color="chipColor"
			>
				{{ service.status }}
			</v-chip>
		</div>

		<v-row>
			<v-col
				lg=7
				md=6
				sm=12
				cols=12
			>
				<template v-if="service.status == 'up' || service.status == 'del'">
					service deployed
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
								<v-btn :disabled="!deployServiceProvider || !deployInstancesGroup" @click="deploy">deploy</v-btn>
							</v-col>
						</v-row>
					</v-form>
				</template>
			</v-col>
			<v-col
				lg=5
				md=6
				sm=12
				cols=12
			>
				Service JSON:
				<pre v-html="syntaxHightlight(JSON.stringify(service, null, 2))"></pre>
			</v-col>
		</v-row>
	</div>
</template>

<script>
import api from "@/api"

export default {
	name: 'service-view',
	data: () => ({
		notFound: false,
		deployServiceProvider: '',
		deployInstancesGroup: '',
	}),
	methods: {
		syntaxHightlight(json){
			json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
			return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+-]?\d+)?)/g, function (match) {
				let cls = 'number';
				if (/^"/.test(match)) {
					if (/:$/.test(match)) {
						cls = 'key';
					} else {
						cls = 'string';
					}
				} else if (/true|false/.test(match)) {
					cls = 'boolean';
				} else if (/null/.test(match)) {
					cls = 'null';
				}
				return '<span class="' + cls + '">' + match + '</span>';
			});
		},
		deploy(){
			if(!this.$refs.deployForm.validate())
				return
				
			api.services.up(this.serviceId, this.deployInstancesGroup, this.deployServiceProvider)
			.then(() => {
				this.$store.dispatch('services/fetch')
			})
		}
	},
	computed: {
		service(){
			const items = this.$store.getters['services/all']
			const item = items.find(el => el.uuid == this.serviceId)
	
			if(item)
				return item
	
			
			return {}
		},
		serviceId(){
			return this.$route.params.serviceId;
		},
		chipColor(){
			const dict = {
				'init': 'orange darken-2',
				'up': 'green darken-2',
				'del': 'gray darken-2'
			}
			return dict[this.service.status] ?? 'blue-grey darken-2'
		},
		servicesProviders(){
			return this.$store.getters['servicesProviders/all'];
		},
		instancesGroups(){
			const result = []
			this.service.instancesGroups

			for (const group in this.service.instancesGroups) {
				result.push({text: group, value: this.service.instancesGroups[group].uuid})
			}
			return result
		}
	},
	created(){
		this.$store.dispatch('services/fetch')
		.then(() => {
			this.notFound = true
		})
	},
	mounted(){
		document.title = `${this.service.title} | NoCloud`

		if (this.service.status != 'up' && this.service.status != 'del') {
			this.$store.dispatch('servicesProviders/fetch')
		}
	}
}
</script>

<style scoped lang="scss">
.page__title{
	color: var(--v-primary-base);
	font-weight: 400;
	font-size: 32px;
	font-family: "Quicksand", sans-serif;
	line-height: 1em;
	margin-bottom: 10px;
}
</style>

<style>
pre {
	padding: 5px;
	margin: 5px;
	background-color: var(--v-background-light-base);
	border-radius: 4px;
}
.string {
	color: green; 
}
.number {
	color: darkorange; 
}
.boolean {
	color: blue; 
}
.null {
	color: magenta; 
}
.key {
	color: red; 
}
</style>
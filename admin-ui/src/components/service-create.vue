<template>
	<v-form v-model='formValid'>
		<v-row
			justify="start"
		>
			<v-col
				cols="6"
				md="4"
				lg="3"
			>
				<v-text-field
					label="service title"
					:rules="rules.req"
					v-model="service.title"
				></v-text-field>
			</v-col>
			<v-col
				cols="6"
				md="4"
				lg="3"
			>
				<v-text-field
					label="service title"
					:rules="rules.req"
					v-model="service.version"
					readonly
				></v-text-field>
			</v-col>
		</v-row>

		<v-row>
			<v-col
				cols="4"
				md="4"
				lg="3"
			>
				<v-list
					dence
					color="background-light"
				>
					<v-subheader>instances {{currentInstancesGroupsIndex}}</v-subheader>

					<v-list-item
						v-for="(instance, index) in instances"
						:key="index"
						@click="() => selectIntance(index)"

						:class="{ 'v-list-item--active': index == currentInstancesGroupsIndex }"
					>
						<v-list-item-icon>
							<v-icon>mdi-playlist-star</v-icon>
						</v-list-item-icon>
						<v-list-item-content>
							<v-list-item-title>{{instance.title || '(Enter title)'}}</v-list-item-title>
						</v-list-item-content>
					</v-list-item>

					<v-list-item @click="() => addInstance()">
						<v-list-item-icon>
							<v-icon>mdi-plus-circle-outline</v-icon>
						</v-list-item-icon>
						<v-list-item-content>
							<v-list-item-title>add instance</v-list-item-title>
						</v-list-item-content>
					</v-list-item>
				</v-list>
			</v-col>
			<v-col
				cols="8"
				md="8"
				lg="9"
			>
				<v-card
					v-if="currentInstancesGroupsIndex != -1"
					color="background-light"
					elevation="0"
					class="pa-4"
					:key="currentInstancesGroups.title"
				>
					<v-btn class="mb-4" @click="() => removeInstance(currentInstancesGroupsIndex)">Remove</v-btn>
					
					<v-text-field
						label="instance title"
						:rules="rules.req"
						v-model="instances[currentInstancesGroupsIndex].title"
					/>

					<v-select
						:items="types"
						v-model="currentInstancesGroups.body.type"
						label="type"
					></v-select>

					<component
						:is="templates[currentInstancesGroups.body.type]"
						:instances-group="JSON.stringify(currentInstancesGroups)"
						@update:instances-group="receiveObject"
					>

					</component>
				</v-card>
				
				<v-card
					v-else
					color="background-light"
					elevation="0"
					class="pa-4"
				>
					no instances group selected
				</v-card>
			</v-col>
		</v-row>
	</v-form>
</template>

<script>
export default {
	name: 'service-create',
	data: () => ({
		formValid: false,
		rules: {
			req: [
				v => !!v || 'required field'
			],
			password: [
				v => !!v || 'password required',
        v => v.length > 6 || 'password must be at least 6 characters length',
			],
			nubmer: [
				v => Number(v) == v || 'must be a correct number'
			]
		},
		service: {
			"version": "1",
			"title": "",
			"context": {},
			"instances_groups": {
			}
		},
		instances: [],
		currentInstancesGroups: {},
		currentInstancesGroupsIndex: -1,
		types: [
			'ione',
			'custom'
		],
		templates: {}
	}),
	methods: {
		addInstance(title = ""){
			if(this.instances.some(inst => inst.title == ''))
				return
			this.instances = [...this.instances, this.defaultInstance(title)]
			if(this.instances.length == 1)
				this.selectIntance(0)
		},
		removeInstance(index){
			this.instances.splice(index, 1)
			let newIndex = index - 1;
			if(this.instances.length > 0 && newIndex == -1)
				newIndex = 0
				
			this.selectIntance(newIndex)
		},
		defaultInstance(title = ""){
			return {
				title: title,
				body: {
					type: 'ione',
					"resources": {
							"ips_public": 1
					},
				}
			}
		},
		selectIntance(index = -1){
			if(index >= 0){
				this.currentInstancesGroups = JSON.parse(JSON.stringify(this.instances[index]))
			} else {
				this.currentInstancesGroups = {}
			}
			this.currentInstancesGroupsIndex = index;
		},
		receiveObject(newVal){
			this.instances[this.currentInstancesGroupsIndex] = JSON.parse(newVal);
			this.selectIntance(this.currentInstancesGroupsIndex)
		}
	},
	created(){
		const types = require.context('@/components/modules/', true, /serviceCreate\.vue$/)
		types.keys().forEach(key => {
			const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/serviceCreate\.vue/i);
			if (matched && matched.length > 1) {
				const type = matched[1]
				this.types.push(type);
				this.templates[type] = () => import(`@/components/modules/${type}/serviceCreate.vue`)
			}
		})

		this.addInstance('test')
	}
}
</script>

<style>

</style>
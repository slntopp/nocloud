<template>
	<div class="module">
		<v-card
			v-for="(instance, index) in instances" :key="index"
			class="mb-4 pa-2"
			elevation="0"
			color="background"
		>
			
			<v-row>
				<v-col
					cols="6"
				>
					<v-text-field
						@change="(newVal) => setValue(index + '.title', newVal)"
						label="title"
						v-model="instance.title"
					>
					</v-text-field>
				</v-col>
				<v-col class="d-flex justify-end">
					<v-btn @click="() => remove(index)">
						remove
					</v-btn>
				</v-col>
			</v-row>

			<v-row>
				<v-col
					cols="6"
				>
					<v-text-field
						@change="(newVal) => setValue(index + '.config.template_id', newVal)"
						label="template_id"
						v-model="instance.config.template_id"
					>
					</v-text-field>
				</v-col>
				<v-col
					cols="6"
				>
					<v-text-field
						@change="(newVal) => setValue(index + '.config.password', newVal)"
						label="password"
						v-model="instance.config.password"
					>
					</v-text-field>
				</v-col>
			</v-row>

			<v-row>
				<v-col
					cols="6"
				>
					<v-text-field
						@change="(newVal) => setValue(index + '.resources.cpu', newVal)"
						label="cpu"
						v-model="instance.resources.cpu"
						type="number"
					>
					</v-text-field>
				</v-col>
				<v-col
					cols="6"
				>
					<v-text-field
						@change="(newVal) => setValue(index + '.resources.ram', newVal)"
						label="ram"
						v-model="instance.resources.ram"
						type="number"
					>
					</v-text-field>
				</v-col>
				<v-col
					cols="6"
				>
					<v-text-field
						@change="(newVal) => setValue(index + '.resources.drive_type', newVal)"
						label="drive type"
						v-model="instance.resources.drive_type"
					>
					</v-text-field>
				</v-col>
				<v-col
					cols="6"
				>
					<v-text-field
						@change="(newVal) => setValue(index + '.resources.drive_size', newVal)"
						label="drive size"
						v-model="instance.resources.drive_size"
						type="number"
					>
					</v-text-field>
				</v-col>
				<v-col
					cols="6"
				>
					<v-text-field
						@change="(newVal) => setValue(index + '.resources.ips_public', newVal)"
						label="ips public"
						v-model="instance.resources.ips_public"
						type="number"
					>
					</v-text-field>
				</v-col>
				<v-col
					cols="6"
				>
					<v-text-field
						@change="(newVal) => setValue(index + '.resources.ips_private', newVal)"
						label="ips private"
						v-model="instance.resources.ips_private"
						type="number"
					>
					</v-text-field>
				</v-col>
			</v-row>

		</v-card>
		<v-row>
			<v-col class="d-flex justify-center">
				<v-btn
					class="mx-2"
					
					small
					color="background"
					@click="addInstance"
				>
					<v-icon dark>
						mdi-plus-circle-outline
					</v-icon>
					add instance
				</v-btn>
			</v-col>
		</v-row>
	</div>
</template>

<script>
export default {
	name: 'ione-create-service-module',
	props: ['instances-group'],
	data: () => ({
		defaultItem: {
			"title": "instance",
			"config": {
				"template_id": "",
				"password": ""
			},
			"resources": {
				"cpu": 1,
				"ram": "1024",
				"drive_type": "SSD",
				"drive_size": "10000",
				"ips_public": 0,
				"ips_private": 0
			}
		},
		types: [
			'SSD', 'HDD'
		],
		// instances: []
	}),
	methods: { 
		addInstance(){
			const item = JSON.parse(JSON.stringify(this.defaultItem));
			const data = JSON.parse(this.instancesGroup)

			item.title += "#" + (data.body.instances.length + 1);

			data.body.instances.push(item);
			this.change(data)
		},
		remove(index){
			// this.instances.splice(index, 1);
			const data = this.group;
			data.body.instances.splice(index, 1);
			this.change(data);
		},
		setValue(path, val){
			const data = JSON.parse(this.instancesGroup)
			setToValue(data.body.instances, val, path);
			this.change(data)
		},
		change(data){
			this.$emit('update:instances-group', JSON.stringify(data))
		}
	},
	computed: {
		inst(){
			return JSON.parse(this.instancesGroup)
		},
		instances(){
			const data = this.group
			return data.body.instances
		},
		group(){
			const data = JSON.parse(this.instancesGroup)
			return data
		}
	},
	created(){
		const data = JSON.parse(this.instancesGroup)
		if(!data.body.instances){
			data.body.instances = []
		}
			
		this.change(data)
	}
}

function setToValue(obj, value, path) {
	path = path.split('.');
	let i;
	for (i = 0; i < path.length - 1; i++){
		if(path[i] === "__proto__" || path[i] === "constructor") 
			throw new Error("Can't use that path because of: " + path[i]);
		obj = obj[path[i]];
	}
	obj[path[i]] = value;
}
</script>

<style>

</style>
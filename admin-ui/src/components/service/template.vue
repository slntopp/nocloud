<template>
	<v-card
		elevation="0"
		color="background-light"
		class="pa-4"
	>
		Service
		<span
			class="service__display-trigger"
			@click="() => ObjectDisplay = (ObjectDisplay == 'YAML' ? 'JSON' : 'YAML')"
		>
			{{ObjectDisplay}}
		</span>
		<v-switch
			style="display: inline-flex"
			v-model="ObjectDisplay"
			true-value="JSON"
			false-value="YAML"
		></v-switch>
		:
		<pre
			v-if="ObjectDisplay == 'YAML'"
			v-html="serviceObjectYAML"
		></pre>
		<pre
			v-else-if="ObjectDisplay == 'JSON'"
			v-html="serviceObjectJSON"
		></pre>
	</v-card>
</template>

<script>
import yaml from "yaml"

export default {
	name: 'service-template',
	data: () => ({
		ObjectDisplay: "YAML",
	}),
	props: {
		service: {
			type: Object,
			required: true
		}
	},
	computed: {
		serviceObjectJSON(){
			let json = JSON.stringify(this.service, null, 2);
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
		serviceObjectYAML(){
			const doc = new yaml.Document();
			doc.contents = this.service;

			return doc.toString();
		}
	}
}
</script>


<style scoped lang="scss">

.service__display-trigger{
	cursor: pointer;
	color: var(--v-primary-base)
}
</style>

<style lang="scss">
pre {
	padding: 5px;
	margin: 5px;
	background-color: var(--v-background-light-base);
	border-radius: 4px;
	white-space: pre-wrap;
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
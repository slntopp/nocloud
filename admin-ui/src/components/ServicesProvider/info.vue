<template>
	<v-card
		elevation="0"
		color="background-light"
		class="pa-4"
	>
		<v-row>
			<v-col>
				<v-text-field
					readonly
					:value="template.uuid"
					label="template uuid"
					style="display: inline-block; width: 330px"
					:append-icon="copyed == 'rootUUID' ? 'mdi-check': 'mdi-content-copy'"
					@click:append="addToClipboard(template.uuid, 'rootUUID')"
				>
				</v-text-field>
			</v-col>
			<v-col>
				<v-text-field
					readonly
					:value="template.type"
					label="template type"
					style="display: inline-block; width: 150px"
				>
				</v-text-field>
			</v-col>
		</v-row>

		secrets:
		<v-row>
			<v-col>
				<v-text-field
					readonly
					:value="template.secrets.group"
					label="group"
					style="display: inline-block; width: 330px"
				>
				</v-text-field>
			</v-col>
			<v-col>
				<v-text-field
					readonly
					:value="template.secrets.host"
					label="host"
					style="display: inline-block; width: 330px"
				>
				</v-text-field>
			</v-col>
			<v-col>
				<v-text-field
					readonly
					:value="template.secrets.user"
					label="user"
					style="display: inline-block; width: 330px"
				>
				</v-text-field>
			</v-col>
			<v-col>
				<v-text-field
					readonly
					:value="template.secrets.pass"
					label="password"
					style="display: inline-block; width: 330px"
					:type="showPassword ? 'text' : 'password'"
					:append-icon="showPassword ? 'mdi-eye-off': 'mdi-eye'"
					@click:append="() => showPassword = !showPassword"
				>
				</v-text-field>
			</v-col>
		</v-row>

		variables:
		<v-row>
			<v-col v-for="(variable, varTitle) in template.vars" :key="varTitle">
				{{varTitle.replaceAll('_', ' ')}}
				<v-row>
					<v-col :cols="12" v-for="(value, key) in variable.value" :key="key">
						<v-text-field
							readonly
							:value="JSON.stringify(value)"
							:label="key"
							style="display: inline-block; width: 200px"
						>
						</v-text-field>
					</v-col>
				</v-row>
			</v-col>
		</v-row>

		<template v-if="template.extentions && Object.keys(template.extentions).length > 0">
			extentions:
			<component
				v-for="(extention, extName) in template.extentions"
				:is="extentionsMap[extName].pageComponent"
				:key="extName"
				:data="extention"
			>
			</component>
		</template>

	</v-card>
</template>

<script>
import extentionsMap from "@/components/extentions/map.js"

export default {
	name: "services-provider-info",
	data: () => ({
		copyed: null,
		opened: [],
		showPassword: false,
		extentionsMap
	}),
	props: {
		template: {
			type: Object,
			required: true
		}
	},
	methods: {
		addToClipboard(text, index){
			navigator.clipboard.writeText(text)
			.then(()=>{
				this.copyed = index
			})
			.catch(res=>{
				console.error(res);
			})
		}
	},
}
</script>

<style>

</style>
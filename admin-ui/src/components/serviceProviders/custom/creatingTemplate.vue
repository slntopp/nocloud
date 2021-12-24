<template>
	<div class="te">
		<v-row align="center">
			<v-col cols="3">
				<v-subheader>
					secrets
				</v-subheader>
			</v-col>

			<v-col
				cols="9"
			>
				<v-text-field
					label="json"
					:value="JSON.stringify(secrets)"
					:error-messages="errors.secrets"
					@change="(data) => changeHandler('secrets', data)"
				></v-text-field>
			</v-col>
		</v-row>
		<v-row align="center">
			<v-col cols="3">
				<v-subheader>
					vars
				</v-subheader>
			</v-col>

			<v-col
				cols="9"
			>
				<v-text-field
					label="json"
					:value="JSON.stringify(vars)"
					:error-messages="errors.vars"
					@change="(data) => changeHandler('vars', data)"
				></v-text-field>
			</v-col>
		</v-row>
		
	</div>
</template>

<script>
export default {
	name: "servicesProviders-create-custom",
	props: {
		secrets: {
			type: Object,
			default: () => ({})
		},
		vars: {
			type: Object,
			default: () => ({})
		},
		passed: {
			type: Boolean,
			default: false
		}
	},
	data: () => ({
		errors: {
			secrets: [],
			vars: []
		}
	}),
	methods: {
		isJSON(str){
			try{
				JSON.parse(str);
				return true;
			} catch {
				return false;
			}
		},
		changeHandler(input, data){
			const isJSON = this.isJSON(data);
			if(!isJSON){
				this.errors[input] = ["is not JSON"];
			} else {
				this.errors[input] = [];
				data = JSON.parse(data);
			}
			this.$emit(`change:${input}`, data)
			this.$emit(`passed`, Object.values(this.errors).every(el => el.length == 0))
		}
	}
}
</script>

<style>

</style>
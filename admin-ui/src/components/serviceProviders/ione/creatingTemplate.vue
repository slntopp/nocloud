<template>
	<div class="sp-ione">
		<v-row align="center">
			<v-col cols="3">
				<v-subheader>
					Host
				</v-subheader>
			</v-col>

			<v-col
				cols="9"
			>
				<v-text-field
					label="example.com"
				></v-text-field>
			</v-col>
			<v-col cols="3">
				<v-subheader>
					credencials
				</v-subheader>
			</v-col>

			<v-col
				cols="9"
			>
				<v-text-field
					label="log:pass"
				></v-text-field>
			</v-col>
			<v-col cols="3">
				<v-subheader>
					group
				</v-subheader>
			</v-col>

			<v-col
				cols="9"
			>
				<v-text-field
					label="100"
				></v-text-field>
			</v-col>
		</v-row>
	</div>
</template>

<script>
export default {
	name: "servicesProviders-create-ione",
	props: {
		secrets: {
			type: Object,
			default: () => ({})
		},
		vars: {
			type: Object,
			default: () => ({})
		},
	}
}
</script>

<style>

</style>
<template>
	<v-card
		elevation="0"
		color="background-light"
		class="pa-4"
	>
		<template-json-editor-new
			:value="template"
			title="Template JSON"
			@save="editServiceProvider"
		/>
	</v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import TemplateJsonEditorNew from "@/components/TemplateJsonEditorNew.vue";

export default {
	name: 'services-provider-template',
	components: { TemplateJsonEditorNew },
	mixins: [snackbar],
	props: {
		template: {
			type: Object,
			required: true
		}
	},
	methods: {
		async editServiceProvider(parsedValue) {
			try {
				const request = parsedValue;
				if (request.locations.length < 1) {
					request.locations = [{ id: '_nocloud.remove' }];
				}
				await api.servicesProviders.update(this.template.uuid, request);
				this.showSnackbarSuccess({
					message: 'Service provider edited successfully'
				});
				this.$router.go();
			} catch (err) {
				this.showSnackbarError({ message: err });
			}
		}
	}
}
</script>

<style scoped lang="scss">
</style>

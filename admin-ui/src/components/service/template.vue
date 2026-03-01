<template>
	<v-card
		elevation="0"
		color="background-light"
		class="pa-4"
	>
		<template-json-editor-new
			:value="template"
			title="Template JSON"
			@save="editService"
		/>
	</v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import TemplateJsonEditorNew from "@/components/TemplateJsonEditorNew.vue";

export default {
	name: 'service-template',
	components: { TemplateJsonEditorNew },
	mixins: [snackbar],
	props: {
		template: {
			type: Object,
			required: true
		}
	},
	methods: {
		async editService(parsedValue) {
			try {
				await api.services._update(parsedValue);
				this.showSnackbarSuccess({
					message: 'Service edited successfully'
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

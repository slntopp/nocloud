<template>
	<v-card
		elevation="0"
		color="background-light"
		class="pa-4"
	>
		<template-json-editor-new
			:value="template"
			title="Template JSON"
			@save="editTemplate"
		/>
	</v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import TemplateJsonEditorNew from "@/components/TemplateJsonEditorNew.vue";

export default {
	name: "instance-template",
	components: { TemplateJsonEditorNew },
	mixins: [snackbar],
	props: {
		template: {
			type: Object,
			required: true
		}
	},
	methods: {
		async editTemplate(parsedValue) {
			try {
				await api.instances.update(this.template.uuid, parsedValue);
				this.showSnackbarSuccess({
					message: 'Instance template edited successfully'
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

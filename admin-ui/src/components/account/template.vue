<template>
	<v-card
		elevation="0"
		color="background-light"
		class="pa-4"
	>
		<template-json-editor-new
			:value="account"
			title="Template JSON"
			@save="editAccount"
		/>
	</v-card>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import TemplateJsonEditorNew from "@/components/TemplateJsonEditorNew.vue";

export default {
	name: 'account-template',
	components: { TemplateJsonEditorNew },
	mixins: [snackbar],
	props: {
		account: {
			type: Object,
			required: true
		}
	},
	methods: {
		async editAccount(parsedValue) {
			try {
				await api.accounts.update(this.account.uuid, parsedValue);
				this.showSnackbarSuccess({
					message: 'Account edited successfully'
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

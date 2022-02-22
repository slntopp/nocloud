<template>
	<v-snackbar
		:value="snackbar.visibility || false"
		@change="(state) => $emit('change', state)"
		:timeout="snackbar.timeout || 3000"
		:color="snackbar.color || 'background'"
	>
		{{snackbar.message || 'missing message'}}
		<template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
			<router-link :to="snackbar.route">
				Look up.
			</router-link>
		</template>
		

		<template v-if="!hideClose" v-slot:action="{ attrs }">
			<v-btn
				:color="snackbar.buttonColor"
				text
				v-bind="attrs"
				@click="$emit('change', false)"
			>
				Close
			</v-btn>
		</template>
	</v-snackbar>
</template>

<script>
export default {
	name: 's-alert',
	props: {
		snackbar: {
			type: Object,
			required: true
		},
		'hide-close': Boolean
	}
}
</script>

<style>

</style>
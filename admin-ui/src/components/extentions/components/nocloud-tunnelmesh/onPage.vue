<template>
	<v-card
		color="background"
		elevation="0"
		class="pa-2"
	>
		nocloud-tunnelmesh

		<v-row>
			<v-col>
				<v-text-field
					readonly
					:value="data.fingerprint"
					label="fingerprint"
				>
				</v-text-field>
			</v-col>
			<v-col>
				<v-text-field
					readonly
					:value="data.hostname"
					label="hostname"
				>
				</v-text-field>
			</v-col>
		</v-row>
	</v-card>
</template>

<script>
export default {
	name: "tunnelmesh-extention-page",
	data: () => ({
	}),
	props: {
		data: {
			type: Object,
			required: true
		}
	},
	computed: {
	},
	methods: {
	},
	mounted() {
	},
}
</script>

<style>

</style>
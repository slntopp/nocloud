<template>
	<div class="module">
		<v-row>
			<v-col>
				<v-list>

				</v-list>
			</v-col>
			<v-col>
				{{inst}}
			</v-col>
		</v-row>
	</div>
</template>

<script>
export default {
	name: 'ione-create-service-module',
	props: ['instances-group'],
	data: () => ({
		
	}),
	computed: {
		inst(){
			return JSON.parse(this.instancesGroup)
		}
	}
}
</script>

<style>

</style>
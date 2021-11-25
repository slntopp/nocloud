<template>
	<div class="widgets gg-15px d-flex px-4 pt-4 flex-wrap">
		<component v-for="widget in Object.keys(widgets)" :key="widget" :is="widgets[widget]">
		</component>
	</div>
</template>

<script>
import HealthWidget from "@/components/widgets/health"

export default {
	name: 'dashboard-view',
	components: {
		HealthWidget
	},
	data: ()=>({
		widgets: [
			'HealthWidget',
		]
	})
}
</script>

<style>

</style>
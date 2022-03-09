<template>
	<div class="widgets gg-15px d-flex px-4 pt-4 flex-wrap" style="align-items: start">
		<component v-for="widget in Object.keys(widgets)" :key="widget" :is="widgets[widget]">
		</component>
	</div>
</template>

<script>
import HealthWidget from "@/components/widgets/health"
import ServicesWidget from "@/components/widgets/services"
import RoutinesWidget from "@/components/widgets/routines"

export default {
	name: 'dashboard-view',
	components: {
		HealthWidget,
		ServicesWidget,
		RoutinesWidget,
	},
	data: ()=>({
		widgets: [
			'HealthWidget',
			'ServicesWidget',
			'RoutinesWidget',
		]
	})
}
</script>

<style>

</style>
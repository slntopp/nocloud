<template>
	<div class="service pa-4">
		<div class="page__title mb-5">
			<router-link :to="{name: 'Services'}">Services</router-link>
			/
			{{ service.title }} 
			<v-chip
				x-small
				:color="chipColor"
			>
			</v-chip>
		</div>
		
		<v-tabs
			class="rounded-t-lg"
			v-model="tabs"
			background-color="background-light"
		>
			<v-tab>Info</v-tab>
			<v-tab>Control</v-tab>
			<v-tab>Template</v-tab>
		</v-tabs>

		<v-tabs-items v-model="tabs" style="background: var(--v-background-light-base)" class="rounded-b-lg">
			<v-tab-item>
				<service-info
					:service="service"
				/>
			</v-tab-item>

			<v-tab-item>
				<service-control
					:service="service"
					:chip-color="chipColor"
				/>
			</v-tab-item>

			<v-tab-item>
				<service-template
					:service="service"
				/>
			</v-tab-item>

		</v-tabs-items>
	</div>
</template>

<script>
import serviceTemplate from "@/components/service/template.vue"
import serviceControl from "@/components/service/control.vue"
import serviceInfo from "@/components/service/info.vue"

export default {
	name: 'service-view',
	components: {
		'service-template': serviceTemplate,
		'service-control': serviceControl,
		'service-info': serviceInfo,
	},
	data: () => ({
		notFound: false,
		tabs: 0
	}),
	
	computed: {
		service(){
			const items = this.$store.getters['services/all']
			const item = items.find(el => el.uuid == this.serviceId)
	
			if(item)
				return item
	
			
			return {}
		},
		serviceId(){
			return this.$route.params.serviceId;
		},
		chipColor(){
			const dict = {
				'init': 'orange darken-2',
				'up': 'green darken-2',
				'del': 'gray darken-2'
			}
			return dict[this.service.status] ?? 'blue-grey darken-2'
		}
	},
	created(){
		this.$store.dispatch('services/fetch')
		.then(() => {
			this.notFound = true
		})
	},
	mounted(){
		document.title = `${this.service.title} | NoCloud`
		this.$store.commit('reloadBtn/setCallback', {func: this.$store.dispatch, params: ['services/fetch']})
	}
}
</script>

<style scoped lang="scss">
.page__title{
	color: var(--v-primary-base);
	font-weight: 400;
	font-size: 32px;
	font-family: "Quicksand", sans-serif;
	line-height: 1em;
	margin-bottom: 10px;
}
</style>
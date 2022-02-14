<template>
	<div class="servicesProviders pa-4 flex-wrap">
		<div class="page__title mb-5">
			<router-link :to="{name: 'ServicesProviders'}">services providers</router-link>
			/
			{{ title }}
		</div>

		<v-tabs
			class="rounded-t-lg"
			v-model="tabsIndex"
			background-color="background-light"
		>
			<v-tab
				v-for="tab in tabs"
				:key="tab.title"
			>
				{{tab.title}}
			</v-tab>
		</v-tabs>

		<v-tabs-items v-model="tabsIndex" style="background: var(--v-background-light-base)" class="rounded-b-lg">
			<v-tab-item
				v-for="tab in tabs"
				:key="tab.title"
			>
				<v-progress-linear
					v-if="loading"
					indeterminate
					class="pt-2"
				/>
				<component
					v-if="!loading && item"
					:is="tab.component"
					:template="item"
				>
				</component>
			</v-tab-item>

		</v-tabs-items>

	</div>
</template>

<script>

export default {
	name: "servicesProvider-page",
	data: () => ({
		found: false,
		tabsIndex: 0,
		tabs: [
			{
				title: 'Info',
				component: () => import('@/components/ServicesProvider/info.vue')
			},
			{
				title: 'Template',
				component: () => import('@/components/ServicesProvider/template.vue')
			},
		]
	}),
	computed: {
		uuid(){
			return this.$route.params.uuid;
		},
		item(){
			const items = this.$store.getters['servicesProviders/all']
			const item = items.find(el => el.uuid == this.uuid)
	
			if(item)
				return item
			
			return null
		},
		title(){
			return this?.item?.title ?? 'not found'
		},
		loading(){
			return this.$store.getters['servicesProviders/isLoading']
		}
	},
	created(){
		this.$store.dispatch('servicesProviders/fetch')
		.then(() => {
			this.found = !!this.service;
			document.title = `${this.title} | NoCloud`
		})
	},
	mounted(){
		document.title = `${this.title} | NoCloud`
		this.$store.commit('reloadBtn/setCallback', {func: this.$store.dispatch, params: ['servicesProviders/fetch']})
	}
}
</script>

<style>
.page__title{
	color: var(--v-primary-base);
	font-weight: 400;
	font-size: 32px;
	font-family: "Quicksand", sans-serif;
	line-height: 1em;
	margin-bottom: 10px;
}
</style>
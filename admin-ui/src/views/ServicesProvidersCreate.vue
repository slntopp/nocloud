<template>
	<div class="servicesProviders-create pa-4">
		<div class="page__title">
			Create service provider
		</div>
		<div class="page__content">
			<v-container>
				<v-row align="center">
					<v-col cols="3">
						<v-subheader>
							Provider type
						</v-subheader>
					</v-col>

					<v-col
						cols="9"
					>
						<v-select
							v-model="provider.type"
							:items="types"
							label="Type"
						></v-select>
					</v-col>
				</v-row>
				
				<v-row align="center">
					<v-col cols="3">
						<v-subheader>
							Provider title
						</v-subheader>
					</v-col>

					<v-col
						cols="9"
					>
						<v-text-field
							v-model="provider.title"
							label="Title"
						></v-text-field>
					</v-col>
				</v-row>

				<v-divider></v-divider>

				<component
					:is="templates[provider.type]"
					:secrets="provider.secrets"
					@change:secrets="test"
					:vars="provider.vars"
					@change:vars="test"
				></component>

				
				<v-btn
					color="background-light"
					class="mr-2"
				>
					create
				</v-btn>
			</v-container>
		</div>
	</div>
</template>

<script>
export default {
	name: "servicesProviders-create",
	data: () => ({
		types: [],
		templates: {},

		provider: {
			type: 'custom',
			title: '',
			secrets: {},
			vars: {}
		},
	}),
	created(){
		const types = require.context('@/components/serviceProviders/', true, /creatingTemplate\.vue$/)
		types.keys().forEach(key => {
			const matched = key.match(/\.\/([A-Za-z0-9-_,\s]*)\/creatingTemplate\.vue/i);
			if (matched && matched.length > 1) {
				const type = matched[1]
				this.types.push(type);
				this.templates[type] = () => import(`@/components/serviceProviders/${type}/creatingTemplate.vue`)
			}
		})
	},
	computed: {
		template(){
			return () => import(`@/components/serviceProviders/${this.type}/creatingTemplate.vue`);
		}
	},
	methods: {
		test(e){
			console.log(e);
		}
	}
}
</script>

<style scoped lang="scss">
.page__title{
	color: #FF00FF;
	font-weight: 400;
	font-size: 32px;
	font-family: "Quicksand";
	// writing-mode: vertical-lr;
	line-height: 1em;
	margin-bottom: 10px;
}

.page__content{
	flex-grow: 1;
	max-width: 750px;
}
</style>
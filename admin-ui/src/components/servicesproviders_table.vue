<template>
	<nocloud-table
		:loading="loading"
		:items="tableData"
		:value="selected"
		@input="handleSelect"
		:single-select="singleSelect"
		:headers="Headers"
		item-key="uuid"
		:footer-error="fetchError"
	>
		

		<!-- <template v-slot:[`item.title`]="{ item }">
			<router-link :to="item.route">
				{{item.title}}
			</router-link>
		</template> -->
	</nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue"

const Headers = [
	{ text: 'title', value: 'titleLink' },
	{ text: 'type', value: 'type' },
	{
		text: 'UUID',
		align: 'start',
		sortable: true,
		value: 'uuid',
	},
];

export default {
	name: "servicesProviders-table",
	components: {
		"nocloud-table": noCloudTable
	},
	props: {
		value: {
			type: Array,
			default: () => []
		},
		"single-select": {
			type: Boolean,
			default: false
		}
	},
	data () {
		return {
			selected: this.value,
			loading: false,
			Headers,
			fetchError: ''
		}
	},
	methods: {
		handleSelect(item){
			this.$emit('input', item)
		}
	},
	computed: {
		tableData(){
			return this.$store.getters['servicesProviders/all'].map(el => ({
				titleLink: el.title,
				type: el.type,
				route: {
					name: 'ServicesProvider',
					params: {uuid: el.uuid}
				}
			}));
		}
	},
	created() {
		this.loading = true;
		this.$store.dispatch('servicesProviders/fetch')
		.then(() => {
			this.fetchError = ''
		})
		.finally(()=>{
			this.loading = false;
		})
		.catch(err => {
			console.log(`err`, err)
			this.fetchError = 'Can\'t reach server'
		})
	},
}
</script>

<style>

</style>
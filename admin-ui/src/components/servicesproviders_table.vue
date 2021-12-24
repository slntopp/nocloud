<template>
	<nocloud-table
		:loading="loading"
		:items="tableData"
		:value="selected"
		@input="handleSelect"
		:single-select="singleSelect"
		:headers="Headers"
		item-key="uuid"
	>

	</nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue"

const Headers = [
	{ text: 'title', value: 'title' },
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
		}
	},
	methods: {
		handleSelect(item){
			this.$emit('input', item)
		}
	},
	computed: {
		tableData(){
			return this.$store.getters['servicesProviders/all'];
		}
	},
	created() {
		this.loading = true;
		this.$store.dispatch('servicesProviders/fetch')
		.finally(()=>{
			this.loading = false;
		})
	},
}
</script>

<style>

</style>
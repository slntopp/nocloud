<template>
	<nocloud-table
		:loading="loading"
		:items="tableData"
		:value="selected"
		@input="handleSelect"
		:single-select="singleSelect"
		:footer-error="fetchError"
	>

	</nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue"

export default {
	name: "namespaces-table",
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
			return this.$store.getters['namespaces/all'];
		}
	},
	created() {
		this.loading = true;
		this.$store.dispatch('namespaces/fetch')
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
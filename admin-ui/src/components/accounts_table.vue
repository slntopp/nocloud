<template>
	<nocloud-table
		:loading="loading"
		:items="tableData"
		:value="selected"
		@input="handleSelect"
		:single-select="singleSelect"
	>

	</nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue"

export default {
	name: "accounts-table",
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
			loading: false
		}
	},
	methods: {
		handleSelect(item){
			this.$emit('input', item)
		}
	},
	computed: {
		tableData(){
			return this.$store.getters['accounts/all'];
		}
	},
	created() {
		this.loading = true;
		this.$store.dispatch('accounts/fetch')
		.finally(()=>{
			this.loading = false;
		})
	},
}
</script>

<style>

</style>
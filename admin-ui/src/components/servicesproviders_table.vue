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
		

		<template v-slot:[`item.state`]="{ value }">
			<v-chip
				small
				:color="chipsColor(value)"
			>
				{{value}}
			</v-chip>
		</template>

	</nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue"

const Headers = [
	{ text: 'title', value: 'titleLink' },
	{ text: 'type', value: 'type' },
	{ text: 'state', value: 'state' },
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
		},
		chipsColor(state){
			if(!state){
				return 'gray'
			}
			switch (state.toLowerCase()) {
				case 'running':
				case 'operation':
					return 'success'
				case 'unknown':
				case 'deleted':
				case 'failure':
					return 'error'
		
				default:
					return 'gray';
			}
		},
	},
	computed: {
		tableData(){
			return this.$store.getters['servicesProviders/all'].map(el => ({
				titleLink: el.title,
				type: el.type,
				route: {
					name: 'ServicesProvider',
					params: {uuid: el.uuid}
				},
				state: el?.state?.state ?? 'UNKNOWN'
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
			this.fetchError = 'Can\'t reach the server'
			if(err.response){
				this.fetchError += `: [ERROR]: ${err.response.data.message}`
			} else {
				this.fetchError += `: [ERROR]: ${err.toJSON().message}`
			}
		})
	},
}
</script>

<style>

</style>
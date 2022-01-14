<template>
	<nocloud-table
		:loading="loading"
		:items="tableData"
		:value="selected"
		@input="handleSelect"
		:single-select="singleSelect"
		:headers="Headers"
		item-key="id"
		no-hide-uuid
	>
	</nocloud-table>
</template>

<script>
import noCloudTable from "@/components/table.vue"

const Headers = [
	{ text: 'location', value: 'location' },
	{ text: 'value', value: 'value' },
	{ text: 'type', value: 'type' },
	{ text: 'ttl', value: 'ttl' },
];

export default {
	name: "zones-table",
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
			const dns = this.$store.getters['dns/getHost'](this.$route.params.dnsname);

			const hosts = dns;
			const result = []
			
			for (const location in hosts) {
				for (const key in hosts[location]) {
					hosts[location][key].forEach(el => {
						const temp = {};
						temp.id = Math.random();
						temp.location = location;
						temp.type = key;
						temp.value = el?.text ?? el.host ?? el.ip ?? "";
						temp.ttl = el.ttl;

						result.push(temp);
					})
				}
			}
			return result;
		}
	},
	created() {
		this.loading = true;
		this.$store.dispatch('dns/fetchHosts', this.$route.params.dnsname)
		.finally(()=>{
			this.loading = false;
		})
	},
}
</script>

<style>

</style>
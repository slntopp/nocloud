<template>
	<v-data-table
		:item-key="itemKey"
		class="elevation-0 background-light rounded-lg"
		:loading="loading"
		loading-text="Loading... Please wait"
		color="background-light"
		:headers="headers"
		:items="items"
		show-select
		:value="selected"
		@input="handleSelect"
		:single-select="singleSelect"
	>
	
		<template
			v-if="!noHideUuid"
			v-slot:[`item.${itemKey}`]="props"
		>

			<template
				v-if="showed.includes(props.index)"
			>
				{{props.value}}
			</template>
			<v-btn
				v-else
				icon
				@click="showID(props.index)"
			>
				<v-icon>mdi-eye-outline</v-icon>
			</v-btn>
			<v-btn
				icon
				@click="addToClipboard(props.value, props.index)"
			>
				<v-icon
					v-if="copyed == props.index"
				>
					mdi-check
				</v-icon>
				<v-icon
					v-else
				>
					mdi-content-copy
				</v-icon>
			</v-btn>
		</template>
		
		<template v-slot:[`item.titleLink`]="{ item }">
			<router-link :to="item.route">
				{{item.titleLink}}
			</router-link>
		</template>

	</v-data-table>
</template>

<script>

const defaultHeaders = [
	{ text: 'title', value: 'title' },
	{
		text: 'UUID',
		align: 'start',
		sortable: true,
		value: 'uuid',
	},
];

export default {
	name: "nocloud-table",
	props: {
		loading: Boolean,
		items: {
			type: Array,
			default: () => ([])
		},
		value: {
			type: Array,
			default: () => ([])
		},
		headers: {
			type: Array,
			default: () => defaultHeaders
		},
		"single-select": {
			type: Boolean,
			default: false
		},
		'item-key': {
			type: String,
			default: 'uuid'
		},
		'no-hide-uuid': {
			type: Boolean,
			default: false
		}
	},
	data(){
		return {
			selected: this.value,
			showed: [],
			copyed: -1
		}
	},
	methods: {
		handleSelect(item){
			this.$emit('input', item)
		},
		addToClipboard(text, index){
			navigator.clipboard.writeText(text)
			.then(()=>{
				this.copyed = index
			})
			.catch(res=>{
				console.log(res);
			})
		},
		showID(index){
			this.showed.push(index);
		}
	}
}
</script>

<style>

</style>
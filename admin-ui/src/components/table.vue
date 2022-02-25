<template>
	<components :is="VDataTable"
		:item-key="itemKey"
		class="elevation-0 background-light rounded-lg"
		:loading="loading"
		loading-text="Loading... Please wait"
		color="background-light"
		:items="items"
		show-select
		:value="selected"
		@input="handleSelect"
		:single-select="singleSelect"
		:headers="headers"
		:expanded="expanded"
		@update:expanded="(nw) => $emit('update:expanded', nw)"
		:show-expand="showExpand"
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

		<template v-for="(_, scopedSlotName) in $scopedSlots" v-slot:[scopedSlotName]="slotData">
			<slot :name="scopedSlotName" v-bind="slotData" />
		</template>
		<template v-for="(_, slotName) in $slots" v-slot:[slotName]>
			<slot :name="slotName" />
		</template>

		<template v-if="footerError.length > 0" v-slot:footer>
			<v-toolbar
				class="mt-2"
				color="error"
				dark
				flat
			>
				<v-toolbar-title class="subheading">
					{{footerError}}
				</v-toolbar-title>
			</v-toolbar>
		</template>

	</components>
</template>

<script>
import {
  VDataTable,
} from 'vuetify/lib'

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
		},
    expanded: {
      type: Array,
      default: () => []
    },
    showSelect: Boolean,
    checkboxColor: String,
    showExpand: Boolean,
    showGroupBy: Boolean,
    height: [Number, String],
    hideDefaultHeader: Boolean,
    caption: String,
    dense: Boolean,
    headerProps: Object,
    calculateWidths: Boolean,
    fixedHeader: Boolean,
    headersLength: Number,
    expandIcon: {
      type: String,
      default: '$expand'
    },
    itemClass: {
      type: [String, Function],
      default: () => ''
    },
    loaderHeight: {
      type: [Number, String],
      default: 4
    },
		'footer-error': {
			type: String,
			default: ""
		}
	},
	data(){
		return {
			selected: this.value,
			showed: [],
			copyed: -1,
			VDataTable
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
				console.error(res);
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
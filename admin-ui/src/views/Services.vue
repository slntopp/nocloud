<template>
	<div class="services pa-4">
		<nocloud-table
			:items="services"
			:headers="headers"
      :expanded.sync="expanded"
      show-expand
		>

			<template
				v-slot:[`item.hash`]="{ item }"
			>
				<v-btn
					icon
					@click="addToClipboard(item.hash, item.index)"
				>
					<v-icon
						v-if="copyed == item.index"
					>
						mdi-check
					</v-icon>
					<v-icon
						v-else
					>
						mdi-content-copy
					</v-icon>
				</v-btn>
				{{hashTrim(item.hash)}}
			</template>

			<template v-slot:[`item.title`]="slotData">
        <div @click="clickColumn(slotData)">{{ slotData.item.title }}
				</div>
			</template>

			<template v-slot:[`item.status`]=" {value} ">
				<v-chip
					small
					:color="chipColor(value)"
				>
					{{value}}
				</v-chip>
			</template>

			<template v-slot:expanded-item="{ headers, item }">
				<td :colspan="headers.length" style="padding: 0">
					<v-card class="pa-4" color="background">
						<v-row>
							<v-col>groups count: {{Object.keys(item.instancesGroups).length}}</v-col>
							<v-col>
								status:
								<v-chip
									small
									:color="chipColor(item.status)"
								>
									{{item.status}}
								</v-chip>
							</v-col>
						</v-row>
					</v-card>
				</td>
			</template>

		</nocloud-table>
	</div>
</template>

<script>
import noCloudTable from "@/components/table.vue"

const headers = [
	{ text: 'title', value: 'title' },
	{ text: 'status', value: 'status' },
	{
		text: 'UUID',
		align: 'start',
		value: 'uuid',
	},
	{
		text: 'hash', value: 'hash'
	}
]

export default {
	name: "Services-view",
	components: {
		"nocloud-table": noCloudTable
	},
	data: () => ({
		headers,
		copyed: -1,
		expanded: []

	}),
	computed: {
		services(){
			return this.$store.getters['services/all']
		}
	},
	created(){
		this.$store.dispatch('services/fetch')
	},
	methods: {
		hashTrim(hash){
			console.log(hash);
			if(hash)
				return hash.slice(0, 8) + "..."
			else
				return "XXXXXXXX..."
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
    clickColumn(slotData) {
      // const indexRow = slotData.index;
      const indexExpanded = this.expanded.findIndex(i => i === slotData.item);
      if (indexExpanded > -1) {
        this.expanded.splice(indexExpanded, 1)
      } else {
        this.expanded.push(slotData.item);
      }
		},
		chipColor(state){
			const dict = {
				'init': 'orange darken-2',
				'up': 'green darken-2'
			}
			return dict[state] ?? 'blue-grey darken-2'
		}
	}
}
</script>

<style>

</style>
<template>
	<v-card
		elevation="0"
		color="background-light"
		class="pa-4"
	>
		<v-row
			align="center"
		>
			<v-col>
				<v-text-field
					readonly
					:value="service.uuid"
					label="service uuid"
					style="display: inline-block; width: 300px"
				>
				</v-text-field>
				<v-btn
					icon
					@click="addToClipboard(service.uuid, 'rootUUID')"
				>
					<v-icon
						v-if="copyed == 'rootUUID'"
					>
						mdi-check
					</v-icon>
					<v-icon
						v-else
					>
						mdi-content-copy
					</v-icon>
				</v-btn>
			</v-col>
		</v-row>
		
		<v-row justify="center" class="px-2 pb-2">
			<v-expansion-panels
				inset
				v-model="opened"
				multiple	
			>
				<v-expansion-panel
					v-for="(group, i) in Object.keys(service.instancesGroups)"
					:key="i"
					style="background: var(--v-background-base)"
				>
					<v-expansion-panel-header>{{group}}</v-expansion-panel-header>
					<v-expansion-panel-content>
						<v-text-field
							readonly
							:value="service.uuid"
							label="group uuid"
							style="display: inline-block; width: 300px"
						>
						</v-text-field>
						<v-btn
							icon
							@click="addToClipboard(service.instancesGroups[group].uuid, `${group}-UUID`)"
						>
							<v-icon
								v-if="copyed == `${group}-UUID`"
							>
								mdi-check
							</v-icon>
							<v-icon
								v-else
							>
								mdi-content-copy
							</v-icon>
						</v-btn>
					</v-expansion-panel-content>
				</v-expansion-panel>
			</v-expansion-panels>
		</v-row>

	</v-card>
</template>

<script>
export default {
	name: "service-info",
	data: () => ({
		copyed: null,
		opened: []
	}),
	props: {
		service: {
			type: Object,
			required: true
		}
	},
	methods: {
		addToClipboard(text, index){
			navigator.clipboard.writeText(text)
			.then(()=>{
				this.copyed = index
			})
			.catch(res=>{
				console.error(res);
			})
		},
	},
	mounted(){
		this.opened.push(0)
	}
}
</script>

<style>

</style>
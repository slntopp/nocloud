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
					style="display: inline-block; width: 330px"
					:append-icon="copyed == 'rootUUID' ? 'mdi-check': 'mdi-content-copy'"
					@click:append="addToClipboard(service.uuid, 'rootUUID')"
				>
				</v-text-field>
			</v-col>
			<v-col>
				<v-text-field
					readonly
					:value="hashpart(service.hash)"
					label="service hash"
					style="display: inline-block; width: 150px"
					:append-icon="copyed == 'rootHash' ? 'mdi-check': 'mdi-content-copy'"
					@click:append="addToClipboard(service.hash, 'rootHash')"
				>
				</v-text-field>
			</v-col>
		</v-row>
		
		groups:
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
					<v-expansion-panel-header>{{group}} | Type: {{service.instancesGroups[group].type}}</v-expansion-panel-header>
					<v-expansion-panel-content>
						<v-row>
							<v-col>
								<v-text-field
									readonly
									:value="location(group)"
									label="location"
									style="display: inline-block; width: 330px"
								>
								</v-text-field>
							</v-col>
							<v-col>
								<v-text-field
									readonly
									:value="service.uuid"
									label="group uuid"
									style="display: inline-block; width: 330px"
									:append-icon="copyed == `${group}-UUID` ? 'mdi-check': 'mdi-content-copy'"
									@click:append="addToClipboard(service.instancesGroups[group].uuid, `${group}-UUID`)"
								>
								</v-text-field>
							</v-col>
							<v-col>
								<v-text-field
									readonly
									:value="hashpart(service.instancesGroups[group].hash)"
									label="group hash"
									style="display: inline-block; width: 150px"
									:append-icon="copyed == `${group}-hash` ? 'mdi-check': 'mdi-content-copy'"
									@click:append="addToClipboard(service.instancesGroups[group].hash, `${group}-hash`)"
								>
								</v-text-field>
							</v-col>
						</v-row>
						Instances:
						<v-row>
							<v-col>
								<v-expansion-panels
									inset
									v-model="openedInstances[group]"
									multiple
								>
									<v-expansion-panel
										v-for="(instance, i) in service.instancesGroups[group].instances"
										:key="i"
										style="background: var(--v-background-light-base)"
									>
										<v-expansion-panel-header>{{instance.title}}</v-expansion-panel-header>
										<v-expansion-panel-content>
											<v-row>
												<v-col>
													<v-text-field
														readonly
														:value="instance.state.meta.state_str"
														label="state"
														style="display: inline-block; width: 100px"
													>
													</v-text-field>
												</v-col>
												<v-col>		
													<v-text-field
														readonly
														:value="instance.state.meta.lcm_state_str"
														label="lcm state"
														style="display: inline-block; width: 100px"
													>
													</v-text-field>
												</v-col>
											</v-row>
											<v-row>
												<v-col>
													<v-text-field
														readonly
														:value="instance.resources.cpu"
														label="CPU"
														style="display: inline-block; width: 100px"
													>
													</v-text-field>
												</v-col>
												<v-col>
													<v-text-field
														readonly
														:value="instance.resources.ram"
														label="RAM"
														style="display: inline-block; width: 100px"
													>
													</v-text-field>
												</v-col>
												<v-col>
													<v-text-field
														readonly
														:value="instance.resources.drive_size"
														label="drive size"
														style="display: inline-block; width: 100px"
													>
													</v-text-field>
												</v-col>
												<v-col>
													<v-text-field
														readonly
														:value="instance.resources.drive_type"
														label="drive type"
														style="display: inline-block; width: 100px"
													>
													</v-text-field>
												</v-col>
												<v-col>
													<v-text-field
														readonly
														:value="instance.resources.ips_private"
														label="ips private"
														style="display: inline-block; width: 100px"
													>
													</v-text-field>
												</v-col>
												<v-col>
													<v-text-field
														readonly
														:value="instance.resources.ips_public"
														label="ips public"
														style="display: inline-block; width: 100px"
													>
													</v-text-field>
												</v-col>
												<v-col>
													<v-text-field
														readonly
														:value="instance.config.template_id"
														label="template id"
														style="display: inline-block; width: 100px"
													>
													</v-text-field>
												</v-col>
											</v-row>
										</v-expansion-panel-content>
									</v-expansion-panel>
								</v-expansion-panels>
							</v-col>
						</v-row>
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
		opened: [],
		openedInstances: {}
	}),
	props: {
		service: {
			type: Object,
			required: true
		}
	},
	computed: {
		servicesProviders(){
			return this.$store.getters['servicesProviders/all'];
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
		hashpart(hash){
			if(hash)
				return hash.slice(0, 8);

			return "WWWWWWWW"
		},
		location(group){
			const lc = this.servicesProviders.find(el => el.uuid == this.service.provisions[this.service.instancesGroups[group].uuid])
			return lc?.title ?? 'none'
		}
	},
	mounted(){
		this.opened.push(0);
		Object.keys(this.service.instancesGroups).forEach(key => {
			this.$set(this.openedInstances, key, [0]);
		});
	}
}
</script>

<style>

</style>
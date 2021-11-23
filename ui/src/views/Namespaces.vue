<template>
	<div class="namespaces pa-4 flex-wrap">
		<div class="buttons__inline pb-4">

			<v-menu
				offset-y
				transition="slide-y-transition"
				bottom
				:close-on-content-click="false"
			>
				<template v-slot:activator="{ on, attrs }">
					<v-btn
						color="background-light"
						class="mr-2"
						v-bind="attrs"
						v-on="on"
					>
						create
					</v-btn>
				</template>
				<v-card class="d-flex pa-4">
					<v-text-field
						dense
						hide-details
						v-model="newNamespace.title"
						@keypress.enter="createNamespace"
					>
					</v-text-field>
					<v-btn
						:loading="newNamespace.loading"
						@click="createNamespace"
					>
						send
					</v-btn>
				</v-card>
			</v-menu>

			<v-btn
				color="background-light"
				class="mr-8"
				:disabled="selected.length < 1"
				@click="deleteSelectedNamespace"
			>
				delete
			</v-btn>
			<v-btn
				color="background-light"
				class="mr-2"
				:disabled="selected.length < 1"
			>
				join
			</v-btn>
			<v-btn
				color="background-light"
				:disabled="selected.length < 1"
			>
				link
			</v-btn>

		</div>

		<namespaces-table
			v-model="selected"
			single-select
		>

		</namespaces-table>
	</div>
</template>

<script>
import namespacesTable from "@/components/namespaces_table.vue"
import api from "@/api.js"

export default {
	name: "namespaces-view",
	components: {
		"namespaces-table": namespacesTable
	},
	data () {
		return {
			selected: [],
			newNamespace: {
				title: '',
				loading: false
			},
		}
	},
	methods: {
		createNamespace(){
			if(this.newNamespace.title.length < 3) return;
			this.newNamespace.loading = true;
			api.namespaces.create(this.newNamespace.title)
			.then(()=>{
				this.newNamespace.title = '';
				this.$store.dispatch('namespaces/fetch');
			})
			.finally(()=>{
				this.newNamespace.loading = false;
			})
		},
		deleteSelectedNamespace(){
			if(this.selected.length > 0){
				const deletePromices = this.selected.map(el => api.namespaces.delete(el.id));
				Promise.all(deletePromices)
				.then(res => {
					if(res.every(el => el.result)){
						console.log('all ok');
					}

					this.selected = [];
					this.$store.dispatch('namespaces/fetch');
				})
				.catch(err => {
					console.log(err);
				})
			}
		}
	}
}
</script>

<style>

</style>
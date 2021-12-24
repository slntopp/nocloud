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

			<v-dialog
					v-model="joinAccount.modalVisible"
					width="500"
					scrollable
				>
      <template v-slot:activator="{ on, attrs }">
					<v-btn
						color="background-light"
						class="mr-2"
						:disabled="selected.length < 1"
						v-bind="attrs"
						v-on="on"
					>
						join
					</v-btn>
				</template>

				<v-card>
					<v-card-title class="">
						Select Accounts (join)
					</v-card-title>

					<v-card-text>
						<accounts-table
							v-model="joinAccount.selected"
							single-select
						>

						</accounts-table>
						<v-select
							class="mt-5"
							label="access level"
							v-model="joinAccount.data.access"
							:items="joinAccount.accessLevels"
						>
						</v-select>
						<v-select
							label="role"
							v-model="joinAccount.data.role"
							:items="['default', 'owner']"
						>
						</v-select>
					</v-card-text>

					<v-divider></v-divider>

					<v-card-actions>
						<v-spacer></v-spacer>
						<v-btn
							color="primary"
							text
							@click="joinSelectedAcounts"
							:loading="joinAccount.loading"
						>
							Join
						</v-btn>
					</v-card-actions>
				</v-card>
			</v-dialog>


			<v-dialog
					v-model="linkAccount.modalVisible"
					width="500"
					scrollable
				>
      <template v-slot:activator="{ on, attrs }">
					<v-btn
						color="background-light"
						:disabled="selected.length < 1"
						v-bind="attrs"
						v-on="on"
					>
						link
					</v-btn>
				</template>

				<v-card>
					<v-card-title class="">
						Select Accounts (link)
					</v-card-title>

					<v-card-text>
						<accounts-table
							v-model="linkAccount.selected"
						>

						</accounts-table>
					</v-card-text>

					<v-divider></v-divider>

					<v-card-actions>
						<v-spacer></v-spacer>
						<v-btn
							color="primary"
							text
							@click="linkSelectedAcounts"
							:loading="linkAccount.loading"
						>
							Link
						</v-btn>
					</v-card-actions>
				</v-card>
			</v-dialog>

		</div>

		<namespaces-table
			v-model="selected"
			single-select
		>

		</namespaces-table>

		<v-snackbar
			v-model="snackbar.visibility"
			:timeout="snackbar.timeout"
		>
			{{snackbar.message}}
			

      <template v-slot:action="{ attrs }">
        <v-btn
          color="blue"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
		</v-snackbar>
	</div>
</template>

<script>
import namespacesTable from "@/components/namespaces_table.vue"
import accountsTable from "@/components/accounts_table.vue"
import api from "@/api.js"

export default {
	name: "namespaces-view",
	components: {
		"namespaces-table": namespacesTable,
		"accounts-table": accountsTable,
	},
	data () {
		return {
			selected: [],
			newNamespace: {
				title: '',
				loading: false,
				modalVisible: false
			},
			linkAccount: {
				modalVisible: false,
				selected: []
			},
			joinAccount: {
				modalVisible: false,
				selected: [],
				accessLevels: [0,1,2,3],
				data: {
					access: 1,
					role: 'default'
				}
			},
			snackbar: {
				visibility: false,
				message: '',
				timeout: 3000,
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
				const deletePromices = this.selected.map(el => api.namespaces.delete(el.uuid));
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
		},
		linkSelectedAcounts(){
			if(this.selected.length > 0){
				const namespace = this.selected[0];
				const linkPromices = this.linkAccount.selected.map(account => api.namespaces.link(namespace.uuid, account.uuid));
				this.linkAccount.loading = true;
				Promise.all(linkPromices)
				.then(res => {
					console.log(res);
					if(res.every(el => el.result)){
						console.log('all ok');
						this.linkAccount.modalVisible = false;
						
						this.snackbar.message = "Successfully."
						this.snackbar.visibility = true;
					}

					this.selected = [];
					this.$store.dispatch('namespaces/fetch');
				})
				.catch(err => {
					console.log(err);
					this.snackbar.message = "Something went wrong... Try later."
					this.snackbar.visibility = true;
				})
				.finally(()=>{
					this.linkAccount.selected = [];
					this.linkAccount.loading = false;
				})
			}
		},
		joinSelectedAcounts(){
			api.namespaces.join(this.selected[0].uuid, {
				account: this.joinAccount.selected[0].uuid,
				access: this.joinAccount.data.access,
				role: this.joinAccount.data.role,
			})
			.then(res => {
				console.log(res);
				if(res.every(el => el.result)){
					console.log('all ok');
					this.joinAccount.modalVisible = false;
					
					this.snackbar.message = "Successfully."
					this.snackbar.visibility = true;
				}

				this.selected = [];
				this.$store.dispatch('namespaces/fetch');
			})
			.catch(err => {
				// this.snackbar.message = "Something went wrong... Try later."
				this.snackbar.message = err.response?.data?.message ?? "Something went wrong... Try later."
				this.snackbar.visibility = true;
			})
			.finally(()=>{
				this.joinAccount.selected = [];
				this.joinAccount.loading = false;
			})
		}
	},
}
</script>

<style>

</style>
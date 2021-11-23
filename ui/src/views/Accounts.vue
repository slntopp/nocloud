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
						v-model="newAccount.title"
						@keypress.enter="createAccount"
					>
					</v-text-field>
					<v-btn
						:loading="newAccount.loading"
						@click="createAccount"
					>
						send
					</v-btn>
				</v-card>
			</v-menu>

			<v-btn
				color="background-light"
				class="mr-8"
				:disabled="selected.length < 1"
				@click="deleteSelectedAccount"
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

		<accounts-table
			v-model="selected"
		>
		</accounts-table>
	</div>
</template>

<script>
import accountsTable from "@/components/accounts_table.vue"
import api from "@/api.js"

export default {
	name: "accounts-view",
	components: {
		"accounts-table": accountsTable,
	},
	data () {
		return {
			selected: [],
			newAccount: {
				title: '',
				loading: false
			},
		}
	},
	computed: {
		tableData(){
			return this.$store.getters['accounts/all'];
		}
	},
	methods: {
		createAccount(){
			if(this.newAccount.title.length < 3) return;
			this.newAccount.loading = true;
			api.accounts.create({
					"title": "testo",
					"auth": {
							"type": "standard",
							"data": ["testo", "pesto"]
					},
					"access": 3
			})
			.then(()=>{
				this.newAccount.title = '';
				this.$store.dispatch('accounts/fetch');
			})
			.finally(()=>{
				this.newAccount.loading = false;
			})
		},
		deleteSelectedAccount(){
			if(this.selected.length > 0){
				const deletePromices = this.selected.map(el => api.accounts.delete(el.id));
				Promise.all(deletePromices)
				.then(res => {
					if(res.every(el => el.result)){
						console.log('all ok');
					}

					this.selected = [];
					this.$store.dispatch('accounts/fetch');
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
<template>
	<div class="servicesProviders pa-4 flex-wrap">
		<div class="buttons__inline pb-4">

			<v-btn
				color="background-light"
				class="mr-8"
				:disabled="selected.length < 1"
				@click="deleteSelectedServicesProviders"
			>
				delete
			</v-btn>
		</div>

		<services-providers
			v-model="selected"
		>

		</services-providers>

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
import api from "@/api.js"
import servicesProviders from "@/components/servicesproviders_table.vue"

export default {
	name: "namespaces-view",
	components: {
		servicesProviders
	},
	data () {
		return {
			selected: [],
			snackbar: {
				visibility: false,
				message: '',
				timeout: 3000,
			},
		}
	},
	methods: {
		deleteSelectedServicesProviders(){
			if(this.selected.length > 0){
				const deletePromices = this.selected.map(el => api.delete(`/sp/${el.uuid}`));
				Promise.all(deletePromices)
				.then(res => {
					if(res.every(el => el.result)){
						console.log('all ok');
					}

					this.selected = [];
					this.$store.dispatch('servicesProviders/fetch');
				})
				.catch(err => {
					console.log(err);
				})
			}
		}
	},
	created(){
		this.$store.dispatch('servicesProviders/fetch')
	}
}
</script>

<style>

</style>
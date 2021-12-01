<template>
	<div class="servicesProviders pa-4 flex-wrap">
		<div class="buttons__inline pb-4">
			
		</div>

		<!-- <namespaces-table
			v-model="selected"
			single-select
		>

		</namespaces-table> -->

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
// import api from "@/api.js"

export default {
	name: "namespaces-view",
	components: {
	},
	data () {
		return {
			snackbar: {
				visibility: false,
				message: '',
				timeout: 3000,
			},
		}
	},
	methods: {
	},
	created(){
		this.$store.dispatch('serviceProviders/fetch')
	}
}
</script>

<style>

</style>
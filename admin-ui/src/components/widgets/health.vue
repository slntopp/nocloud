<template>
  <widget
    title="Health"
    :loading="loading"
  >
    <v-alert

      v-bind="alertAttrs"
    >
      {{alertText}}
    </v-alert>

		<v-btn
			@click="checkHealth"
		>
			retry
		</v-btn>
  </widget>
</template>

<script>
import widget from "./widget.vue";
import api from "@/api.js"

export default {
  name: 'health-widget',
  components: {
    widget
  },
  data: ()=>({
    loading: false,
    isHealthOk: false
  }),
  computed: {
    alertText(){
      if(this.loading){
        return 'Loading...'
      }

      return this.isHealthOk ? "All systems works just fine" : "Something went wrong";
    },
    alertAttrs(){
      if(this.loading){
        return {
          icon: "mdi-help-circle",
          color: "grey darken-1"
        }
      }

      if(this.isHealthOk){
        return {
          type: "success"
        }
      } else {
        return {
          type: "error"
        }
      }
    }
  },
  created(){
		this.checkHealth();
  },
	methods: {
		checkHealth(){
			this.loading = true;
			api.health.ping()
			.then(res => {
				if(res.response == "PONG"){
					this.isHealthOk = true;
				}
			})
			.catch(err => {
				console.error(err);
			})
			.finally(()=>{
				this.loading = false;
			})
		}
	}
}
</script>

<style>

</style>
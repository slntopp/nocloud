<template>
  <widget
    title="Services"
    :loading="loading"
  >
    <v-alert
			type="error"
			v-if="err"
    >
      {{err.name}}: {{err.message}}
    </v-alert>

		<v-list
			dense
			class="mb-4"
			color="transparent"
		>
			<v-list-item v-for="item in state" :key="item.service" class="px-0">
				<v-list-item-content>
					<v-list-item-title>
						{{item.status.service}}{{ item.routine ? ": " + item.routine : "" }}
					</v-list-item-title>

					<v-list-item-subtitle v-if="lastExecution(item.lastExecution)">
						Last execution: {{lastExecution(item.lastExecution)}}
					</v-list-item-subtitle>
				</v-list-item-content>
				
				<v-list-item-icon>
					<v-chip
						small
						:color="item.status.status == 'RUNNING' ? 'success' : 'error'"
					>
						{{item.status.status}}
					</v-chip>
				</v-list-item-icon>
			</v-list-item>

		</v-list>

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
  name: 'services-widget',
  components: {
    widget
  },
  data: ()=>({
    loading: false,
		err: null,
    state: {}
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
			api.health.routines()
			.then(res => {
				this.state = res.routines;
				this.err = null;
			})
			.catch(err => {
				console.error(err);
				this.err = err;
			})
			.finally(()=>{
				this.loading = false;
			})
		},
		lastExecution(time){
			if(!time) return ""

			const date = new Date(time)
			// console.log(date)
			// const month = date.getMonth();
			// const day = date.getDate();

			return new Intl.DateTimeFormat().format(date)
		}
	}
}
</script>

<style>

</style>
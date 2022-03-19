<template>
  <widget
    title="Routines"
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
						Last execution: {{ts2str(item.lastExecution)}}
					</v-list-item-subtitle>
				</v-list-item-content>
				
				<v-list-item-icon>
					<v-chip
						small
						:color="chipsColor(item.status.status)"
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

const formatDateNumber = (num, n = 2) => {
  num = num.toString();
  while (num.length < n) {
    num = "0" + num;
  }
  return num;
};
const date2Object = (date) => {
  return {
    day: date.getDate(),
    month: date.getMonth(),
    year: date.getFullYear(),
    hour: date.getHours(),
    minute: date.getMinutes(),
    second: date.getSeconds(),
  };
};

export default {
  name: 'routines-widget',
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
				this.state = res.routines.filter(el => el.status.status !== 'NOEXIST');
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
			return new Intl.DateTimeFormat().format(date)
		},
		chipsColor(state){
			switch (state) {
				case 'RUNNING':
					return 'success'
				case 'INTERNAL':
					return 'error'
				case 'STOPPED':
					return 'warning'
				case 'NOEXIST':
					return 'gray'
		
				default:
					return 'gray';
			}
		},
    ts2str(ts) {
      let today = date2Object(new Date());
      let date = new Date(Date.parse(ts));
      date = date2Object(date);
      let result = "";
      // Day month section
      if (
        Number(date.month) == today.month &&
        Number(date.year) == today.year
      ) {
        if (Number(date.day) == today.day) {
          result += "Today";
        } else if (Number(date.day) == today.day - 1) {
          result += "Yesterday";
        }
      } else {
        result +=
          formatDateNumber(date.day) + "." + formatDateNumber(date.month + 1);
      }
      // Year section
      if (Number(date.year) != today.year) {
        result += `.${formatDateNumber(date.year, 4)}`;
      }
      // Time section
      result += ` ${formatDateNumber(date.hour)}:${formatDateNumber(
        date.minute
      )}:${formatDateNumber(date.second)}`;
      return result;
    },
	}
}
</script>

<style>

</style>

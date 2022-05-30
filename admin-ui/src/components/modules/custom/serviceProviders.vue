<template>
	<div class="te">
		<v-row>
			<v-col cols="3">
				<v-subheader>
					secrets
				</v-subheader>
			</v-col>
			<v-col cols="9">
				<json-editor
          :json="secrets"
          @changeValue="(data) => changeHandler('secrets', data)"
        />
			</v-col>
		</v-row>
		<v-row>
			<v-col cols="3">
				<v-subheader>
					vars
				</v-subheader>
			</v-col>
			<v-col cols="9">
				<json-editor
          :json="vars"
          @changeValue="(data) => changeHandler('vars', data)"
        />
			</v-col>
		</v-row>
	</div>
</template>

<script>
import JsonEditor from '@/components/JsonEditor';

export default {
	name: "servicesProviders-create-custom",
  components: { JsonEditor },
	props: {
		secrets: {
			type: Object,
			default: () => ({})
		},
		vars: {
			type: Object,
			default: () => ({})
		},
		passed: {
			type: Boolean,
			default: false
		}
	},
	data: () => ({
		errors: {
			secrets: [],
			vars: []
		}
	}),
	methods: {
		isJSON(str){
			try{
				JSON.parse(str);
				return true;
			} catch {
				return false;
			}
		},
		changeHandler(input, data){
			const isJSON = this.isJSON(JSON.stringify(data));

			if(!isJSON){
				this.errors[input] = ["is not JSON"];
			} else {
				this.errors[input] = [];
			}
			this.$emit(`change:${input}`, data)
			this.$emit(`passed`, Object.values(this.errors).every(el => el.length == 0))
		}
	}
}
</script>

<style scoped lang="scss">
.te > * {
  margin-top: 30px;
}
</style>
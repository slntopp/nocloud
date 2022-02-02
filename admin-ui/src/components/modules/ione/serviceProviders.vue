<template>
	<div class="sp-ione">
		<v-row v-for="field in Object.keys(fields)" :key="field" align="center">
			<v-col cols="3">
				<v-subheader>
					{{fields[field].subheader || field}}

					<v-tooltip
						v-if="field == 'host' && hostWarning"
						bottom
						color="warning"
					>
						<template v-slot:activator="{ on, attrs }">
							<v-icon
								class="ml-2"
								color="warning"
								v-bind="attrs"
								v-on="on"
							>
								mdi-alert-outline
							</v-icon>
						</template>
						
						<span>Non-standard RPC path: "{{hostWarning}}" instead of "/RPC2"</span>
					</v-tooltip>
				</v-subheader>
			</v-col>

			<v-col
				cols="9"
			>
				<v-text-field
					@change="(data) => changeHandler(field, data)"
					:value="getValue(field)"
					:label="fields[field].label"
					:rules="fields[field].rules"
					:error-messages="errors[field]"
					:type="fields[field].type"
					v-bind="fields[field].bind || {}"
				></v-text-field>
			</v-col>
		</v-row>
	</div>
</template>

<script>

function isJSON(str){
	try{
		JSON.parse(str);
		return true;
	} catch {
		return false;
	}
}

// const domainRegex = /^((https?:\/\/)|(www.))(?:(\.?[a-zA-Z0-9-]+){1,}|(\d+\.\d+.\d+.\d+))(\.[a-zA-Z]{2,})?(:\d{4})?\/?$/;
// const domainRegex = /^(https?):\/\/(((?!-))(xn--|_{1,1})?[a-z0-9-]{0,61}[a-z0-9]{1,1}\.)*(xn--)?([a-z0-9][a-z0-9-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$/
export default {
	name: "servicesProviders-create-ione",
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
		hostWarning: false,
		errors: {
			host: [],
			username: [],
			password: [],
			group: [],
			schedule: [],
			schedule_ds: [],
			public_ip_pool: [],
			private_vnets_pool: [],
		},
		values: {
			host: "",
			username: "",
			password: "",
			group: "",
			schedule: "",
			schedule_ds: "",
			public_ip_pool : "",
			private_vnets_pool: "",
		},
		fields: {
			host: {
				type: 'text',
				rules: [
					(value) => !!value || 'Field is required',
					(value) => {
						try{
							new URL(value)
							return true
						} catch(err){
							return 'Is not valid domain'
						}
					}
				],
				label: "example.com"
			},
			username: {
				type: 'text',
				rules: [
					(value) => !!value || 'Field is required'
				],
				label: "username"
			},
			password: {
				type: 'password',
				rules: [
					(value) => !!value || 'Field is required'
				],
				label: "password"
			},
			group: {
				type: 'number',
				rules: [
					(value) => !!value || 'Field is required',
					(value) => value == Number(value) || 'wrong group id'
				],
				label: "100",
				bind: {
					min: 0
				}
			},
			schedule: {
				type: 'text',
				isJSON: true,
				rules: [
					(value) => !!value || 'Field is required',
					(value) => isJSON(value) || "is not valid JSON"
				],
				label: "JSON"
			},
			schedule_ds: {
				type: 'text',
				subheader: "schedule data store",
				isJSON: true,
				rules: [
					(value) => !!value || 'Field is required',
					(value) => isJSON(value) || "is not valid JSON"
				],
				label: "JSON"
			},
			public_ip_pool: {
				type: 'number',
				subheader: "public ip pool",
				rules: [
					(value) => !!value || 'Field is required',
				],
				label: "pip",
				bind: {
					min: 0
				}
			},
			private_vnets_pool: {
				type: 'number',
				subheader: "private vnets pool",
				rules: [
					(value) => !!value || 'Field is required',
				],
				label: "pvp",
				bind: {
					min: 0
				}
			},
		}
	}),
	methods: {
		requiredField(value){
			return !!value || 'Field is required';
		},
		isDomain(value){
			const reg = /^(https?):\/\/(((?!-))(xn--|_{1,1})?[a-z0-9-]{0,61}[a-z0-9]{1,1}\.)*(xn--)?([a-z0-9][a-z0-9-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$/
			return !!value.match(reg) || 'Is not valid domain';
		},
		isNumber(value){
			return value === Number(value) || 'Is not valid domain';
		},
		changeHandler(input, data){
			this.values[input] = data;
			let errors = {};

			Object.keys(this.fields).forEach(fieldName => {
				this.fields[fieldName].rules.forEach(rule => {
					const result = rule(this.values[fieldName]);
					if(typeof result == 'string'){
						this.errors[fieldName] = [result];
						errors[fieldName] = result;
					} else {
						this.errors[fieldName] = [];
					}
				})
			});

			// console.error(`errors`, errors);

			const secrets = {}
			if(this.values.host){
				secrets.host = this.values.host;
			}
			if(this.values.username){
				secrets.user = this.values.username;
			}
			if(this.values.password){
				secrets.pass = this.values.password;
			}
			if(this.values.group){
				secrets.group = +this.values.group;
			}

			const vars = {}
			if(this.values.schedule){
				if(isJSON(this.values.schedule)){
					vars.sched = JSON.parse(this.values.schedule);
					errors.sched = []
				} else {
					errors.sched = ["is not valid JSON"]
				}
			}
			if(this.values.schedule_ds){
				if(isJSON(this.values.schedule_ds)){
					vars.sched_ds = JSON.parse(this.values.schedule_ds);
					errors.sched_ds = []
				} else {
					errors.sched_ds = ["is not valid JSON"]
				}
			}
			if(this.values.public_ip_pool){
				vars.public_ip_pool = {value: {default: +this.values.public_ip_pool}}
			}
			if(this.values.private_vnets_pool){
				vars.private_vnets_pool = {value: {default: +this.values.private_vnets_pool}}
			}

			const result = {
				secrets,
				vars
			}

			// console.error(`errors`, errors, Object.keys(errors).length);

			this.$emit(`change:secrets`, secrets)
			this.$emit(`change:vars`, vars)
			this.$emit(`change:full`, result)
			this.$emit(`passed`, Object.keys(errors).length == 0)
		},
		getValue(fieldName){
			switch (fieldName) {
				case 'domain':
					return this.secrets.host
				case 'username':
					return this.secrets.user
				case 'password':
					return this.secrets.pass
				case 'group':
					return this.secrets.group
				case 'schedule':
					return JSON.stringify(this.vars.sched)
				case 'schedule_ds':
					return JSON.stringify(this.vars.sched_ds)
				case 'public_ip_pool':
					return this.vars.public_ip_pool?.value?.default ?? ""
				case 'private_vnets_pool':
					return this.vars.private_vnets_pool?.value?.default ?? ""
				default:
					return "";
			}
		}
	},
	watch: {
		"secrets.host"(newVal){
			try{
				const url = new URL(newVal);
				if(url.pathname !== "/RPC2")
					this.hostWarning = url.pathname
				else
					this.hostWarning = false
			}
			catch(err){
				this.hostWarning = false
			}
		}
	}
}
</script>

<style>

</style>
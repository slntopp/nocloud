<template>
	<v-card
		color="background-light"
	>
		<v-card-title>
			<v-row
				justify="space-between"
			>
				<v-col>
					NoCloud Tunnelmesh
				</v-col>
				<v-col cols=1>
					<v-icon
						large
						color="primary"
						@click="$emit('remove')"
					>
						mdi-close
					</v-icon>
				</v-col>
			</v-row>
		</v-card-title>
		<v-card-text>
			<v-row align="center">
				<v-col cols="2">
					<v-subheader>
						hostname
					</v-subheader>
				</v-col>

				<v-col
					cols="10"
				>
					<v-text-field
						@change="(data) => emitNewHostname(data)"
						:value="hostname"
						label="enter at main form"
						:rules="rules"
						:error-messages="errors"
						readonly
					></v-text-field>
				</v-col>
			</v-row>
			<v-row align="center">
				<v-col cols="2">
					<v-subheader>
						cerificate
					</v-subheader>
				</v-col>

				<v-col
					cols="10"
				>
					<v-file-input
						v-model="cert"
						accept="application/x-x509-ca-cert"
						label="File input"
						truncate-length="15"
						@input="certInput"
					></v-file-input>
					<v-btn @click="certInput">
						s
					</v-btn>
				</v-col>
			</v-row>
		</v-card-text>
	</v-card>
</template>

<script>
import { PEM, ASN1 } from '@sardinefish/asn1';
import createHash from "sha256-uint8array";

export default {
	name: "tunnelmesh-extention",
	data(){
		return {
			rules: [
				(value) => !!value || 'Field is required',
				(value) => !!value.match(/^((https?:\/\/)|(www.))(?:(\.?[a-zA-Z-]+){1,}|(\d+\.\d+.\d+.\d+))(\.[a-zA-Z]{2,})?(:\d{4})?\/?$/) || 'Is not valid domain'
			],
			errors: [],

			fingerprint: "",
			cert: null
		}
	},
	props: {
		provider: {
			type: Object,
			default: () => ({})
		},
		data: {
			type: Object,
			default: () => ({})
		}
	},
	computed: {
		hostname() {
			if(this.provider?.secrets?.host){
				return this.provider?.secrets?.host
			} else {
				return ""
			}
		}
	},
	methods: {
		emitNewHostname(hostname){
			const temp = {...this.secrets};
			temp.secrets = {};
			temp.secrets.host = hostname;
			console.log(temp);
			this.$emit('change:provider', temp)
			this.$emit('change:data', {hostname: hostname, fingerprint: this.fingerprint})
		},
		certInput(){
			// const pems = PEM.parse(this.cert);
			// const ans1 = ASN1.fromDER(pems[0].body);
			this.cert.text()
			.then(res=> {
				// console.log(res);
				
				const buf = Buffer.from(res)
				console.log(`buf`, buf);
				console.log(`cert`, this.cert);
				const pems = PEM.parse(buf);
				console.log(`pems`, pems);
				const ans1 = ASN1.fromDER(pems[0].body);
				console.log(`asn`, ans1);
				console.log(`asnhash`, createHash(ans1.bytes));
			})
			PEM, ASN1
			// console.log(`pems`, pems);
			// console.log(`asn`, ans1);
		}
	},
	mounted() {
		let host = '';
		if(this.provider?.secrets?.host){
			host = this.provider?.secrets?.host;
		} else {
			this.$emit('change:provider', {...this.secrets, secrets: {host}})
		}
		this.$emit('change:data', {hostname: host, fingerprint: this.fingerprint})
	}
}
</script>

<style>

</style>
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
					class="d-flex"
				>
					<v-file-input
						:class="{'flex-grow-0 flex-shrink-1': !!fingerprint}"
						v-model="cert"
						accept=".crt"
						label="File input"
						truncate-length="15"
						@change="certInput"
						:hide-input="!!fingerprint"
					></v-file-input>
					<v-text-field
						v-if="fingerprint"
						class="flex-grow-1 flex-shrink-0"
						readonly
						:error-messages="errors"
						label="your certificate hash"
						:value="fingerprint"
					>

					</v-text-field>
				</v-col>
			</v-row>
		</v-card-text>
	</v-card>
</template>

<script>
import { PEM, ASN1 } from '@sardinefish/asn1';
import { createHash } from "sha256-uint8array";

export default {
	name: "tunnelmesh-extention",
	data(){
		return {
			rules: [
				(value) => !!value || 'Field is required',
			],
			errors: [],

			fingerprint: "",
			cert: null,
			certError: []
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
				const domainname = new URL(this.provider?.secrets?.host)
				return domainname.hostname;
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
			this.$emit('change:data', {hostname: this.hostname, fingerprint: this.fingerprint})
		},
		certInput(){
			this.cert.text()
			.then(res=> {
				const buf = Buffer.from(res)
				const pems = PEM.parse(buf);
				const ans1 = ASN1.fromDER(pems[0].body);
				const hashdec = createHash().update(ans1._der).digest()
				let hash = ''
				for (const el of hashdec) {
					hash += el.toString(16);
				}
				this.fingerprint = hash;
				this.$emit('change:data', {hostname: this.hostname, fingerprint: this.fingerprint})
			})
			.catch(() => {
				this.certError = ['Wrong file'];
			})
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
	},
	watch: {
		hostname(){
			this.$emit('change:data', {hostname: this.hostname, fingerprint: this.fingerprint})
		}
	}
}
</script>

<style>

</style>
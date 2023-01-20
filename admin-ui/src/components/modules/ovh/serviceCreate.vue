<template>
	<div class="module">
		<v-card
			class="mb-4 pa-2"
			color="background"
			elevation="0"
			v-for="(instance, index) of instances"
      :key="index"
		>
			<v-row>
				<v-col cols="6">
					<v-text-field
						label="title"
						v-model="instance.title"
            :rules="rules.req"
						@change="(value) => setValue(index + '.title', value)"
					/>
				</v-col>
				<v-col class="d-flex justify-end">
					<v-btn @click="() => remove(index)">Remove</v-btn>
				</v-col>
			</v-row>

			<v-row>
				<v-col cols="6">
					<v-select
						label="model"
						v-model="config[index].flavor"
            :items="flavors"
            :rules="rules.req"
            :loading="isFlavorsLoading"
						@change="(value) => setValue(index + '.config.flavorId', value)"
					/>
				</v-col>
				<v-col cols="6">
					<v-select
						label="region"
						v-model="instance.config.region"
						:items="regions"
            :rules="rules.req"
            :disabled="!config[index].flavor"
            @change="(value) => setValue(index + '.config.region', value)"
					/>
				</v-col>
				<v-col cols="6">
					<v-select
						label="OS type"
						v-model="config[index].os"
            :items="['baremetal-linux', 'bsd', 'linux', 'windows']"
					/>
				</v-col>
				<v-col cols="6">
					<v-select
						label="OS"
            item-text="name"
            item-value="id"
						v-model="instance.config.imageId"
						:items="images"
            :rules="rules.req"
            :loading="isOSLoading"
            :disabled="!instance.config.flavorId"
            @change="(value) => setValue(index + '.config.imageId', value)"
					/>
				</v-col>
				<v-col cols="6">
					<v-text-field
						label="post-installation script"
						v-model="instance.config.userData"
						@change="(value) => setValue(index + '.config.userData', value)"
					/>
				</v-col>
				<v-col cols="6">
					<v-select
						label="price model"
            item-text="title"
            item-value="uuid"
            v-model="instance.plan"
            :items="plans.list"
            :rules="planRules"
						@change="(value) => setValue(index + '.billing_plan', value)"
					/>
				</v-col>
				<v-col cols="6">
					<v-select
						label="product"
            v-model="instance.productTitle"
            v-if="instance.products?.length > 0"
            :items="instance.products"
						@change="(value) => setValue(index + '.product', value)"
					/>
				</v-col>
				<v-col cols="6" class="d-flex align-center">
          Payment:
					<v-switch
            class="d-inline-block ml-2"
						v-model="instance.config.monthlyBilling"
						:label="(instance.config.monthlyBilling) ? 'monthly' : 'hourly'"
						@change="(value) => setValue(index + '.config.monthlyBilling', value)"
					/>
				</v-col>
			</v-row>
		</v-card>
		<v-row>
			<v-col class="d-flex justify-center">
				<v-btn
          small
          class="mx-2"
          color="background"
          :disabled="isDisabled"
          @click="addInstance"
        >
					<v-icon dark>mdi-plus-circle-outline</v-icon>
					add instance
				</v-btn>
			</v-col>
		</v-row>
	</div>
</template>

<script>
import api from '@/api.js';

export default {
	name: 'ovh-create-service-module',
	props: ['instances-group', 'plans', 'planRules'],
	data: () => ({
		defaultItem: {
			"title": "instance",
			"config": {
				"type": "vm",
        "flavorId": null,
        "region": null,
        "imageId": null
			},
      "billing_plan": {}
		},
    rules: {
      req: [(v) => !!v || "required field"]
    },
    config: {},

    isFlavorsLoading: false,
    isOSLoading: false,
    allFlavors: [],
    flavors: [],
    regions: [],
    images: []
	}),
	methods: {
    addProducts(instance) {
      const { plan, billing_plan } = instance;
      const products = this.plans.list.find((el) => el.uuid === plan.uuid)
        ?.products;

      if (billing_plan.kind === 'STATIC') {
        instance.products = [];
        Object.values(products || {}).forEach(({ title }) => {
          instance.products.push(title);
        });
      } else {
        delete instance.products;
        delete instance.product;
      }
    },
		addInstance(){
			const item = JSON.parse(JSON.stringify(this.defaultItem));
			const data = JSON.parse(this.instancesGroup);
      const i = data.body.instances.length;

			item.title += "#" + (i + 1);
      this.config[i] = { flavor: null, os: null };
			data.body.instances.push(item);
			this.change(data);
		},
    getOS({ region, flavor, os }){
      const data = JSON.parse(this.instancesGroup);

      this.isOSLoading = true;
      api.post(`/sp/${data.sp}/invoke`, {
        method: 'images',
        params: { flavor, region, os }
      })
        .then(({ meta }) => {
          this.images = meta.result;
        })
        .finally(() => {
          this.isOSLoading = false;
        });
    },
		remove(index){
			const data = JSON.parse(this.instancesGroup);

			data.body.instances.splice(index, 1);
			this.change(data);
		},
		setValue(path, val){
			const data = JSON.parse(this.instancesGroup)
      const i = path.split('.')[0]

      if (path.includes('plan')) {
        const plan = this.plans.list.find(({ uuid }) => val === uuid)
        const j = plan.title.length - 14

        data.body.instances[i].plan = val
        val = { ...plan, title: plan.title.slice(0, j) }
      }
      if (path.includes('product')) {
        const plan = data.body.instances[i].billing_plan
        const [product] = Object.entries(plan.products)
          .find(([, prod]) => prod.title === val)

        data.body.instances[i].productTitle = val
        val = product
      }
      if (path.includes('flavor')) {
        this.allFlavors.forEach((el) => {
          if (el.name === val) this.regions.push(el.region)
        })
        val = null
      }
      if (path.includes('region')) {
        const { config } = data.body.instances[i];

        config.flavorId = this.allFlavors.find((el) =>
          (el.region === val) && (el.name === this.config[i].flavor)
        )?.id;
        this.config[i].region = val;
        this.getOS(this.config[i]);
      }

			setToValue(data.body.instances, val, path)
      if (path.includes('plan')) this.addProducts(data.body.instances[i])
			this.change(data)
		},
		change(data){
			this.$emit('update:instances-group', JSON.stringify(data))
		}
	},
	computed: {
		instances() {
			return JSON.parse(this.instancesGroup).body.instances;
		},
    isDisabled() {
      const isOVH = JSON.parse(this.instancesGroup).body.type === 'ovh';
      const isSpEmpty = JSON.parse(this.instancesGroup).sp;

      return isOVH && !isSpEmpty;
    }
	},
	created() {
		const data = JSON.parse(this.instancesGroup);

		if (!data.body.instances) data.body.instances = [];
    else {
      data.body.instances.forEach((inst, i, arr) => {
        if (inst.billingPlan) {
          arr[i].billing_plan = inst.billingPlan;
          delete arr[i].billingPlan;
        }
        arr[i].plan = inst.billing_plan.uuid;
      });
    }

		this.change(data);
	},
  watch: {
    instances() {
      const data = JSON.parse(this.instancesGroup);

      if (data.body.type !== 'ovh') return;
      if (this.flavors.length > 0) return;
      if (this.regions.length > 0) return;
      if (data.body.instances.length < 1) return;

      this.isFlavorsLoading = true;
      api.post(`/sp/${data.sp}/invoke`, { method: 'flavors' })
        .then(({ meta }) => {
          this.allFlavors = meta.result;
          meta.result.forEach((el) => {
            if (this.flavors.includes(el.name)) return;
            this.flavors.push(el.name);
          });
        })
        .finally(() => {
          this.isFlavorsLoading = false;
        });
    }
  }
}

function setToValue(obj, value, path) {
	path = path.split('.');
	let i;
	for (i = 0; i < path.length - 1; i++){
		if(path[i] === "__proto__" || path[i] === "constructor")
			throw new Error("Can't use that path because of: " + path[i]);
		obj = obj[path[i]];
	}

	obj[path[i]] = value;
}
</script>

<template>
	<div class="module">
		<v-card
			class="mb-4 pa-2"
			color="background"
			elevation="0"
			v-for="(instance, index) of instances"
      :key="index"
      :id="instance.uuid"
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
						label="price model"
            item-text="title"
            item-value="uuid"
            v-model="instance.plan"
            :items="plans.list"
            :rules="planRules"
						@change="(value) => setValue(index + '.billing_plan', value)"
					/>
				</v-col>
				<v-col cols="6" v-if="instance.products?.length > 0">
					<v-select
						label="product"
            v-model="instance.productTitle"
            :items="instance.products"
						@change="(value) => setValue(index + '.product', value)"
					/>
				</v-col>
				<v-col cols="6">
					<v-select
						label="tariff"
            item-text="title"
            item-value="code"
						v-model="instance.config.planCode"
            :items="flavors"
            :rules="rules.req"
            :loading="isFlavorsLoading"
						@change="(value) => setValue(index + '.config.planCode', value)"
					/>
				</v-col>
				<v-col cols="6">
					<v-select
						label="region"
						v-model="instance.config.datacenter"
						:items="regions"
            :rules="rules.req"
            :disabled="!instance.config.planCode"
            @change="(value) => setValue(index + '.config.datacenter', value)"
					/>
				</v-col>
				<v-col cols="6">
					<v-select
						label="OS"
						v-model="instance.config.os"
						:items="images"
            :rules="rules.req"
            :disabled="!instance.config.planCode"
            @change="(value) => setValue(index + '.config.os', value)"
					/>
				</v-col>
				<v-col cols="6" class="d-flex align-center">
          Payment:
					<v-switch
            class="d-inline-block ml-2"
            true-value="P1Y"
            false-value="P1M"
						v-model="instance.config.duration"
						:label="(instance.config.duration === 'P1Y') ? 'yearly' : 'monthly'"
						@change="(value) => setValue(index + '.config.duration', value)"
					/>
				</v-col>
				<v-col cols="6" class="d-flex align-center">
          Existing:
					<v-switch
            class="d-inline-block ml-2"
						v-model="instance.data.existing"
						@change="(value) => setValue(index + '.data.existing', value)"
					/>
				</v-col>
				<v-col cols="6" class="d-flex align-center" v-if="instance.data.existing">
					<v-text-field
            label="VPS name"
						v-model="instance.config.vpsName"
            :rules="rules.req"
						@change="(value) => setValue(index + '.data.vpsName', value)"
					/>
				</v-col>
			</v-row>
		</v-card>
		<v-row>
			<v-col class="d-flex justify-center">
				<v-btn
          small
          id="button"
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
			title: "instance",
			config: {
				type: "vps",
        planCode: null,
        datacenter: null,
        os: null,
        duration: 'P1M',
        pricingMode: 'default'
			},
      data: { existing: false },
      billing_plan: {}
		},
    rules: {
      req: [(v) => !!v || "required field"]
    },

    isFlavorsLoading: false,
    flavors: [],
    regions: [],
    images: [],
    meta: {}
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
			data.body.instances.push(item);
			this.change(data);
		},
		remove(index){
			const data = JSON.parse(this.instancesGroup);

			data.body.instances.splice(index, 1);
			this.change(data);
		},
    fetchPlans() {
      const data = JSON.parse(this.instancesGroup);

      if (data.body.type !== 'ovh') return;
      if (this.regions.length > 0) return;
      if (this.instances.length > 0) return;

      this.isFlavorsLoading = true;
      api.post(`/sp/${data.sp}/invoke`, { method: 'get_plans' })
        .then(({ meta }) => { this.meta = meta })
        .finally(() => { this.isFlavorsLoading = false });
    },
		setValue(path, val){
			const data = JSON.parse(this.instancesGroup)
      const i = path.split('.')[0]

      if (path.includes('billing_plan')) {
        const plan = this.plans.list.find(({ uuid }) => val === uuid)
        const title = plan.title.split(' ')

        title.pop()
        this.flavors = Object.keys(plan.products).map((el) => ({
          code: el.split(' ')[1],
          title: plan.products[el].title
        }))

        data.body.instances[i].plan = val
        val = { ...plan, title: title.join(' ') }
      }

      if (path.includes('product')) {
        const plan = data.body.instances[i].billing_plan
        const [product] = Object.entries(plan.products)
          .find(([, prod]) => prod.title === val)

        data.body.instances[i].productTitle = val
        val = product
      }

      if (path.includes('planCode')) {
        const plan = this.meta.catalog.plans
          .find(({ planCode }) => planCode === val)

        plan.configurations.forEach((el) => {
          el.values.sort()
          if (el.name.includes('os')) this.images = el.values
          if (el.name.includes('datacenter')) this.regions = el.values
        })
      }

      if (path.includes('duration')) {
        this.setValue(`${i}.config.pricingMode`, (val) ? 'default' : 'upfront12')
      }

			setToValue(data.body.instances, val, path)
      if (path.includes('billing_plan')) this.addProducts(data.body.instances[i])
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
      const group = JSON.parse(this.instancesGroup);

      return group.body.type === 'ovh' && !group.sp;
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
    instances() { this.fetchPlans() }
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

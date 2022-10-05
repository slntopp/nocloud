<template>
  <div class="module">
    <v-card
      v-for="(instance, index) in instances"
      :key="index"
      class="mb-4 pa-2"
      elevation="0"
      color="background"
    >
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="(newVal) => setValue(index + '.title', newVal)"
            label="title"
            v-model="instance.title"
          />
        </v-col>
        <v-col class="d-flex justify-end">
          <v-btn @click="() => remove(index)"> remove </v-btn>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) =>
                setValue(index + '.resources.user.first_name', +newVal)
            "
            label="first name"
            v-model="instance.resources.user.first_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.last_name', newVal)
            "
            label="last name"
            v-model="instance.resources.user.last_name"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.org_name', +newVal)
            "
            label="organization name"
            v-model="instance.resources.user.org_name"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.address1', +newVal)
            "
            label="address1"
            v-model="instance.resources.user.address1"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.address2', +newVal)
            "
            label="address2"
            v-model="instance.resources.user.address2"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.city', +newVal)
            "
            label="city"
            v-model="instance.resources.user.city"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.country', +newVal)
            "
            label="country"
            v-model="instance.resources.user.country"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.state', +newVal)
            "
            label="state"
            v-model="instance.resources.user.state"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) =>
                setValue(index + '.resources.user.postal_code', +newVal)
            "
            label="postal_code"
            v-model="instance.resources.user.postal_code"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.phone', +newVal)
            "
            label="phone"
            v-model="instance.resources.user.phone"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.user.email', +newVal)
            "
            label="email"
            v-model="instance.resources.user.email"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.reg_username', +newVal)
            "
            label="reg_username"
            v-model="instance.resources.reg_username"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            @change="
              (newVal) => setValue(index + '.resources.reg_password', +newVal)
            "
            label="reg_password"
            v-model="instance.resources.reg_password"
          />
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="4">
          <v-switch
            @change="
              (newVal) => setValue(index + '.resources.auto_renew', +newVal)
            "
            v-model="instance.resources.auto_renew"
            label="auto_renew"
          />
        </v-col>
        <v-col cols="4">
          <v-switch
            @change="
              (newVal) => setValue(index + '.resources.who_is_privacy', +newVal)
            "
            v-model="instance.resources.who_is_privacy"
            label="who_is_privacy"
          />
        </v-col>
        <v-col cols="4">
          <v-switch
            @change="
              (newVal) => setValue(index + '.resources.lock_domain', +newVal)
            "
            v-model="instance.resources.lock_domain"
            label="lock_domain"
          />
        </v-col>
      </v-row>
      <v-row class="align-center">
        <v-col cols="4">
          <v-text-field label="domain" v-model="searchDomainString" />
        </v-col>
        <v-col cols="2">
          <v-btn
            @click="searchDomains"
            :disabled="!searchDomainString"
            small
            dark
            :loading="isDomainsLoading"
          >
            <v-icon dark> mdi-magnify </v-icon>
            Search
          </v-btn>
        </v-col>
        <v-col cols="2" class="d-flex justify-center align-center">
          <p v-if="!isPriceLoading">
            Price: {{ domainPrice || "select domain and period" }}
          </p>
          <v-progress-circular v-else indeterminate color="primary" />
        </v-col>
        <v-col cols="2" >
          <v-select v-model="selectedPeriodIndex" :items="domainPeriods" />
        </v-col>
      </v-row>
      <template v-if="avaliableDomains.length">
        <v-row class="pa-md-4">
          <v-list
            :disabled="isPriceLoading"
            width="80%"
            flat
            dark
            color="rgb(0,0,51)"
          >
            <v-subheader>domains</v-subheader>
            <v-list-item-group
              @change="(newVal) => changeDomain(newVal, index)"
              v-model="selectedDomainIndex"
              color="primary"
            >
              <v-list-item v-for="(item, i) in visibleDomains" :key="i">
                <v-list-item-content>
                  <v-list-item-title>
                    {{ item }}
                  </v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list-item-group>
          </v-list>
        </v-row>
        <v-row>
          <v-col cols="12">
            <div class="text-center">
              <v-pagination
                color="rgb(0,0,51)"
                v-model="selectedDomainPage"
                :length="domainPaginationLength"
              />
            </div>
          </v-col>
        </v-row>
      </template>
    </v-card>
    <v-row>
      <v-col class="d-flex justify-center">
        <v-btn
          :disabled="isOpenSRS"
          class="mx-2"
          small
          color="background"
          @click="addInstance"
        >
          <v-icon dark> mdi-plus-circle-outline </v-icon>
          add instance
        </v-btn>
      </v-col>
    </v-row>
  </div>
</template>

<script>
import api from "@/api";

export default {
  name: "ione-create-service-module",
  props: ["instances-group", "plans", "planRules"],
  data: () => ({
    periodRule: (v) => {
      if (!isNaN(parseFloat(v)) && v >= 0 && v <= 10) return true;
      return "Number has to be between 0 and 10";
    },
    prices: {},
    selectedPeriodIndex: 0,
    avaliableDomains: [],
    searchDomainString: "",
    selectedDomainPage: 1,
    selectedDomainIndex: null,
    isDomainsLoading: false,
    isPriceLoading: false,
    defaultItem: {
      title: "instance",
      resources: {
        user: {
          first_name: "",
          last_name: "",
          org_name: "",
          address1: "",
          address2: "",
          city: "",
          country: "",
          state: "",
          postal_code: "",
          phone: "",
          email: "",
        },
        reg_username: "",
        reg_password: "",
        domain: "",
        period: 1,
        auto_renew: true,
        who_is_privacy: false,
        lock_domain: true,
      },
    },
  }),
  methods: {
    addInstance() {
      const item = JSON.parse(JSON.stringify(this.defaultItem));
      const data = JSON.parse(this.instancesGroup);
      item.title += "#" + (data.body.instances.length + 1);

      data.body.instances.push(item);
      this.change(data);
    },
    remove(index) {
      const data = JSON.parse(this.instancesGroup);

      data.body.instances.splice(index, 1);
      this.change(data);
    },
    setValue(path, val) {
      if (val === undefined) return;

      const data = JSON.parse(this.instancesGroup);

      if (path.includes("domain")) {
        val = this.visibleDomains[val];
      }

      setToValue(data.body.instances, val, path);
      this.change(data);
    },
    change(data) {
      this.$emit("update:instances-group", JSON.stringify(data));
    },
    searchDomains() {
      this.isDomainsLoading = true;
      api.servicesProviders
        .action({
          uuid: this.spUuid,
          action: "get_domains",
          params: {
            searchString: this.searchDomainString,
            gTLD: true,
            p_ccTLD: false,
            m_ccTLD: false,
          },
        })
        .then((data) => {
          const freeDomains = [];
          data.meta.domains.forEach((d) => {
            if (d.Status === "available") {
              freeDomains.push(d.Domain);
            }
          });

          this.avaliableDomains = freeDomains;
        })
        .finally(() => {
          this.isDomainsLoading = false;
        });
    },
    changeDomain(newVal, index) {
      this.isPriceLoading = true;

      this.setValue(index + ".resources.domain", newVal);

      const domain = this.visibleDomains[newVal];

      api.servicesProviders
        .action({
          uuid: this.spUuid,
          action: "get_domain_price",
          params: {
            domain,
          },
        })
        .then((data) => {
          this.prices = data.meta.prices;
        })
        .finally(() => {
          this.isPriceLoading = false;
        });
    },
  },
  computed: {
    instances() {
      const data = JSON.parse(this.instancesGroup);
      return data.body.instances;
    },
    isOpenSRS() {
      const isOpenSrsSp =
        JSON.parse(this.instancesGroup).body.type === "opensrs";
      const isSpEmpty = JSON.parse(this.instancesGroup).sp;
      return isOpenSrsSp && !isSpEmpty;
    },
    spUuid() {
      return JSON.parse(this.instancesGroup).sp;
    },
    domainPaginationLength() {
      return Math.ceil(this.avaliableDomains.length / 10) - 1;
    },
    visibleDomains() {
      const start = this.selectedDomainPage * 10;
      return this.avaliableDomains.slice(start, start + 10);
    },
    domainPeriods() {
      if (Object.keys(this.prices).length === 0) {
        return null;
      }

      return Object.keys(this.prices);
    },
    domainPrice() {
      return this.prices[this.selectedPeriodIndex];
    },
  },
  created() {
    const data = JSON.parse(this.instancesGroup);
    if (!data.body.instances) {
      data.body.instances = [];
    }

    this.change(data);
  },
  watch: {
    selectedDomainPage() {
      this.selectedDomainIndex = null;
    },
  },
};

function setToValue(obj, value, path) {
  path = path.split(".");
  let i;
  for (i = 0; i < path.length - 1; i++) {
    if (path[i] === "__proto__" || path[i] === "constructor")
      throw new Error("Can't use that path because of: " + path[i]);
    obj = obj[path[i]];
  }

  obj[path[i]] = value;
}
</script>

<style></style>

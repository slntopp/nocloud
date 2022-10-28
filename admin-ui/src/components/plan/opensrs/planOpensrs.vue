<template>
  <v-form ref="opensrsForm" v-model="isValid">
    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Precision</v-subheader>
      </v-col>
      <v-col cols="9">
        <v-text-field
          type="number"
          @change="changeFee"
          :rules="generalRule"
          label="precision"
          v-model="currentFee.precision"
        />
      </v-col>
    </v-row>
    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Round</v-subheader>
      </v-col>
      <v-col cols="9">
        <v-text-field
          type="number"
          @change="changeFee"
          :rules="generalRule"
          label="round"
          v-model="currentFee.round"
        />
      </v-col>
    </v-row>
    <v-row align="center">
      <v-col cols="3">
        <v-subheader>Default</v-subheader>
      </v-col>
      <v-col cols="9">
        <v-text-field
          type="number"
          @change="changeFee"
          :rules="generalRule"
          label="default"
          v-model="currentFee.default"
        />
      </v-col>
    </v-row>
    <v-expansion-panels>
      <v-expansion-panel>
        <v-expansion-panel-header color="background-light">
          Ranges
        </v-expansion-panel-header>
        <v-expansion-panel-content color="background-light">
          <v-list v-if="currentFee.ranges.length > 0" color="background-light">
            <v-list-item-group>
              <template v-for="(item, index) in currentFee.ranges">
                <v-list-item :key="generateKey(index)">
                  <template v-slot:default>
                    <v-list-item-content>
                      <v-list-item-title>
                        From:{{ item.from }} To:{{ item.to }} Factor:{{
                          item.factor
                        }}
                      </v-list-item-title>
                    </v-list-item-content>

                    <v-list-item-action>
                      <v-list-item-action-text></v-list-item-action-text>

                      <v-icon
                        @click="deleteRange(index)"
                        color="grey lighten-1"
                      >
                        mdi-delete
                      </v-icon>
                    </v-list-item-action>
                  </template>
                </v-list-item>

                <v-divider
                  v-if="index < currentFee.ranges.length - 1"
                  :key="index"
                ></v-divider>
              </template>
            </v-list-item-group>
          </v-list>
          <p v-else class="text-center">Ranges clear</p>
          <v-form
            class="d-flex ma-5"
            ref="newRangeForm"
            v-model="isNewRangeValid"
          >
            <v-col cols="3">
              <v-text-field
                type="number"
                label="from"
                v-model="newRange.from"
                :rules="generalRule"
              />
            </v-col>
            <v-col cols="3">
              <v-text-field
                type="number"
                label="to"
                v-model="newRange.to"
                :rules="generalRule"
              />
            </v-col>
            <v-col cols="3">
              <v-text-field
                type="number"
                label="factor"
                v-model="newRange.factor"
                :rules="generalRule"
              />
            </v-col>
            <v-col cols="2" class="d-flex justify-center align-center">
              <v-btn color="background-light" @click="addRange">Add</v-btn>
            </v-col>
          </v-form>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>
  </v-form>
</template>

<script>
export default {
  props: {
    fee: { type: Object },
    isEdit: { type: Boolean, default: false },
  },
  name: "plan-opensrs",
  data() {
    return {
      currentFee: {
        ranges: [],
        default: 0,
        precision: 0,
        round: 0,
      },
      newRange: { from: 0, to: 0, factor: 0 },
      isValid: false,
      isNewRangeValid: false,
      generalRule: [(v) => !!v || v === 0 || "This field is required!"],
    };
  },
  created() {
    if (this.isEdit) {
      this.currentFee = this.fee;
    }
  },
  methods: {
    deleteRange(index) {
      this.currentFee.ranges = this.currentFee.ranges.filter(
        (item, i) => i !== index
      );
    },
    addRange() {
      if (!this.isNewRangeValid) {
        this.$refs.newRangeForm.validate();
        return;
      }
      this.currentFee.ranges.push(this.newRange);
      this.newRange = { from: 0, to: 0, factor: 0 };
    },
    changeFee() {
      if (!this.isValid) {
        this.$refs.opensrsForm.validate();
        return;
      }
      this.$emit("changeFee", this.currentFee);
    },
    generateKey(id) {
      return id + Math.random().toString(16).slice(2);
    },
  },
  watch: {
    "currentFee.ranges"() {
      this.changeFee();
    },
    isValid(newValue) {
      this.$emit("onValid", newValue);
    },
  },
};
</script>

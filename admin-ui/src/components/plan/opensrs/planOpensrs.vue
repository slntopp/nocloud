<template>
  <v-form ref="opensrsForm" v-model="isValid">
    <v-row>
      <v-col>
        <v-expansion-panels :value="0">
          <v-expansion-panel>
            <v-expansion-panel-header
              color="background"
              style="padding-left: 16px"
            >
              Ranged margin (%)
            </v-expansion-panel-header>
            <v-expansion-panel-content color="background">
              <v-list v-if="currentFee.ranges.length > 0" color="background">
                <v-list-item-group>
                  <template v-for="(item, index) in currentFee.ranges">
                    <v-list-item :key="generateKey(index)">
                      <template v-slot:default>
                        <v-list-item-content>
                          <v-list-item-title>
                            Price from: {{ item.from }}
                            {{ defaultCurrency?.code }}, Price to:
                            {{ item.to }} {{ defaultCurrency?.code }}, Margin:
                            {{ item.factor }}%
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
              <p v-else class="text-center">Margin clear</p>
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
                  <v-btn color="secondary" @click="addRange">Add</v-btn>
                </v-col>
              </v-form>
            </v-expansion-panel-content>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-col>
      <v-col>
        <v-row align="center">
          <v-col cols="3">
            <v-subheader>Fixed margin (%)</v-subheader>
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
        <v-row align="center">
          <v-col cols="3">
            <v-subheader>Round</v-subheader>
          </v-col>
          <v-col cols="9">
            <v-select
              item-text="key"
              item-value="value"
              @change="changeFee"
              v-model="currentFee.round"
              :items="availableRoundes"
            />
          </v-col>
        </v-row>
        <v-row align="center">
          <v-col cols="3">
            <v-subheader>Rounding precision</v-subheader>
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
      </v-col>
    </v-row>
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
        round: 2,
      },
      availableRoundes: [
        { key: "rounding down", value: 1 },
        { key: "rounding arithmetically", value: 2 },
        { key: "rounding up", value: 3 },
      ],
      newRange: { from: 0, to: 0, factor: 0 },
      isValid: false,
      isNewRangeValid: false,
      generalRule: [(v) => !!v || v === 0 || "This field is required!"],
    };
  },
  created() {
    if (this.isEdit) {
      this.currentFee = { ...this.currentFee, ...this.fee };

      if (typeof this.fee?.round !== "string") return;
      const roundes = [
        { key: "floor", value: 1 },
        { key: "round", value: 2 },
        { key: "ceil", value: 3 },
      ];
      const round = roundes.find(
        ({ key }) => key === this.fee?.round?.toLowerCase()
      );

      if (round) this.currentFee.round = round.value;
      else this.currentFee.round = 2;
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

      const { to, from, factor } = this.newRange;
      const range = this.currentFee.ranges.find((el) => {
        if (el.from <= from && from <= el.to) return true;
        if (el.from <= to && to <= el.to) return true;
        if (el.from >= from && to >= el.to) return true;
      });

      if (+to <= +from) {
        alert("The value to must be greater than the value from!");
        return;
      }
      if (+from < 0) {
        alert("The value to must be greater or equal than 0!");
        return;
      }
      if (range) {
        alert("The Rule for proposed price range has been existed already!");
        return;
      }

      this.currentFee.ranges.push({ to: +to, from: +from, factor: +factor });

      this.newRange = { from: 0, to: 0, factor: 0 };
    },
    changeFee() {
      if (!this.isValid) {
        this.$refs.opensrsForm.validate();
        return;
      }

      const { precision, round, ranges } = this.currentFee;

      this.$emit("changeFee", {
        precision: +precision,
        round,
        default: +this.currentFee.default,
        ranges,
      });
    },
    generateKey(id) {
      return id + Math.random().toString(16).slice(2);
    },
  },
  computed: {
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
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

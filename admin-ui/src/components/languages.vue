<template>
  <v-select
    class="mr-3 ml-3 mt-5 select-langs"
    :items="availableLangs"
    v-model="currentLang"
    @change="changeLang"
  >
  </v-select>
</template>

<script>
import config from "@/config.js";

export default {
  name: "app-languages",
  data() {
    return { currentLang: "" };
  },
  created() {
    const lang = localStorage.getItem("lang");
    if (lang != undefined) this.$i18n.locale = lang;
    this.currentLang = this.$i18n.locale.toUpperCase();
  },
  computed: {
    availableLangs() {
      return config.languages.map((lang) => lang.toUpperCase());
    },
  },
  methods: {
    changeLang(lang) {
      this.$i18n.locale = lang.toLowerCase();
      localStorage.setItem("lang", this.$i18n.locale);
    },
  },
};
</script>

<style>
.select-langs {
  max-width: 70px;
}
</style>

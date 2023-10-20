<script>
import { mapGetters } from "vuex";

export default {
  name: "themes-selector",
  computed: {
    ...mapGetters("app", ["theme"]),
  },
  created() {
    let savedTheme = localStorage.getItem("nocloud-theme");
    if (!savedTheme && window.matchMedia) {
      if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
        savedTheme = "dark";
      } else {
        savedTheme = "light";
      }
    }

    this.$store.commit("app/setTheme", savedTheme || undefined);
    this.$vuetify.theme.dark = this.theme === "dark";
  },
  watch: {
    theme() {
      this.$vuetify.theme.dark = this.theme === "dark";
    },
  },
};
</script>

<script setup>
import { computed, watch, ref, onMounted } from "vue";
import { useStore } from "@/store";

const store = useStore();
const systemTheme = ref("");

onMounted(() => {
  if (window.matchMedia) {
    if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
      systemTheme.value = "dark";
    } else {
      systemTheme.value = "light";
    }
  }
});

const themes = computed(() => {
  const themes = [
    { title: "dark", icon: "mdi-moon-waning-crescent" },
    { title: "light", icon: "mdi-brightness-5" },
  ];
  if (systemTheme.value) {
    themes.push({ title: "system", icon: "mdi-cog" });
  }
  return themes;
});

const theme = computed(() => store.getters["app/theme"]);

const currentTheme = computed(() => {
  return themes.value.find((t) => t.title === theme.value);
});

const changeTheme = (theme) => {
  if (theme === "system") {
    theme = systemTheme.value;
  }
  store.commit("app/setTheme", theme);
};

watch(theme, (newValue) => {
  localStorage.setItem("nocloud-theme", newValue);
});
</script>

<template>
  <v-menu offset-y transition="slide-y-transition">
    <template v-slot:activator="{ on, attrs }">
      <v-btn class="mx-2" icon v-bind="attrs" v-on="on">
        <v-icon>{{ currentTheme.icon }} </v-icon>
      </v-btn>
    </template>
    <v-list flat>
      <v-list-item-group>
        <v-list-item
          @change="changeTheme(theme.title)"
          :key="theme.title"
          v-for="theme in themes"
        >
          <div class="d-flex align-center justify-space-between">
            <v-icon>{{ theme.icon }}</v-icon>
            <span class="mx-3">{{
              theme.title.slice(0, 1).toUpperCase() + theme.title.slice(1)
            }}</span>
          </div>
        </v-list-item>
      </v-list-item-group>
    </v-list>
  </v-menu>
</template>
<style scoped></style>

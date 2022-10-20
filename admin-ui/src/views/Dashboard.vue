<template>
  <div
    class="widgets align-start gg-15px pa-4"
    :class="{
      'd-grid': width < 1375,
      'd-flex': width >= 1375,
    }"
  >
    <component
      v-for="widget of widgets"
      :key="widget"
      :is="widget"
      :class="{
        'grid-row': widget === 'ServicesWidget' && width >= 1010,
      }"
    />
  </div>
</template>

<script>
import HealthWidget from "@/components/widgets/health";
import ServicesWidget from "@/components/widgets/services";
import RoutinesWidget from "@/components/widgets/routines";

export default {
  name: "dashboard-view",
  components: {
    HealthWidget,
    ServicesWidget,
    RoutinesWidget,
  },
  data: () => ({
    widgets: ["HealthWidget", "ServicesWidget", "RoutinesWidget"],
  }),
  computed: {
    width() {
      return document.documentElement.clientWidth;
    },
  },
  mounted(){
    this.$store.commit("reloadBtn/setCallback", {
      event:()=>this.$router.go(),
    });
  }
};
</script>

<style scoped lang="scss">
.d-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;

  @media (max-width: 1010px) {
    grid-template-columns: 1fr;
    justify-items: center;
  }
}
.grid-row {
  grid-row: 1 / 3;
  justify-self: end;
}
</style>

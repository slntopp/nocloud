<template>
  <component :is="icon" />
</template>

<script>
export default {
  name: "ant-icon",
  props: ["name"],
  computed: {
    icon() {
      let name = this.name.charAt(0).toUpperCase() + this.name.slice(1);

      return () => import(`@ant-design/icons-vue/${name}Outlined`);
    },
  },
};
</script>

<style scoped></style>

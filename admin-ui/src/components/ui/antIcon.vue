<template>
  <antd-icon style="font-size: 24px" :type="icon" />
</template>

<script>
import antdIcon from "@ant-design/icons-vue";

export default {
  name: "ant-icon",
  props: ["name"],
  components: { antdIcon },
  data: () => ({ icon: {} }),
  async mounted() {
    const iconsRes = await import("@ant-design/icons");
    let displayName = this.name;

    if (displayName.endsWith("Outlined")) {
      displayName = displayName.replace("Outlined", "Outline");
    } else if (displayName.endsWith("Filled")) {
      displayName = displayName.replace("Filled", "Fill");
    }
    const [name, icon] = Object.entries(iconsRes)?.find(
      ([name]) => name === displayName
    );
    this.icon = { name, ...icon };
  },
};
</script>

<style scoped></style>

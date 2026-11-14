<script>
import "./jodit.min.css";

export default {
  name: "editor-container",
  props: [
    'modelValue'
  ]
};

</script>

<template>
  <div class="jodit-wysiwyg" v-html="modelValue"></div>
</template>
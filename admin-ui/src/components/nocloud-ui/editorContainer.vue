<script>
import "./jodit.min.css";

export default {
  name: "editor-container",
  props: [
    'value'
  ]
};

</script>

<template>
  <div class="jodit-wysiwyg" v-html="value"></div>
</template>
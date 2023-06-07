<template>
  <vue-editor
    :editor-toolbar="toolbarOptions"
    class="html-editor"
    :value="props.value"
    @input="emit('input', $event)"
  />
</template>

<script setup>
import { VueEditor } from "vue2-editor";


const props = defineProps(["value"]);
const emit = defineEmits(["input"]);

const toolbarOptions = [
  [{ header: [false, 1, 2, 3, 4, 5, 6] }],
  [{ size: [] }],
  ["bold", "italic", "underline", "strike"], // toggled buttons
  [
    { align: "" },
    { align: "center" },
    { align: "right" },
    { align: "justify" },
  ],
  ["blockquote", "code-block"],
  [{ list: "ordered" }, { list: "bullet" }, { list: "check" }],
  [{ indent: "-1" }, { indent: "+1" }],
  [{ color: [] }, { background: [] }],
  ["link", "image", "video"],
  ["clean"],
];
</script>

<style  lang="scss">
.html-editor {
  span.ql-picker-label {
    color: white;
  }
}

.quillWrapper .ql-snow .ql-stroke {
  stroke: rgb(255 255 255 / 95%) !important;
}
.ql-snow .ql-fill {
  fill: white;
}

.quillWrapper .ql-editor {
  color: white;
}
.quillWrapper .ql-editor ul[data-checked="false"] > li::before {
  color: white !important;
}
.quillWrapper .ql-editor ul[data-checked="true"] > li::before {
  color: white !important;
}

.ql-active {
  color: #e06ffe !important;
  fill: #e06ffe !important;
  stroke: #e06ffe !important;
}
</style>

<template>
  <div @click="emits('click')">
    <v-chip
      v-for="tag in tags.slice(0, 2)"
      :key="tag.title"
      class="ml-1 pa-1"
      small
      outlined
      color="primary"
      >{{ tag.title }}
      <v-btn icon x-small @click.stop="deleteParam(tag)">
        <v-icon small>mdi-close</v-icon>
      </v-btn>
    </v-chip>
    <v-chip v-if="tags.length > 2" class="ml-1 pa-1" small outlined color="primary"
      >and more {{ tags.length - 2 }}</v-chip
    >
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useStore } from "@/store";

const emits = defineEmits(["click"]);

const store = useStore();

const filter = computed(() => store.getters["appSearch/filter"]);
const fields = computed(() => store.getters["appSearch/fields"]);

const tags = computed(() => {
  const tags = [];
  Object.keys(filter.value).forEach((key) => {
    const field = findField(key) || {};
    const filterValue = filter.value[key];
    switch (field.type) {
      case "input": {
        if (!filterValue) {
          return;
        }

        return tags.push({
          title: `${field.title}:${filterValue}`,
          value: filterValue,
          key,
        });
      }
      case "select": {
        const val = [];
        if (field.single) {
          if (!filterValue) {
            return;
          }
          val.push(filterValue);
        } else {
          val.push(...filterValue);
        }

        return tags.push(
          ...val.map((v) => {
            let title = v;
            if (field.item) {
              title = field.items.find((i) => i[field.item.value] === v)?.title;
            }

            return {
              title: `${field.title}:${title}`,
              value: v,
              key,
            };
          })
        );
      }
      case "number-range": {
        if (Object.keys(filterValue).length === 0) {
          return;
        }

        let valueTitle = "";
        if (filterValue.from) {
          valueTitle += "from " + filterValue.from;
        }
        if (filterValue.to) {
          valueTitle += " to " + filterValue.to;
        }
        return tags.push({
          title: `${field.title}:${valueTitle.trim()}`,
          value: filterValue,
          key,
        });
      }
      case "logic-select": {
        if (filterValue !== false && filterValue !== true) {
          return;
        }

        return tags.push({
          title: `${field.title}:${filterValue ? "Enabled" : "Disabled"}`,
          value: filterValue,
          key,
        });
      }
      case "date": {
        if (!filterValue || filterValue.length === 0) {
          return;
        }

        let valueTitle = "";
        if (filterValue[0]) {
          valueTitle += "from " + filterValue[0];
        }
        if (filterValue[1]) {
          valueTitle += " to " + filterValue[1];
        }
        return tags.push({
          title: `${field.title}:${valueTitle.trim()}`,
          value: filterValue,
          key,
        });
      }
      default: {
        return;
      }
    }
  });

  return tags.map((t) => ({
    ...t,
    title: t.title.length > 20 ? t.title.slice(0, 20) + "..." : t.title,
  }));
});

const findField = (key) => {
  return fields.value.find((f) => f.key === key);
};
const deleteParam = ({ key, value }) => {
  const field = findField(key);

  let newValue;
  switch (field.type) {
    case "input":
    case "date":
    case "logic-select": {
      newValue = undefined;
      break;
    }
    case "number-range": {
      newValue = {};
      break;
    }
    case "select": {
      if (field.single) {
        newValue = undefined;
        break;
      }
      newValue = filter.value[key].filter((f) => f !== value);
      break;
    }
  }

  store.commit("appSearch/setFilterValue", { key, value: newValue });
};
</script>

<style scoped></style>

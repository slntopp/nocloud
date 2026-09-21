<template>
  <div class="notifications">
    <span v-if="!count" class="notifications__empty">—</span>

    <template v-else>
      <v-menu
        v-for="tag of tags"
        :key="tag.kind"
        open-on-hover
        offset-y
        max-width="420"
        :close-delay="150"
        @input="(open) => open && load()"
      >
        <template v-slot:activator="{ on }">
          <v-chip
            v-on="on"
            x-small
            :color="tag.color"
            :title="tag.hint"
            @click.stop="openJournal"
          >
            <v-icon x-small left>{{ tag.icon }}</v-icon>
            {{ tag.count }}
          </v-chip>
        </template>

        <v-card>
          <v-card-text class="py-2 px-3">
            <div class="notifications__head">
              {{ tag.hint }}
            </div>

            <v-progress-linear v-if="loading" indeterminate height="2" class="my-2" />

            <div v-else-if="!rows.length" class="notifications__none">
              nothing in the journal
            </div>

            <div v-for="(row, i) of rows" :key="i" class="notifications__row">
              <span class="notifications__when">{{ when(row.ts) }}</span>
              <v-icon x-small>{{ row.source === "nocloud" ? mail : ticket }}</v-icon>
              <span class="notifications__what">{{ title(row) }}</span>
            </div>

            <div class="notifications__more">click to open the journal</div>
          </v-card-text>
        </v-card>
      </v-menu>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, toRefs } from "vue";
import { useStore } from "@/store";
import { useRouter } from "vue-router/composables";

// How much has gone out to this customer, or about this service, over the
// last thirty days: letters nocloud sent and tickets the campaigns plugin
// opened, counted apart. A letter is the billing writing by itself, a ticket
// is somebody writing on purpose, and one number for both would answer a
// question nobody asked.
const props = defineProps({
  uuid: { type: String, required: true },
  kind: { type: String, default: "accounts" },
  // Whose journal to open on a click. For a customer it is the customer; for
  // a service it is its owner, because the journal filters by customer.
  account: { type: String, default: "" },
});
const { uuid, kind } = toRefs(props);

const store = useStore();
const router = useRouter();

const mail = "mdi-email-outline";
const ticket = "mdi-ticket-outline";

const loading = ref(false);
const asked = ref(false);

const account = computed(() => props.account || uuid.value);

const count = computed(() => {
  const counted = store.getters["campaigns/count"](uuid.value);
  // Nothing at all is shown as a dash rather than as two zeros: a table is
  // read by running an eye down it, and zeros in every row hide the rows that
  // are not zero.
  return counted && (counted.letters || counted.tickets) ? counted : null;
});

const rows = computed(() => store.getters["campaigns/recent"](uuid.value) || []);

// Colour by how much: how often we are bothering this customer, not how
// important it is. Two numbers, one scale.
function color(n) {
  if (n >= 6) return "error";
  if (n >= 3) return "warning";
  return "info";
}

const tags = computed(() => {
  const out = [];
  if (count.value?.letters) {
    out.push({
      kind: "letters",
      icon: mail,
      count: count.value.letters,
      color: color(count.value.letters),
      hint: `${count.value.letters} letters from nocloud in 30 days`,
    });
  }
  if (count.value?.tickets) {
    out.push({
      kind: "tickets",
      icon: ticket,
      count: count.value.tickets,
      color: color(count.value.tickets),
      hint: `${count.value.tickets} campaign tickets in 30 days`,
    });
  }
  return out;
});

// The last ten, fetched the first time somebody hovers and kept afterwards:
// hovering down a table would otherwise be a request per row.
async function load() {
  if (asked.value) return;
  asked.value = true;
  loading.value = true;
  try {
    await store.dispatch("campaigns/recent", { uuid: uuid.value, kind: kind.value });
  } catch {
    // A column is not worth an error on somebody's screen: the plugin may be
    // down, or this admin may not be allowed to read its journal. The tag
    // then simply has nothing to show.
  } finally {
    loading.value = false;
  }
}

// A letter carries the event's key, a ticket the template's name.
function title(row) {
  return row.templateTitle || row.event || row.template || "—";
}

function when(ts) {
  return ts ? new Date(Number(ts) * 1000).toLocaleDateString() : "";
}

// The journal of the plugin, already filtered by this customer. The plugin
// takes a redirect the same way the chats one does, so this is the route that
// already exists rather than a new way in.
function openJournal() {
  const plugin = (store.getters["plugins/all"] || []).find((p) =>
    (p.url || "").includes("campaigns.ui")
  );
  if (!plugin) return;

  router.push({
    name: "Plugin",
    params: {
      title: plugin.title,
      url: plugin.url,
      params: { redirect: `dashboard/sends?account=${account.value}` },
    },
    query: { url: plugin.url },
  });
}

onMounted(() => {
  // Same here: a table that cannot reach the plugin is a table without this
  // column, not a table with an error in it.
  store.dispatch("campaigns/count", { uuid: uuid.value, kind: kind.value }).catch(() => {});
});
</script>

<style scoped>
.notifications {
  display: flex;
  align-items: center;
  gap: 4px;
}
.notifications__empty {
  opacity: 0.4;
}
.notifications__head {
  font-size: 12px;
  opacity: 0.7;
  margin-bottom: 4px;
}
.notifications__row {
  display: grid;
  grid-template-columns: 72px 16px 1fr;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  padding: 1px 0;
}
.notifications__when {
  opacity: 0.6;
}
.notifications__what {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.notifications__none,
.notifications__more {
  font-size: 12px;
  opacity: 0.5;
  padding-top: 4px;
}
</style>

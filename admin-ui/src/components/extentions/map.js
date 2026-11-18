import { defineAsyncComponent } from "vue";
export default {
	"nocloud-tunnelmesh": {
		title: "NoCloud Tunnelmesh",
		component: defineAsyncComponent(() => import('./components/nocloud-tunnelmesh/onCreate.vue')),
		pageComponent: defineAsyncComponent(() => import('./components/nocloud-tunnelmesh/onPage.vue')),
	}
}
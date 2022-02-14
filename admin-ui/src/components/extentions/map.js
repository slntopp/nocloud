export default {
	"nocloud-tunnelmesh": {
		title: "NoCloud Tunnelmesh",
		component: () => import('./components/nocloud-tunnelmesh/onCreate.vue'),
		pageComponent: () => import('./components/nocloud-tunnelmesh/onPage.vue'),
	}
}
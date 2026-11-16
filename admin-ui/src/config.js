export default {
	languages: ['en', 'ru'],
	colors: globalThis.colors,
	dark: globalThis.dark,
	logoSrc: globalThis.logoSrc ?? "",
	dye: {
		aside: {
			background: globalThis?.dye?.aside?.background,
			links: globalThis?.dye?.aside?.links,
			whiteText: globalThis?.dye?.aside?.whiteText,
		}
	},
	navTitles: globalThis.navTitles ?? {},
	// ponytail: support.by's chat as the default, set aiChatUrl in public/js/config.js for another installation
	aiChatUrl: globalThis.aiChatUrl ?? "https://ai.support.by",
}
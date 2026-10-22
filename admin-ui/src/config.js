export default {
	languages: ['en', 'ru'],
	colors: global.colors,
	dark: global.dark,
	logoSrc: global.logoSrc ?? "",
	dye: {
		aside: {
			background: global?.dye?.aside?.background,
			links: global?.dye?.aside?.links,
			whiteText: global?.dye?.aside?.whiteText,
		}
	},
	navTitles: global.navTitles ?? {},
	// ponytail: support.by's chat as the default, set aiChatUrl in public/js/config.js for another installation
	aiChatUrl: global.aiChatUrl ?? "https://ai.support.by",
}
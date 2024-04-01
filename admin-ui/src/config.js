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
	timeZone: global.timeZone
}
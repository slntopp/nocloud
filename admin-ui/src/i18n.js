import config from "./config";
import { createI18n } from "vue-i18n";

function loadLocaleMessages() {
  const locales = import.meta.glob("./locales/*.json", { eager: true });
  const messages = {};
  Object.entries(locales).forEach(([key, module]) => {
    const matched = key.match(/([A-Za-z0-9-_]+)\.json$/i);
    if (matched && matched.length > 1) {
      messages[matched[1]] = module.default;
    }
  });
  return messages;
}

const AppLangs = config.languages;
const localLang = localStorage.getItem("lang");

const lang = AppLangs.find((appLang) => appLang === localLang) || "en";

export default createI18n({
  // legacy mode keeps this.$t / this.$i18n.locale working in options API components
  legacy: true,
  globalInjection: true,
  locale: lang,
  silentTranslationWarn: true,
  silentFallbackWarn: true,
  missingWarn: false,
  fallbackWarn: false,
  fallbackLocale: import.meta.env.VITE_I18N_FALLBACK_LOCALE || "en",
  messages: loadLocaleMessages(),
  pluralRules: {
    ru: function (choice, choicesLength) {
      // this === VueI18n instance, so the locale property also exists here
      if (choice === 0) {
        return 0;
      }
      const teen = choice > 10 && choice < 20;
      const endsWithOne = choice % 10 === 1;
      if (choicesLength < 4) {
        return !teen && endsWithOne ? 1 : 2;
      }
      if (!teen && endsWithOne) {
        return 1;
      }
      if (!teen && choice % 10 >= 2 && choice % 10 <= 4) {
        return 2;
      }
      return choicesLength < 4 ? 2 : 3;
    },
  },
});

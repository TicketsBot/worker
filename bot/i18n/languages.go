package i18n

import (
	translations "github.com/TicketsBot/database/translations"
)

const (
	English       translations.Language = "en"
	French        translations.Language = "fr"
	Spanish       translations.Language = "es"
	German        translations.Language = "de"
	Dutch         translations.Language = "nl"
	Polish        translations.Language = "pl"
	Norwegian     translations.Language = "no"
	Turkish       translations.Language = "tr"
	Swedish       translations.Language = "sv"
	Arabic        translations.Language = "ar"
	Hungarian     translations.Language = "hu"
	Russian       translations.Language = "ru"
	PortugueseBR  translations.Language = "br"
	Chinese       translations.Language = "cn"
	ChineseTaiwan translations.Language = "tw"
)

var Flags = map[translations.Language]string{
	English:       "🇬🇧",
	French:        "🇫🇷",
	Spanish:       "🇪🇸",
	German:        "🇩🇪",
	Dutch:         "🇳🇱",
	Polish:        "🇵🇱",
	Norwegian:     "🇳🇴",
	Turkish:       "🇹🇷",
	Swedish:       "🇸🇪",
	Arabic:        "🇸🇦",
	Hungarian:     "🇭🇺",
	Russian:       "🇷🇺",
	PortugueseBR:  "🇧🇷",
	Chinese:       "🇨🇳",
	ChineseTaiwan: "🇹🇼",
}

// https://discord.com/developers/docs/dispatch/field-values
var Locales = map[string]translations.Language{
	"en-US": English,
	"en-GB": English,
	"zh-CN": Chinese,
	"zh-TW": ChineseTaiwan,
	"cs":    English, // Czech
	"da":    English, // Danish
	"nl":    Dutch,
	"fr":    French,
	"de":    German,
	"hu":    Hungarian,
	"it":    English, // Italian
	"ja":    English, // Japanese
	"ko":    English, // Korean
	"no":    Norwegian,
	"pl":    Polish,
	"pt-BR": PortugueseBR,
	"ru":    Russian,
	"es-ES": Spanish,
	"sv-SE": Swedish,
	"tr":    Turkish,
	"bg":    English, // Bulgarian
	"uk":    English, // Ukrainian
	"fi":    English, // Finnish
	"hr":    English, // Croatian
	"ro":    English, // Romanian
	"lt":    English, // Lithuanian
}

package i18n

type Language string

func (l Language) String() string {
	return string(l)
}

const (
	Arabic        Language = "ar"
	Bulgarian     Language = "bg"
	Czech         Language = "cz"
	Danish        Language = "dk"
	German        Language = "de"
	Greek         Language = "el"
	English       Language = "en"
	Spanish       Language = "es"
	Finnish       Language = "fi"
	French        Language = "fr"
	Croatian      Language = "hr"
	Hungarian     Language = "hu"
	Italian       Language = "it"
	Japanese      Language = "jp"
	Korean        Language = "kr"
	Lithuanian    Language = "lt"
	Dutch         Language = "nl"
	Norwegian     Language = "no"
	Polish        Language = "pl"
	PortugueseBR  Language = "br"
	Portuguese    Language = "pt"
	Romanian      Language = "ro"
	Russian       Language = "ru"
	Slovak        Language = "sk"
	Swedish       Language = "sv"
	Thai          Language = "th"
	Turkish       Language = "tr"
	Ukrainian     Language = "ua"
	Vietnamese    Language = "vn"
	Chinese       Language = "cn"
	ChineseTaiwan Language = "tw"
)

var Flags = map[Language]string{
	Arabic:        "🇸🇦",
	Bulgarian:     "🇧🇬",
	Czech:         "🇨🇿",
	Danish:        "🇩🇰",
	German:        "🇩🇪",
	Greek:         "🇬🇷",
	English:       "🇬🇧",
	Spanish:       "🇪🇸",
	Finnish:       "🇫🇮",
	French:        "🇫🇷",
	Croatian:      "🇭🇷",
	Hungarian:     "🇭🇺",
	Italian:       "🇮🇹",
	Japanese:      "🇯🇵",
	Korean:        "🇰🇷",
	Lithuanian:    "🇱🇹",
	Dutch:         "🇳🇱",
	Norwegian:     "🇳🇴",
	Polish:        "🇵🇱",
	PortugueseBR:  "🇧🇷",
	Portuguese:    "🇵🇹",
	Romanian:      "🇷🇴",
	Russian:       "🇷🇺",
	Slovak:        "🇸🇰",
	Swedish:       "🇸🇪",
	Thai:          "🇹🇭",
	Turkish:       "🇹🇷",
	Ukrainian:     "🇺🇦",
	Vietnamese:    "🇻🇳",
	Chinese:       "🇨🇳",
	ChineseTaiwan: "🇹🇼",
}

// https://discord.com/developers/docs/reference#locales
var DiscordLocales = map[string]Language{
	"en-US": English,
	"en-GB": English,
	"bg":    Bulgarian,
	"zh-CN": Chinese,
	"zh-TW": ChineseTaiwan,
	"hr":    Croatian,
	"cs":    Czech,
	"da":    Danish,
	"nl":    Dutch,
	"el":    Greek,
	"fi":    Finnish,
	"fr":    French,
	"de":    German,
	"hu":    Hungarian,
	"it":    Italian,
	"ja":    Japanese,
	"ko":    Korean,
	"lt":    Lithuanian,
	"no":    Norwegian,
	"pl":    Polish,
	"pt-BR": PortugueseBR,
	"ro":    Romanian,
	"ru":    Russian,
	"es-ES": Spanish,
	"sv-SE": Swedish,
	"th":    Thai,
	"tr":    Turkish,
	"uk":    Ukrainian,
}

// Used by CrowdIn
var FullLocales = map[string]Language{
	"ar-SA": Arabic,
	"bg-BG": Bulgarian,
	"cs-CZ": Czech,
	"da-DK": Danish,
	"de-DE": German,
	"el-GR": Greek,
	"en-GB": English,
	"es-ES": Spanish,
	"fi-FI": Finnish,
	"fr-FR": French,
	"hr-HR": Croatian,
	"hu-HU": Hungarian,
	"it-IT": Italian,
	"ja-JP": Japanese,
	"ko-KR": Korean,
	"lt-LT": Lithuanian,
	"nl-NL": Dutch,
	"no-NO": Norwegian,
	"pl-PL": Polish,
	"pt-BR": PortugueseBR,
	"pt-PT": Portuguese,
	"ro-RO": Romanian,
	"ru-RU": Russian,
	"sk-SK": Slovak,
	"sv-SE": Swedish,
	"th-TH": Thai,
	"tr-TR": Turkish,
	"uk-UA": Ukrainian,
	"vi-VN": Vietnamese,
	"zh-CN": Chinese,
	"zh-TW": ChineseTaiwan,
}

var LanguagesAlphabetical = []Language{
	Arabic,
	Bulgarian,
	Czech,
	Danish,
	German,
	Greek,
	English,
	Spanish,
	Finnish,
	French,
	Croatian,
	Hungarian,
	Italian,
	Japanese,
	Korean,
	Lithuanian,
	Dutch,
	Norwegian,
	Polish,
	PortugueseBR,
	Portuguese,
	Romanian,
	Russian,
	Slovak,
	Swedish,
	Thai,
	Turkish,
	Ukrainian,
	Vietnamese,
	Chinese,
	ChineseTaiwan,
}

var FullNames = map[Language]string{
	Arabic:        "Arabic",
	Bulgarian:     "Bulgarian",
	Czech:         "Czech",
	Danish:        "Danish",
	German:        "German",
	Greek:         "Greek",
	English:       "English",
	Spanish:       "Spanish",
	Finnish:       "Finnish",
	French:        "French",
	Croatian:      "Croatian",
	Hungarian:     "Hungarian",
	Italian:       "Italian",
	Japanese:      "Japanese",
	Korean:        "Korean",
	Lithuanian:    "Lithuanian",
	Dutch:         "Dutch",
	Norwegian:     "Norwegian",
	Polish:        "Polish",
	PortugueseBR:  "Portuguese (Brazilian)",
	Portuguese:    "Portuguese",
	Romanian:      "Romanian",
	Russian:       "Russian",
	Slovak:        "Slovak",
	Swedish:       "Swedish",
	Thai:          "Thai",
	Turkish:       "Turkish",
	Ukrainian:     "Ukrainian",
	Vietnamese:    "Vietnamese",
	Chinese:       "Chinese",
	ChineseTaiwan: "Chinese (Taiwan)",
}

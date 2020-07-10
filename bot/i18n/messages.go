package i18n

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/dbclient"
	"strings"
)

var messages map[translations.Language]map[translations.MessageId]string

func LoadMessages(db *database.Database) (err error) {
	messages, err = db.Translations.GetAll()
	return
}

func GetMessage(language translations.Language, id translations.MessageId) string {
	if messages[language] == nil {
		if language == translations.English {
			return "error: lang en is missing"
		}

		language = translations.English // default to english
		return GetMessage(language, id)
	}

	value, ok := messages[language][id]
	if !ok {
		if language == translations.English {
			return fmt.Sprintf("error: translation for %d is missing", id)
		}

		language = translations.English // default to english
		return GetMessage(language, id)
	}

	return strings.Replace(value, "\\n", "\n", -1)
}

func GetMessageFromGuild(guildId uint64, id translations.MessageId) string {
	language, err := dbclient.Client.ActiveLanguage.Get(guildId)
	if err != nil {
		sentry.Error(err)
		language = translations.English
	}

	if language == "" {
		language = translations.English
	}

	return GetMessage(language, id)
}

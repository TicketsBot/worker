package i18n

type MessageId string

// note: %s = a placeholder
var (
	MessageNoPermission MessageId = "generic.no_permission"
	MessageOwnerOnly    MessageId = "generic.owner_only"

	MessageAbout MessageId = "commands.about"

	MessagePremium MessageId = "commands.premium"

	MessageVote              MessageId = "commands.vote"
	MessageInvalidArgument   MessageId = "generic.invalid_argument"
	MessageCloseNoPermission MessageId = "close.no_permission"

	MessageTagCreateInvalidArguments MessageId = "commands.tags.create.invalid_arguments"
	MessageTagCreateTooLong          MessageId = "commands.tags.create.too_long"
	MessageTagCreateAlreadyExists    MessageId = "commands.tags.create.already_exists"

	MessageTagDeleteInvalidArguments MessageId = "commands.tags.delete.invalid_arguments"
	MessageTagDeleteDoesNotExist     MessageId = "commands.tags.delete.not_exist"

	MessageTagList MessageId = "commands.tags.list"

	MessageTagInvalidArguments MessageId = "commands.tags.get.invalid_arguments"
	MessageTagInvalidTag       MessageId = "commands.tags.get.invalid_tag"

	MessageTicketOpened MessageId = "open.success"

	MessageAddAdminNoMembers   MessageId = "commands.addadmin.no_members"
	MessageAddSupportNoMembers MessageId = "commands.addsupport.no_members"

	MessageAddNoMembers        MessageId = "commands.add.no_members"
	MessageAddNoChannel        MessageId = "commands.add.no_channel"
	MessageAddChannelNotTicket MessageId = "commands.add.not_ticket"
	MessageAddNoPermission     MessageId = "commands.add.no_permission"

	MessageBlacklisted        MessageId = "commands.blacklist.success"
	MessageBlacklistNoMembers MessageId = "commands.blacklist.no_members"
	MessageBlacklistSelf      MessageId = "commands.blacklist.self"
	MessageBlacklistStaff     MessageId = "command.blacklist.staff"

	MessageClaimed           MessageId = "commands.claim.success"
	MessageClaimNoPermission MessageId = "commands.claim.no_permission"

	MessageHelpDMFailed MessageId = "commands.help.failed"

	MessagePanel MessageId = "commands.panel"

	MessageAlreadyPremium    MessageId = "commands.premium.already_premium"
	MessageInvalidPremiumKey MessageId = "commands.premium.invalid_key"

	MessageRemoveAdminNoMembers MessageId = "commands.removeadmin.no_members"
	MessageOwnerMustBeAdmin     MessageId = "commands.removeadmin.owner"
	MessageRemoveStaffSelf      MessageId = "commands.removeadmin.self"

	MessageRemoveSupportNoMembers MessageId = "commands.removesupport.no_members"

	MessageRemoveNoMembers         MessageId = "commands.remove.no_members"
	MessageRemoveNoPermission      MessageId = "commands.remove.no_permission"
	MessageRemoveCannotRemoveStaff MessageId = "commands.remove.staff"

	MessageRenamed           MessageId = "commands.rename.success"
	MessageRenameMissingName MessageId = "commands.rename.missing_name"
	MessageRenameTooLong     MessageId = "commands.rename.too_long"

	MessageNotClaimed            MessageId = "commands.unclaim.not_claimed"
	MessageOnlyClaimerCanUnclaim MessageId = "commands.unclaim.not_claimer"
	MessageUnclaimed             MessageId = "commands.unclaim.success"

	MessageDisabledLogChannel MessageId = "setup.disabling_archiving"
	MessageInvalidCategory    MessageId = "setup.invalid_category"
	MessageCreatedCategory    MessageId = "setup.category_created"
	MessageInvalidPrefix      MessageId = "setup.prefix.invalid"
	MessageInvalidTicketLimit MessageId = "setup.ticket_limit.invalid"

	MessageNotATicketChannel MessageId = "generic.not_ticket"
	MessageInvalidUser       MessageId = "generic.invalid_user"

	MessageTicketLimitReached MessageId = "commands.open.ticket_limit"
	MessageTooManyTickets     MessageId = "commands.open.too_many_tickets"

	MessageAutoCloseConfigure MessageId = "commands.autoclose.configure"
	MessageAutoCloseExclude   MessageId = "commands.autoclose.exclude.success"

	SetupArchiveChannel  MessageId = "setup.info.archive_channel"
	SetupChannelCategory MessageId = "setup.info.category"
	SetupPrefix          MessageId = "setup.info.prefix"
	SetupTicketLimit     MessageId = "setup.info.ticket_limit"
	SetupWelcomeMessage  MessageId = "setup.info.welcome_message"

	HelpAdmin              MessageId = "help.admin"
	HelpAdminCheckPerms    MessageId = "help.admin.check_permissions"
	HelpAdminDebug         MessageId = "help.admin.debug"
	HelpAdminForceClose    MessageId = "help.admin.force_close"
	HelpAdminGC            MessageId = "help.admin.gc"
	HelpAdminGenPremium    MessageId = "help.admin.generate_premium"
	HelpAdminGetOwner      MessageId = "help.admin.get_owner"
	HelpAdminPing          MessageId = "help.admin.ping"
	HelpAdminSeed          MessageId = "help.admin.seed"
	HelpAdminUpdateSchema  MessageId = "help.admin.update_schema"
	HelpAdminUsers         MessageId = "help.admin.users"
	HelpAbout              MessageId = "help.about"
	HelpAutoClose          MessageId = "help.autoclose"
	HelpAutoCloseExclude   MessageId = "help.autoclose.exclude"
	HelpAutoCloseConfigure MessageId = "help.autoclose.configure"
	HelpVote               MessageId = "help.vote"
	HelpAddAdmin           MessageId = "help.addadmin"
	HelpAddSupport         MessageId = "help.addsupport"
	HelpBlacklist          MessageId = "help.blacklist"
	HelpCancel             MessageId = "help.cancel"
	HelpPanel              MessageId = "help.panel"
	HelpPremium            MessageId = "help.premium"
	HelpRemoveSupport      MessageId = "help.removesupport"
	HelpSetup              MessageId = "help.setup"
	HelpViewStaff          MessageId = "help.viewstaff"
	HelpStats              MessageId = "help.stats"
	HelpStatsServer        MessageId = "help.statsserver"
	HelpManageTags         MessageId = "help.managetags"
	HelpTagAdd             MessageId = "help.taggadd"
	HelpTagDelete          MessageId = "help.tagdelete"
	HelpTagList            MessageId = "help.taglist"
	HelpTag                MessageId = "help.tag"
	HelpAdd                MessageId = "help.add"
	HelpClaim              MessageId = "help.claim"
	HelpClose              MessageId = "help.close"
	HelpOpen               MessageId = "help.open"
	HelpRemove             MessageId = "help.remove"
	HelpRename             MessageId = "help.rename"
	HelpTransfer           MessageId = "help.transfer"
	HelpUnclaim            MessageId = "help.unclaim"
	HelpHelp               MessageId = "help.help"
	HelpRemoveAdmin        MessageId = "help.removeadmin"
	HelpLanguage           MessageId = "help.language"

	MessageLanguageInvalidLanguage MessageId = "commands.language.invalid"
	MessageLanguageHelpWanted      MessageId = "commands.language.help_wanted"

	HelpAdminSetMessage   MessageId = "help.admin.set_message"
	HelpAdminCheckPremium MessageId = "help.admin.check_premium"
	HelpAdminBlacklist    MessageId = "help.admin.blacklist"
	HelpAdminUnblacklist  MessageId = "help.admin.unblacklist"

	SetupChoose                    MessageId = "setup.info.choose"
	SetupAutoDescription           MessageId = "setup.info.auto"
	SetupPrefixDescription         MessageId = "setup.info.prefix"
	SetupDashboardDescription      MessageId = "setup.info.dashboard"
	SetupLimitDescription          MessageId = "setup.info.ticket_limit"
	SetupWelcomeMessageDescription MessageId = "setup.info.welcome_message"
	SetupTranscriptsDescription    MessageId = "setup.info.transcripts"
	SetupCategoryDescription       MessageId = "setup.info.category"
	SetupReactionPanelsDescription MessageId = "setup.info.panels"

	SetupAutoRolesSuccess             MessageId = "setup.auto.roles.success"
	SetupAutoRolesFailure             MessageId = "setup.auto.roles.failure"
	SetupAutoTranscriptChannelSuccess MessageId = "setup.auto.transcript.success"
	SetupAutoTranscriptChannelFailure MessageId = "setup.auto.transcript.failure"
	SetupAutoCategorySuccess          MessageId = "setup.auto.category.success"
	SetupAutoCategoryFailure          MessageId = "setup.auto.category.failure"
	SetupAutoCompleted                MessageId = "setup.auto.completed"

	SetupPrefixInvalid  MessageId = "setup.prefix.invalid"
	SetupPrefixComplete MessageId = "setup.prefix.success"

	SetupWelcomeMessageInvalid  MessageId = "setup.welcome_message.invalid"
	SetupWelcomeMessageComplete MessageId = "setup.welcome_message.success"

	SetupLimitInvalid  MessageId = "setup.ticket_limit.invalid"
	SetupLimitComplete MessageId = "setup.ticket_limit.success"

	SetupTranscriptsInvalid  MessageId = "setup.transcript.invalid"
	SetupTranscriptsComplete MessageId = "setup.transcript.success"

	SetupCategoryInvalid  MessageId = "setup.category.invalid"
	SetupCategoryComplete MessageId = "setup.category.success"

	MessageOwnerIsAlreadyAdmin MessageId = "commands.addadmin.owner"
	MessageHelpInvite          MessageId = "help.invite"
	MessageInvite              MessageId = "commands.invite"
)

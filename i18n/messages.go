package i18n

type MessageId string

// note: %s = a placeholder
var (
	MessageNoPermission MessageId = "generic.no_permission"
	MessageOwnerOnly    MessageId = "generic.owner_only"

	Error   MessageId = "generic.error"
	Success MessageId = "generic.success"
	Admin   MessageId = "generic.admin"
	Ticket  MessageId = "generic.ticket"

	TitlePremiumOnly       MessageId = "generic.title.premium_only"
	TitleAbout             MessageId = "generic.title.about"
	TitleVote              MessageId = "generic.title.vote"
	TitleTags              MessageId = "generic.title.tags"
	TitleAutoclose         MessageId = "generic.title.autoclose"
	TitleInvite            MessageId = "generic.title.invite"
	TitleClose             MessageId = "generic.title.close"
	TitleClaim             MessageId = "generic.title.claim"
	TitleBlacklist         MessageId = "generic.title.blacklist"
	TitleBlacklisted       MessageId = "generic.title.blacklisted"
	TitleAddAdmin          MessageId = "generic.title.add_admin"
	TitleAddSupport        MessageId = "generic.title.add_support"
	TitleRemoveAdmin       MessageId = "generic.title.remove_admin"
	TitleRemoveSupport     MessageId = "generic.title.remove_support"
	TitleLanguage          MessageId = "generic.title.language"
	TitleSetup             MessageId = "generic.title.setup"
	TitlePremium           MessageId = "generic.title.premium"
	TitlePanel             MessageId = "generic.title.panel"
	TitleRemove            MessageId = "generic.title.remove"
	TitleRename            MessageId = "generic.title.rename"
	TitleAdd               MessageId = "generic.title.add"
	TitleClaimed           MessageId = "generic.title.claimed"
	TitleUnclaimed         MessageId = "generic.title.unclaimed"
	TitleCloseConfirmation MessageId = "generic.title.close_confirmation"
	TitleHelp              MessageId = "generic.title.help"
	TitleCloseRequest      MessageId = "generic.title.close_request"

	MessageAbout   MessageId = "commands.about"
	MessagePremium MessageId = "commands.premium"

	MessageVote               MessageId = "commands.vote"
	MessageInvalidArgument    MessageId = "generic.invalid_argument"
	MessageCloseNoPermission  MessageId = "close.no_permission"
	MessageCloseReasonTooLong MessageId = "close.reason_too_long"
	MessageCloseConfirmation  MessageId = "close.confirmation"

	MessageTag                       MessageId = "commands.tag.generic"
	MessageTagCreateInvalidArguments MessageId = "commands.tags.create.invalid_arguments"
	MessageTagCreateTooLong          MessageId = "commands.tags.create.too_long"
	MessageTagCreateAlreadyExists    MessageId = "commands.tags.create.already_exists"
	MessageTagCreateSuccess          MessageId = "commands.tags.create.success"

	MessageTagDeleteInvalidArguments MessageId = "commands.tags.delete.invalid_arguments"
	MessageTagDeleteDoesNotExist     MessageId = "commands.tags.delete.not_exist"
	MessageTagDeleteSuccess          MessageId = "commands.tags.delete.success"

	MessageTagList MessageId = "commands.tags.list"

	MessageTagInvalidArguments MessageId = "commands.tags.get.invalid_arguments"
	MessageTagInvalidTag       MessageId = "commands.tags.get.invalid_tag"

	MessageOpenRatelimited MessageId = "open.ratelimited"
	MessageTicketOpened    MessageId = "open.success"

	MessageAddAdminNoMembers   MessageId = "commands.addadmin.no_members"
	MessageAddAdminSuccess     MessageId = "commands.addadmin.success"
	MessageAddSupportNoMembers MessageId = "commands.addsupport.no_members"
	MessageAddSupportSuccess   MessageId = "commands.addsupport.success"

	MessageAddNoMembers        MessageId = "commands.add.no_members"
	MessageAddNoChannel        MessageId = "commands.add.no_channel"
	MessageAddChannelNotTicket MessageId = "commands.add.not_ticket"
	MessageAddNoPermission     MessageId = "commands.add.no_permission"
	MessageAddSuccess          MessageId = "commands.add.success"

	MessageBlacklisted        MessageId = "generic.error.blacklisted"
	MessageBlacklistNoMembers MessageId = "commands.blacklist.no_members"
	MessageBlacklistSelf      MessageId = "commands.blacklist.self"
	MessageBlacklistStaff     MessageId = "command.blacklist.staff"
	MessageBlacklistAdd       MessageId = "commands.blacklist.add.success"
	MessageBlacklistRemove    MessageId = "commands.blacklist.remove.success"

	MessageClaimed           MessageId = "commands.claim.success"
	MessageClaimNoPermission MessageId = "commands.claim.no_permission"

	MessagePanel MessageId = "commands.panel"

	MessageAlreadyPremium    MessageId = "commands.premium.already_premium"
	MessageInvalidPremiumKey MessageId = "commands.premium.invalid_key"
	MessagePremiumSuccess    MessageId = "commands.premium.success"

	MessageRemoveAdminNoMembers MessageId = "commands.removeadmin.no_members"
	MessageRemoveAdminSuccess   MessageId = "commands.removeadmin.success"

	MessageOwnerMustBeAdmin MessageId = "commands.removeadmin.owner"
	MessageRemoveStaffSelf  MessageId = "commands.removeadmin.self"

	MessageRemoveSupportNoMembers MessageId = "commands.removesupport.no_members"
	MessageRemoveSupportSuccess   MessageId = "commands.removesupport.success"

	MessageRemoveNoPermission      MessageId = "commands.remove.no_permission"
	MessageRemoveCannotRemoveStaff MessageId = "commands.remove.staff"
	MessageRemoveSuccess           MessageId = "commands.remove.success"

	MessageRenamed           MessageId = "commands.rename.success"
	MessageRenameMissingName MessageId = "commands.rename.missing_name"
	MessageRenameTooLong     MessageId = "commands.rename.too_long"

	MessageNotClaimed            MessageId = "commands.unclaim.not_claimed"
	MessageOnlyClaimerCanUnclaim MessageId = "commands.unclaim.not_claimer"
	MessageUnclaimed             MessageId = "commands.unclaim.success"

	MessageNotATicketChannel MessageId = "generic.not_ticket"
	MessageInvalidUser       MessageId = "generic.invalid_user"

	MessageTicketLimitReached MessageId = "commands.open.ticket_limit"
	MessageTooManyTickets     MessageId = "commands.open.too_many_tickets"
	MessageTicketStartedFrom  MessageId = "commands.open.from"
	MessageMovedToTicket      MessageId = "commands.open.from.moved"

	MessageCloseRequestNoReason     MessageId = "commands.close_request.no_reason"
	MessageCloseRequestWithReason   MessageId = "commands.close_request.with_reason"
	MessageCloseRequestNoPermission MessageId = "commands.close_request.no_permission"
	MessageCloseRequestDenied       MessageId = "commands.close_request.denied"
	MessageCloseRequestAccept       MessageId = "commands.close_request.accept"
	MessageCloseRequestDeny         MessageId = "commands.close_request.deny"

	MessageAutoCloseConfigure MessageId = "commands.autoclose.configure"
	MessageAutoCloseExclude   MessageId = "commands.autoclose.exclude.success"

	SetupArchiveChannel  MessageId = "setup.info.archive_channel"
	SetupChannelCategory MessageId = "setup.info.category"
	SetupPrefix          MessageId = "setup.info.prefix"
	SetupTicketLimit     MessageId = "setup.info.ticket_limit"
	SetupWelcomeMessage  MessageId = "setup.info.welcome_message"

	MessageLanguageInvalidLanguage MessageId = "commands.language.invalid"
	MessageLanguageHelpWanted      MessageId = "commands.language.help_wanted"
	MessageLanguageSuccess         MessageId = "commands.language.success"

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

	MessageFeedbackDisabled MessageId = "feedback.disabled"
	MessageFeedbackSuccess  MessageId = "feedback.success"

	MessageButtonGuildOnly MessageId = "button.guild_only"
	MessageButtonDMOnly    MessageId = "button.dms_only"

	HelpAdmin              MessageId = "help.admin"
	HelpAdminForceClose    MessageId = "help.admin.force_close"
	HelpAdminGenPremium    MessageId = "help.admin.generate_premium"
	HelpAdminGetOwner      MessageId = "help.admin.get_owner"
	HelpAdminUpdateSchema  MessageId = "help.admin.update_schema"
	HelpAbout              MessageId = "help.about"
	HelpAutoClose          MessageId = "help.autoclose"
	HelpAutoCloseExclude   MessageId = "help.autoclose.exclude"
	HelpAutoCloseConfigure MessageId = "help.autoclose.configure"
	HelpVote               MessageId = "help.vote"
	HelpAddAdmin           MessageId = "help.addadmin"
	HelpAddSupport         MessageId = "help.addsupport"
	HelpBlacklist          MessageId = "help.blacklist"
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
	HelpCloseRequest       MessageId = "help.close_request"
	HelpOpen               MessageId = "help.open"
	HelpRemove             MessageId = "help.remove"
	HelpRename             MessageId = "help.rename"
	HelpTransfer           MessageId = "help.transfer"
	HelpUnclaim            MessageId = "help.unclaim"
	HelpHelp               MessageId = "help.help"
	HelpRemoveAdmin        MessageId = "help.removeadmin"
	HelpLanguage           MessageId = "help.language"
)

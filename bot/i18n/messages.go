package i18n

import (
	translations "github.com/TicketsBot/database/translations"
)

// note: %s = a placeholder
const (
	MessageNoPermission translations.MessageId = iota // You do not have permission for this
	MessageOwnerOnly                                  // This command is reserved for the bot owner only

	/*
		note: write the message on one line, and use \n to force a linebreak.

		Tickets is an easy to use and highly configurable ticket and support bot.
		Support server: https://discord.gg/VtV3rSk
		Commands: Type `t!help`
		Get started by running `t!setup`
	*/
	MessageAbout

	/*
		note: write the message on one line, and use \n to force a linebreak. please leave the + symbols at the beginning of lines also, it makes the line green.

		You can purchase a premium subscription from <https://www.patreon.com/ticketsbot>.
		Premium provides many benefits, such as:
		```diff
		+ Create unlimited ticket panels
		+ Customise bot name, avatar & status (whitelabel)
		+ Detailed statistics about the server, users and support staff
		+ No branding in the footer of messages
		+ Support development and help me pay the server costs
		```

		Alternatively, you can vote at <https://vote.ticketsbot.net> to get premium free for 24 hours
	*/
	MessagePremium

	MessageModmailOpened     // Your modmail ticket in %s has been opened! Use `t!close` to close the session.
	MessageVote              // Click here to vote for 24 hours of free premium:\nhttps://vote.ticketsbot.net
	MessageInvalidArgument   // Invalid argument: refer to usage
	MessageCloseNoPermission // You are not permitted to close this ticket

	MessageTagCreateInvalidArguments // You must specify a tag ID and contents
	MessageTagCreateTooLong          // Tag IDs cannot be longer than 16 characters
	MessageTagCreateAlreadyExists    // A tag with the ID `%s` already exists. You can delete the tag using `t!tag delete %s`

	MessageTagDeleteInvalidArguments // You must specify a tag ID to delete
	MessageTagDeleteDoesNotExist     // A tag with the ID `%s` could not be found

	// note: the first %s is a list of tag IDs - you do not need to worry about these. The second %s is the bot's prefix.
	MessageTagList // IDs for all tags: \n%s\nTo view the contents of a tag, run `%stag <ID>`

	MessageTagInvalidArguments // You must provide the ID of the tag. For more help with tag, visit <https://ticketsbot.net/tags>.
	MessageTagInvalidTag       // Invalid tag. For more help with tags, visit <https://ticketsbot.net/tags>.

	// note: %s will be the channel mention
	MessageTicketOpened // Opened a new ticket: %s

	MessageAddAdminNoMembers   // You need to mention a user or name a role to grant admin privileges to
	MessageAddSupportNoMembers // You need to mention a user or name a role to grant support representative privileges to

	MessageAddNoMembers        // You need to mention members to add to the ticket
	MessageAddNoChannel        // You need to mention a ticket channel to add the user(s) in
	MessageAddChannelNotTicket // The mentioned channel is not a ticket
	MessageAddNoPermission     // You don't have permission to add people to this ticket

	MessageBlacklisted        // You are blacklisted in this server!
	MessageBlacklistNoMembers // You need to mention a user to toggle the blacklist state for
	MessageBlacklistSelf      // You cannot blacklist yourself
	MessageBlacklistStaff     // You cannot blacklist staff

	// note: %s will be a user mention
	MessageClaimed // Your ticket will be handled by %s

	MessageHelpDMFailed // I couldn't send you a direct message: make sure your privacy settings aren't too high

	// note: %d will be the server ID. please leave the URL alone, and leave it wrapped in angle brackets.
	MessagePanel // Visit <https://panel.ticketsbot.net/manage/%d/panels> to configure a panel

	// note: %s is the timestamp when premium expires
	MessageAlreadyPremium // This guild already has premium. It expires on %s

	MessageInvalidPremiumKey // Invalid key. Ensure that you have copied it correctly.

	MessageRemoveAdminNoMembers // You need to mention a user or name a role to revoke admin privileges from
	MessageOwnerMustBeAdmin     // The guild owner must be an admin
	MessageRemoveStaffSelf      // You cannot revoke your own privileges

	MessageRemoveSupportNoMembers // You need to mention a user or name a role to grant support representative privileges to

	MessageRemoveNoMembers         // You need to mention members to remove from the ticket
	MessageRemoveNoPermission      // You don't have permission to remove people from this ticket
	MessageRemoveCannotRemoveStaff // You cannot remove staff from a ticket

	// note: MAKE SURE YOU KEEP THE <#%d> FORMAT. This will mention the channel.
	MessageRenamed           // This ticket has been renamed to <#%d>
	MessageRenameMissingName // You need to specify a new name for this ticket

	// note: %s will be the prefix
	MessageAlreadyInSetup // You are already in setup mode (use `%scancel` to exit)

	MessageNotClaimed            // This ticket is not claimed
	MessageOnlyClaimerCanUnclaim // Only admins and the user who claimed the ticket can unclaim the ticket
	MessageUnclaimed             // All support representatives can now respond to the ticket

	MessageDisabledLogChannel // Invalid channel, disabling log channel
	MessageInvalidCategory    // Invalid category\nDefaulting to using no category

	// note: %s is the category name
	MessageCreatedCategory // I have created the channel category %s for you, you may need to adjust permissions yourself

	// note: %s is the default prefix
	MessageInvalidPrefix // The maximum prefix length is 8 characters\nDefaulting to `%s`

	// note: %d is the default ticket limit
	MessageInvalidTicketLimit // Invalid ticket limit (must be 1-10). Defaulting to `%d`

	MessageNotATicketChannel // This is not a ticket channel
	MessageInvalidUser       // Couldn't find the target user

	// note: %d will be replaced with the ticket limit, %s will be replaced with "ticket" or "tickets", for grammar purposes
	MessageTicketLimitReached // You are only able to open %d %s at once

	MessageTooManyTickets // There are too many tickets in the ticket category. Ask an admin to close some, or to move them to another category

	SetupArchiveChannel  // Please specify you wish ticket logs to be sent to after tickets have been closed\nExample: `#logs`
	SetupChannelCategory // Type the **name** of the **channel category** that you would like tickets to be created under
	SetupPrefix          // Type the prefix that you would like to use for the bot\nThe prefix is the characters that come *before* the command (excluding the actual command itself)\nExample: `t!`
	SetupTicketLimit     // Specify the maximum amount of tickets that a **single user** should be able to have open at once
	SetupWelcomeMessage  // Type the message that should be sent by the bot when a ticket is opened

	HelpAdmin             // Bot management
	HelpAdminCheckPerms   // Checks permissions for the bot on the channel
	HelpAdminDebug        // Provides debugging information
	HelpAdminForceClose   // Sets the state of the provided tickets to closed
	HelpAdminGC           // Forces a GC sweep
	HelpAdminGenPremium   // Generate premium keys
	HelpAdminGetOwner     // Gets the owner of a server
	HelpAdminPing         // Measures WS latency to Discord
	HelpAdminSeed         // Seeds the cache with members
	HelpAdminUpdateSchema // Updates the database schema
	HelpAdminUsers        // Prints the total seen member count
	HelpAbout             // Tells you information about the bot
	HelpVote              // Gives you a link to vote for free premium
	HelpAddAdmin          // Grants a user or role admin privileges
	HelpAddSupport        // Adds a user or role as a support representative
	HelpBlacklist         // Toggles whether users are allowed to interact with the bot
	HelpCancel            // Cancels the setup process
	HelpPanel             // Creates a panel to enable users to open a ticket with 1 click
	HelpPremium           // Activate a premium key for your server
	HelpRemoveSupport     // Revokes a user's or role's support representative privileges
	HelpSetup             // Allows you to easily configure the bot
	HelpSync              // Syncs the bot's database to the channels - useful if you a Discord outage has taken place
	HelpViewStaff         // Lists the staff members and roles
	HelpStats             // Shows you statistics about users, support staff and the server
	HelpStatsServer       // Shows you statistics about the server
	HelpManageTags        // Add, delete or list tags
	HelpTagAdd            // Adds a new tag
	HelpTagDelete         // Deletes a tag
	HelpTagList           // Lists all tags
	HelpTag               // Sends a message snippet
	HelpAdd               // Adds a user to a ticket
	HelpClaim             // Assigns a single staff member to a ticket
	HelpClose             // Closes the current ticket
	HelpOpen              // Opens a new ticket
	HelpRemove            // Removes a user from a ticket
	HelpRename            // Renames the current ticket
	HelpTransfer          // Transfers a claimed ticket to another user
	HelpUnclaim           // Removes the claim on the current ticket
	HelpHelp              // Shows you a list of commands
	HelpRemoveAdmin       // Revokes a user's or role's admin privileges
	HelpLanguage          // Changes the language the bot uses

	// note: %s will be a list of languages. You do not need to worry about this.
	MessageLanguageInvalidLanguage // You need to specify a language code or flag. Available languages:\n%s

	HelpAdminSetMessage   // Override a message
	HelpAdminCheckPremium // Check the premium status of a server
	HelpAdminBlacklist    // Blacklist a guild from using the bot
	HelpAdminUnblacklist  // Remove a guild from the guild blacklist

	HelpSetupEasy                  // Prompts you with questions to configure the bot
	SetupChoose                    // Please select an option. You can use t!setup or t!auto to get up and running quickly, or use the other commands to fine tune settings
	SetupEasyDescription           // Prompts you with a series of questions to set up the bot
	SetupAutoDescription           // Automatically selects options & creates the required channels
	SetupPrefixDescription         // Changes the prefix (e.g. `t!`) of the bot.\n**Example:** `t!setup prefix -`
	SetupDashboardDescription      // More settings are available on the [dashboard](https://panel.ticketsbot.net)
	SetupLimitDescription          // Sets the maximum amount of tickets a **single user** can open **at once**.\n**Example:** `t!setup limit 3`
	SetupWelcomeMessageDescription // Sets the message sent to tickets when opened.\n**Example:** `t!setup welcomemessage Thanks for opening a ticket!`
	SetupTranscriptsDescription    // Sets the channel ticket transcripts are sent to.\n**Example:** `t!setup transcripts #logs`
	SetupCategoryDescription       // Sets the channel category tickets are created under.\n**Example:** `t!setup transcripts Support`
	SetupReactionPanelsDescription // Visit the [dashboard](https://panel.ticketsbot.net/maange/%d/panels) to create reaction panels

	SetupAutoRolesSuccess             // Created `Tickets Admin` & `Tickets Support` roles
	SetupAutoRolesFailure             // I am missing the `Manage Roles` permission
	SetupAutoTranscriptChannelSuccess // Created <#%d> channel
	SetupAutoTranscriptChannelFailure // I am missing the `Manage Channels` permission
	SetupAutoCategorySuccess          // Created Tickets category
	SetupAutoCategoryFailure          // I am missing the `Manage Channels` permission
	SetupAutoCompleted                // Setup complete! Click [here](https://panel.ticketsbot.net/manage/%d/panels) to setup a reaction panel, and add the <@&%d> and <@&%d> roles to your staff members

	SetupPrefixInvalid  // Prefixes must be a maximum of 8 characters and **not** include the command\n**Example:** `t!setup prefix -`
	SetupPrefixComplete // The prefix has been changed to %s. You can use `%sopen` to open a ticket

	SetupWelcomeMessageInvalid  // Welcome messages must be a maximum of 1024 characters long\n**Example:**`t!setup wm Thanks for opening a ticket!`
	SetupWelcomeMessageComplete // The welcome message has been updated. Open a ticket to see it in action

	SetupLimitInvalid  // The ticket limit must be a number in the range 1-10. It is a **per-user** limit for tickets open **at one time**\n**Example:** `t!setup limit 1`
	SetupLimitComplete // The ticket limit has been updated to `%d`

	SetupTranscriptsInvalid  // You must mention a valid channel in the server (e.g. <#%d>)\n**Example:** `t!setup transcripts #logs`
	SetupTranscriptsComplete // The transcripts channel has been changed to <#%d>

	SetupCategoryInvalid  // You must provide the name of a [**channel category**](https://support.discord.com/hc/en-us/articles/115001580171-Channel-Categories-101)\n**Example:** `t!setup category Tickets`
	SetupCategoryComplete // The ticket channel category has been changed to `%s`

	MessageOwnerIsAlreadyAdmin // The owner of the server already inherits administrator permissions
	MessageHelpInvite          // Provides an invite link for the bot
	MessageInvite              // [Click here](<https://invite.ticketsbot.net>)
)

var Messages = map[string]translations.MessageId{
	"No permission":                            MessageNoPermission,
	"Owner only command":                       MessageOwnerOnly,
	"About command":                            MessageAbout,
	"Premium features":                         MessagePremium,
	"Modmail opened":                           MessageModmailOpened,
	"Vote":                                     MessageVote,
	"Invalid argument":                         MessageInvalidArgument,
	"t!close no permission":                    MessageCloseNoPermission,
	"Tag create invalid arguments":             MessageTagCreateInvalidArguments,
	"Tag create ID too long":                   MessageTagCreateTooLong,
	"Tag create ID already exists":             MessageTagCreateAlreadyExists,
	"Tag delete missing ID":                    MessageTagDeleteInvalidArguments,
	"Tag delete ID doesn't exist":              MessageTagDeleteDoesNotExist,
	"Tag list":                                 MessageTagList,
	"Tag missing ID":                           MessageTagInvalidArguments,
	"Tag invalid ID":                           MessageTagInvalidTag,
	"Ticket opened":                            MessageTicketOpened,
	"t!addadmin no mentions":                   MessageAddAdminNoMembers,
	"t!addsupport no mentions":                 MessageAddSupportNoMembers,
	"t!add no users":                           MessageAddNoMembers,
	"t!add no channel":                         MessageAddNoChannel,
	"t!add channel not a ticket":               MessageAddChannelNotTicket,
	"t!add no permission":                      MessageAddNoPermission,
	"You are blacklisted":                      MessageBlacklisted,
	"t!blacklist no mentions":                  MessageBlacklistNoMembers,
	"t!blacklist on self":                      MessageBlacklistSelf,
	"t!blacklist on staff":                     MessageBlacklistStaff,
	"t!claim":                                  MessageClaimed,
	"t!help privacy settings too high":         MessageHelpDMFailed,
	"t!panel":                                  MessagePanel,
	"t!premium already premium":                MessageAlreadyPremium,
	"t!premium invalid key":                    MessageInvalidPremiumKey,
	"t!removeadmin no mentions":                MessageRemoveAdminNoMembers,
	"t!removeadmin on owner":                   MessageOwnerMustBeAdmin,
	"t!removeadmin on self":                    MessageRemoveStaffSelf,
	"t!removesupport no mentions":              MessageRemoveSupportNoMembers,
	"t!remove no mentions":                     MessageRemoveNoMembers,
	"t!remove no permission":                   MessageRemoveNoPermission,
	"t!remove on staff":                        MessageRemoveCannotRemoveStaff,
	"t!rename success":                         MessageRenamed,
	"t!rename missing name":                    MessageRenameMissingName,
	"t!setup already in setup":                 MessageAlreadyInSetup,
	"t!unclaim already not claimed":            MessageNotClaimed,
	"t!unclaim only claimed can unclaim":       MessageOnlyClaimerCanUnclaim,
	"t!unclaim success":                        MessageUnclaimed,
	"Disabled log channel":                     MessageDisabledLogChannel,
	"Invalid category":                         MessageInvalidCategory,
	"Created category":                         MessageCreatedCategory,
	"Invalid prefix":                           MessageInvalidPrefix,
	"Invalid ticket limit":                     MessageInvalidTicketLimit,
	"Not a ticket channel":                     MessageNotATicketChannel,
	"Invalid user":                             MessageInvalidUser,
	"Ticket limit reached":                     MessageTicketLimitReached,
	"Too many tickets":                         MessageTooManyTickets,
	"t!setup: Archive channel":                 SetupArchiveChannel,
	"t!setup: Channel category":                SetupChannelCategory,
	"t!setup: Prefix":                          SetupPrefix,
	"t!setup: Ticket limit":                    SetupTicketLimit,
	"t!setup: Welcome message":                 SetupWelcomeMessage,
	"Help: t!admin":                            HelpAdmin,
	"Help: t!admin checkperms":                 HelpAdminCheckPerms,
	"Help: t!admin debug":                      HelpAdminDebug,
	"Help: t!admin forceclose":                 HelpAdminForceClose,
	"Help: t!admin gc":                         HelpAdminGC,
	"Help: t!admin genpremium":                 HelpAdminGenPremium,
	"Help: t!admin getowner":                   HelpAdminGetOwner,
	"Help: t!admin ping":                       HelpAdminPing,
	"Help: t!admin seed":                       HelpAdminSeed,
	"Help: t!admin updateschema":               HelpAdminUpdateSchema,
	"Help: t!admin users":                      HelpAdminUsers,
	"Help: t!about":                            HelpAbout,
	"Help: t!vote":                             HelpVote,
	"Help: t!addadmin":                         HelpAddAdmin,
	"Help: t!addsupport":                       HelpAddSupport,
	"Help: t!blacklist":                        HelpBlacklist,
	"Help: t!cancel":                           HelpCancel,
	"Help: t!panel":                            HelpPanel,
	"Help: t!premium":                          HelpPremium,
	"Help: t!removesupport":                    HelpRemoveSupport,
	"Help: t!setup":                            HelpSetup,
	"Help: t!sync":                             HelpSync,
	"Help: t!viewstaff":                        HelpViewStaff,
	"Help: t!stats":                            HelpStats,
	"Help: t!stats server":                     HelpStatsServer,
	"Help: t!managetags":                       HelpManageTags,
	"Help: t!managetags add":                   HelpTagAdd,
	"Help: t!managetags delete":                HelpTagDelete,
	"Help: t!managetags list":                  HelpTagList,
	"Help: t!tag":                              HelpTag,
	"Help: t!add":                              HelpAdd,
	"Help: t!claim":                            HelpClaim,
	"Help: t!close":                            HelpClose,
	"Help: t!open":                             HelpOpen,
	"Help: t!remove":                           HelpRemove,
	"Help: t!rename":                           HelpRename,
	"Help: t!transfer":                         HelpTransfer,
	"Help: t!unclaim":                          HelpUnclaim,
	"Help: t!help":                             HelpHelp,
	"Help: t!removeadmin":                      HelpRemoveAdmin,
	"Help: t!language":                         HelpLanguage,
	"t!language invalid language":              MessageLanguageInvalidLanguage,
	"Help: t!admin setmessage":                 HelpAdminSetMessage,
	"Help: t!admin checkpremium":               HelpAdminCheckPremium,
	"Help: t!admin blacklist":                  HelpAdminBlacklist,
	"Help: t!admin unblacklist":                HelpAdminUnblacklist,
	"Help: t!setup ez":                         HelpSetupEasy,
	"t!setup: choose subcommand":               SetupChoose,
	"t!setup: ez description":                  SetupEasyDescription,
	"t!setup: auto description":                SetupAutoDescription,
	"t!setup: prefix description":              SetupPrefixDescription,
	"t!setup: dashboard description":           SetupDashboardDescription,
	"t!setup: limit description":               SetupLimitDescription,
	"t!setup: welcome message description":     SetupWelcomeMessageDescription,
	"t!setup: transcripts description":         SetupTranscriptsDescription,
	"t!setup: category description":            SetupCategoryDescription,
	"t!setup: reaction panels description":     SetupReactionPanelsDescription,
	"t!setup auto: roles success":              SetupAutoRolesSuccess,
	"t!setup auto: roles failure":              SetupAutoRolesFailure,
	"t!setup auto: transcript channel success": SetupAutoTranscriptChannelSuccess,
	"t!setup auto: transcript channel failure": SetupAutoTranscriptChannelFailure,
	"t!setup auto: category success":           SetupAutoCategorySuccess,
	"t!setup auto: category failure":           SetupAutoCategoryFailure,
	"t!setup auto: completed":                  SetupAutoCompleted,
	"t!setup prefix: invalid":                  SetupPrefixInvalid,
	"t!setup prefix: success":                  SetupPrefixComplete,
	"t!setup welcome message: invalid":         SetupWelcomeMessageInvalid,
	"t!setup welcome message: success":         SetupWelcomeMessageComplete,
	"t!setup ticket limit: invalid":            SetupLimitInvalid,
	"t!setup ticket limit: success":            SetupLimitComplete,
	"t!setup transcripts: invalid":             SetupTranscriptsInvalid,
	"t!setup transcripts: success":             SetupTranscriptsComplete,
	"t!setup category: invalid":                SetupCategoryInvalid,
	"t!setup category: success":                SetupCategoryComplete,
	"t!addadmin: owner already admin":          MessageOwnerIsAlreadyAdmin,
	"t!help: t!invite":                         MessageHelpInvite,
	"t!invite":                                 MessageInvite,
}

func GetSimpleName(id translations.MessageId) *string {
	for simpleName, messageId := range Messages {
		if id == messageId {
			return &simpleName
		}
	}

	return nil
}

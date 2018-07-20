package gdrive2discord

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"../discord"
	"../google/drive"
	"../google/userinfo"
)

var actionColors = []string{
	drive.Deleted:  "#ffcccc",
	drive.Created:  "#ccffcc",
	drive.Modified: "#ccccff",
	drive.Shared:   "#ccccff",
	drive.Viewed:   "#ccccff",
}

func infixZeroWidthSpace(source string) string {
	if utf8.RuneCountInString(source) < 2 {
		return source
	}
	firstRune, width := utf8.DecodeRuneInString(source)
	rest := source[width:]
	return string(firstRune) + "\u200B" + rest
}

func preventNotification(source string) string {
	split := strings.Split(source, " ")
	for i, word := range split {
		split[i] = infixZeroWidthSpace(word)
	}
	return strings.Join(split, " ")
}

func CreateDiscordAttachment(change *drive.ChangeItem) *discord.Attachment {
	var editor string
	if len(change.File.LastModifyingUser.EmailAddress) > 0 && len(change.File.LastModifyingUser.DisplayName) > 0 {
		editor = fmt.Sprintf("%s - **%s**", change.File.LastModifyingUser.EmailAddress, preventNotification(change.File.LastModifyingUser.DisplayName))
	} else if len(change.File.LastModifyingUser.DisplayName) > 0 {
		editor = preventNotification(change.File.LastModifyingUser.DisplayName)
	} else {
		editor = "Unknown"
	}
	return &discord.Attachment{
		Fallback: fmt.Sprintf("➡️ __Changes Detected to %s :__ %s - %s", change.Type, change.File.AlternateLink, change.File.Title),
		Fields: []discord.Field{
			{
				Title: fmt.Sprintf("%s %s", change.LastAction, change.Type),
				Value: fmt.Sprintf("%s - **%s**", change.File.AlternateLink, change.File.Title),
				Short: true,
			},
			{
				Title: "Editor",
				Value: editor,
				Short: true,
			},
		},
	}
}

func CreateDiscordMessage(subscription *Subscription, userState *UserState, folders *drive.Folders, version string) *discord.Message {
	var attachments = make([]discord.Attachment, 0, len(userState.Gdrive.ChangeSet))
	var roots = subscription.GoogleInterestingFolderIds
	for i := 0; i != len(userState.Gdrive.ChangeSet); i++ {
		if len(roots) == 0 || folders.FolderIsOrIsContainedInAny(userState.Gdrive.ChangeSet[i].File.Parents, roots) {
			attachments = append(attachments, *CreateDiscordAttachment(&userState.Gdrive.ChangeSet[i]))
		}

	}
	return &discord.Message{
		Username:    "Google Drive",
		Text:        "Activity on gdrive:",
		IconUrl:     fmt.Sprintf("https://giveawaynetwork.xyz/assets/img/google-drive-logo.png", version),
		Attachments: attachments,
	}
}

func CreateDiscordWelcomeMessage(redirectUri string, gUserInfo *userinfo.UserInfo, version string) *discord.Message {
	return &discord.Message{
		Username: "Google Drive",
		Text:     fmt.Sprintf("A %s integration has been configured by %s. Activities on Google Drive documents will be notified here. Forked & adapted by Mxb", redirectUri, gUserInfo.DisplayName),
		IconUrl:  fmt.Sprintf("https://giveawaynetwork.xyz/assets/img/google-drive-logo.png", version),
	}
}

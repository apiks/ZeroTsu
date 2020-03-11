package commands

import (
	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/functionality"
)

// Sorts all spoiler roles created with the create command between the two opt-in dummy roles alphabetically
func sortRolesCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		spoilerRoles      []*discordgo.Role
		underSpoilerRoles []*discordgo.Role
		rolesOrdered      []*discordgo.Role

		controlNum int
	)

	guildSettings := db.GetGuildSettings(m.GuildID)
	guildSpoilerMap := db.GetGuildSpoilerMap(m.GuildID)

	if len(guildSpoilerMap) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: No Spoiler roles detected. Please use `"+guildSettings.GetPrefix()+"create` command to create a valid role before using this command")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
		return
	}

	// Confirms whether optins exist
	err := common.OptInsHandler(s, m.ChannelID, m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Fetches info from the server and puts it in debPre
	debPre, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Refreshes the positions of all roles in the server (because when created roles start at 0)
	for i := 0; i < len(debPre); i++ {
		spoilerRoles = append(spoilerRoles, debPre[i])
	}

	// Pushes the refreshed positions to the server
	_, err = s.GuildRoleReorder(m.GuildID, spoilerRoles)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	time.Sleep(time.Millisecond * 333)

	// Resets the value of spoilerRoles
	spoilerRoles = nil

	// Fetches the refreshed info from the server and puts it in deb
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
		return
	}

	// Saves the original opt-in-above position
	for i := 0; i < len(deb); i++ {
		if deb[i].ID == guildSettings.GetOptInAbove().GetID() {
			optInAbove := guildSettings.GetOptInAbove()
			optInAbove = optInAbove.SetPosition(deb[i].Position)
			guildSettings = guildSettings.SetOptInAbove(optInAbove)

			db.SetGuildSettings(m.GuildID, guildSettings)
			break
		}
	}

	// Adds all spoiler roles in SpoilerMap in the spoilerRoles slice
	// Adds all non-spoiler roles under opt-in-above (including it) in the underSpoilerRoles slice
	for i := 0; i < len(deb); i++ {
		_, ok := guildSpoilerMap[deb[i].ID]
		if ok {
			spoilerRoles = append(spoilerRoles, guildSpoilerMap[deb[i].ID])
			if deb[i].Position < guildSettings.GetOptInAbove().GetPosition() {
				controlNum++
			}
		} else if !ok &&
			deb[i].Position <= guildSettings.GetOptInAbove().GetPosition() &&
			deb[i].ID != m.GuildID {
			underSpoilerRoles = append(underSpoilerRoles, deb[i])
		}
	}

	// If there are spoiler roles under opt-in-above it goes in to move and sort
	if controlNum > 0 {

		// Sorts the spoilerRoles slice (all spoiler roles) alphabetically
		sort.Sort(common.SortRoleByAlphabet(spoilerRoles))

		// Moves the sorted spoiler roles above opt-in-above
		for i := 0; i < len(spoilerRoles); i++ {
			spoilerRoles[i].Position = guildSettings.GetOptInAbove().GetPosition()
		}

		// Moves every non-spoiler role below opt-in-above (including it) down an amount equal to the amount of roles in the
		// spoilerRoles slice that are below opt-in-above
		for i := 0; i < len(underSpoilerRoles); i++ {
			underSpoilerRoles[i].Position = underSpoilerRoles[i].Position - controlNum
		}

		// Concatenates the two ordered role slices
		rolesOrdered = append(spoilerRoles, underSpoilerRoles...)

		//Pushes the ordered role list to the server
		_, err = s.GuildRoleReorder(m.GuildID, rolesOrdered)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		time.Sleep(time.Millisecond * 333)

		// Fetches info from the server and puts it in debPost
		debPost, err := s.GuildRoles(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		// Refreshes deb
		deb, err = s.GuildRoles(m.GuildID)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		// Saves the new opt-in-above position
		for i := 0; i < len(debPost); i++ {
			if deb[i].ID == guildSettings.GetOptInAbove().GetID() {
				optInAbove := guildSettings.GetOptInAbove()
				optInAbove = optInAbove.SetPosition(deb[i].Position)
				guildSettings = guildSettings.SetOptInAbove(optInAbove)

				db.SetGuildSettings(m.GuildID, guildSettings)
				break
			}
		}

		for i := range spoilerRoles {
			spoilerRoles[i].Position = guildSettings.GetOptInAbove().GetPosition() + len(spoilerRoles) - i
			db.SetGuildSpoilerRole(m.GuildID, spoilerRoles[i])
		}

		// Pushes the sorted list to the server
		_, err = s.GuildRoleReorder(m.GuildID, spoilerRoles)
		if err != nil {
			common.CommandErrorHandler(s, m, guildSettings.BotLog, err)
			return
		}

		if m.Author.ID == s.State.User.ID {
			return
		}

		_, err = s.ChannelMessageSend(m.ChannelID, "Roles sorted.")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Spoiler roles already sorted or the spoiler roles are above the opt-in-above (in which case please move them manually.)")
		if err != nil {
			common.LogError(s, guildSettings.BotLog, err)
			return
		}
	}
}

func init() {
	Add(&Command{
		Execute:    sortRolesCommand,
		Trigger:    "sortroles",
		Desc:       "Sorts all spoiler roles alphabetically between dummy opt-in roles",
		Permission: functionality.Mod,
		Module:     "misc",
	})
}

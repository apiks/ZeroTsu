package commands

import (
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

// Sorts all spoiler roles created with the create command between the two opt-in dummy roles alphabetically
func sortRolesCommand(s *discordgo.Session, m *discordgo.Message) {

	var (
		spoilerRoles      []*discordgo.Role
		underSpoilerRoles []*discordgo.Role
		rolesOrdered      []*discordgo.Role

		controlNum int
	)

	if misc.SpoilerMap == nil {

		return
	}

	// Fetches info from the server and puts it in debPre
	debPre, err := s.GuildRoles(config.ServerID)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Refreshes the positions of all roles in the server (because when created roles start at 0)
	for i := 0; i < len(debPre); i++ {

		spoilerRoles = append(spoilerRoles, debPre[i])
	}

	// Pushes the refreshed positions to the server
	_, err = s.GuildRoleReorder(config.ServerID, spoilerRoles)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	time.Sleep(time.Millisecond * 333)

	// Resets the value of spoilerRoles
	spoilerRoles = nil

	// Fetches the refreshed info from the server and puts it in deb
	deb, err := s.GuildRoles(config.ServerID)
	if err != nil {

		misc.CommandErrorHandler(s, m, err)
		return
	}

	// Saves the original opt-in-above position
	for i := 0; i < len(deb); i++ {
		if deb[i].Name == config.OptInAbove {

			misc.OptinAbovePosition = deb[i].Position
		}
	}

	// Adds all spoiler roles in SpoilerMap in the spoilerRoles slice
	// Adds all non-spoiler roles under opt-in-above (including it) in the underSpoilerRoles slice
	for i := 0; i < len(deb); i++ {
		_, ok := misc.SpoilerMap[deb[i].ID]
		if ok == true {

			spoilerRoles = append(spoilerRoles, misc.SpoilerMap[deb[i].ID])

			if deb[i].Position < misc.OptinAbovePosition {

				controlNum++
			}
		} else if ok == false && deb[i].Position <= misc.OptinAbovePosition && deb[i].ID != config.ServerID {

			underSpoilerRoles = append(underSpoilerRoles, deb[i])
		}
	}

	// If there are spoiler roles under opt-in-above it goes in to move and sort
	if controlNum > 0 {

		// Sorts the spoilerRoles slice (all spoiler roles) alphabetically
		sort.Sort(misc.SortRoleByAlphabet(spoilerRoles))

		// Moves the sorted spoiler roles above opt-in-above
		for i := 0; i < len(spoilerRoles); i++ {

			spoilerRoles[i].Position = misc.OptinAbovePosition
		}

		// Moves every non-spoiler role below opt-in-above (including it) down an amount equal to the amount of roles in the
		// spoilerRoles slice that are below opt-in-above
		for i := 0; i < len(underSpoilerRoles); i++ {

			underSpoilerRoles[i].Position = underSpoilerRoles[i].Position - controlNum
		}

		// Concatenates the two ordered role slices
		rolesOrdered = append(spoilerRoles, underSpoilerRoles...)

		//Pushes the ordered role list to the server
		_, err = s.GuildRoleReorder(config.ServerID, rolesOrdered)
		if err != nil {

			misc.CommandErrorHandler(s, m, err)
			return
		}

		time.Sleep(time.Millisecond * 333)

		// Fetches info from the server and puts it in debPost
		debPost, err := s.GuildRoles(config.ServerID)
		if err != nil {

			misc.CommandErrorHandler(s, m, err)
			return
		}

		// Saves the new opt-in-above position
		for i := 0; i < len(debPost); i++ {
			if deb[i].Name == config.OptInAbove {

				misc.OptinAbovePosition = deb[i].Position
			}
		}

		for i := 0; i < len(spoilerRoles); i++ {

			spoilerRoles[i].Position = misc.OptinAbovePosition + len(spoilerRoles) - i
			misc.SpoilerMap[spoilerRoles[i].ID].Position = spoilerRoles[i].Position
		}

		// Pushes the sorted list to the server
		_, err = s.GuildRoleReorder(config.ServerID, spoilerRoles)
		if err != nil {

			misc.CommandErrorHandler(s, m, err)
			return
		}

		time.Sleep(time.Millisecond * 333)

		if m.Author.ID == config.BotID {

			return
		}

		_, err = s.ChannelMessageSend(m.ChannelID, "Roles sorted.")
		if err != nil {

			_, err = s.ChannelMessageSend(config.BotLogID, err.Error())
			if err != nil {

				return
			}
			return
		}
	}
}

func init() {
	add(&command{
		execute:  sortRolesCommand,
		trigger:  "sortroles",
		desc:     "Sorts all spoiler roles alphabetically between dummy optin roles.",
		elevated: true,
	})
}
package commands

import (
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"

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

	if m.Author.ID != s.State.User.ID {
		misc.MapMutex.Lock()
	}
	guildBotLog := misc.GuildMap[m.GuildID].GuildConfig.BotLog.ID
	if len(misc.GuildMap[m.GuildID].SpoilerMap) == 0 {
		if m.Author.ID != s.State.User.ID {
			misc.MapMutex.Unlock()
		}
		return
	}
	if m.Author.ID != s.State.User.ID {
		misc.MapMutex.Unlock()
	}

	// Confirms whether optins exist
	if m.Author.ID == s.State.User.ID {
		misc.MapMutex.Unlock()
	}
	err := misc.OptInsHandler(s, m.ChannelID, m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}
	if m.Author.ID == s.State.User.ID {
		misc.MapMutex.Lock()
	}

	// Fetches info from the server and puts it in debPre
	debPre, err := s.GuildRoles(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Refreshes the positions of all roles in the server (because when created roles start at 0)
	for i := 0; i < len(debPre); i++ {
		spoilerRoles = append(spoilerRoles, debPre[i])
	}

	// Pushes the refreshed positions to the server
	_, err = s.GuildRoleReorder(m.GuildID, spoilerRoles)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	time.Sleep(time.Millisecond * 333)

	// Resets the value of spoilerRoles
	spoilerRoles = nil

	// Fetches the refreshed info from the server and puts it in deb
	deb, err := s.GuildRoles(m.GuildID)
	if err != nil {
		misc.CommandErrorHandler(s, m, err, guildBotLog)
		return
	}

	// Saves the original opt-in-above position
	if m.Author.ID != s.State.User.ID {
		misc.MapMutex.Lock()
	}
	for i := 0; i < len(deb); i++ {
		if deb[i].ID == misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.ID {
			misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position = deb[i].Position
		}
	}

	// Adds all spoiler roles in SpoilerMap in the spoilerRoles slice
	// Adds all non-spoiler roles under opt-in-above (including it) in the underSpoilerRoles slice
	for i := 0; i < len(deb); i++ {
		_, ok := misc.GuildMap[m.GuildID].SpoilerMap[deb[i].ID]
		if ok {
			spoilerRoles = append(spoilerRoles, misc.GuildMap[m.GuildID].SpoilerMap[deb[i].ID])
			if deb[i].Position < misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position {
				controlNum++
			}
		} else if !ok &&
			deb[i].Position <= misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position &&
			deb[i].ID != m.GuildID {
			underSpoilerRoles = append(underSpoilerRoles, deb[i])
		}
	}
	if m.Author.ID != s.State.User.ID {
		misc.MapMutex.Unlock()
	}

	// If there are spoiler roles under opt-in-above it goes in to move and sort
	if controlNum > 0 {

		// Sorts the spoilerRoles slice (all spoiler roles) alphabetically
		sort.Sort(misc.SortRoleByAlphabet(spoilerRoles))

		// Moves the sorted spoiler roles above opt-in-above
		if m.Author.ID != s.State.User.ID {
			misc.MapMutex.Lock()
		}
		for i := 0; i < len(spoilerRoles); i++ {
			spoilerRoles[i].Position = misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position
		}
		if m.Author.ID != s.State.User.ID {
			misc.MapMutex.Unlock()
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
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}

		time.Sleep(time.Millisecond * 333)

		// Fetches info from the server and puts it in debPost
		debPost, err := s.GuildRoles(m.GuildID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}

		// Refreshes deb
		deb, err = s.GuildRoles(m.GuildID)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}

		// Saves the new opt-in-above position
		if m.Author.ID != s.State.User.ID {
			misc.MapMutex.Lock()
		}
		for i := 0; i < len(debPost); i++ {
			if deb[i].ID == misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.ID {
				misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position = deb[i].Position
			}
		}

		for i := 0; i < len(spoilerRoles); i++ {
			spoilerRoles[i].Position = misc.GuildMap[m.GuildID].GuildConfig.OptInAbove.Position + len(spoilerRoles) - i
			misc.GuildMap[m.GuildID].SpoilerMap[spoilerRoles[i].ID].Position = spoilerRoles[i].Position
		}
		if m.Author.ID != s.State.User.ID {
			misc.MapMutex.Unlock()
		}

		// Pushes the sorted list to the server
		_, err = s.GuildRoleReorder(m.GuildID, spoilerRoles)
		if err != nil {
			misc.CommandErrorHandler(s, m, err, guildBotLog)
			return
		}

		if m.Author.ID == s.State.User.ID {
			return
		}

		_, err = s.ChannelMessageSend(m.ChannelID, "Roles sorted.")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
			if err != nil {
				return
			}
			return
		}
	} else {
		_, err = s.ChannelMessageSend(m.ChannelID, "Error: Spoiler roles already sorted or the spoiler roles are above the opt-in-above (in which case please move them manually.)")
		if err != nil {
			_, err = s.ChannelMessageSend(guildBotLog, err.Error()+"\n"+misc.ErrorLocation(err))
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
		desc:     "Sorts all spoiler roles alphabetically between dummy opt-in roles",
		elevated: true,
		category: "misc",
	})
}

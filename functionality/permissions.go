package functionality

import (
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/db"
)

const (
	User Permission = iota
	Mod
	Admin
	Owner
)

type Permission int

// Checks if a user has admin permissions or is a privileged role
func HasElevatedPermissions(s *discordgo.Session, userID string, guildID string) bool {
	mem, err := s.State.Member(guildID, userID)
	if err != nil {
		mem, err = s.GuildMember(guildID, userID)
		if err != nil {
			return false
		}
	}
	mem.GuildID = guildID

	if isAdmin, _ := MemberIsAdmin(s, guildID, mem, discordgo.PermissionAdministrator); isAdmin {
		return true
	}

	guild, err := s.State.Guild(guildID)
	if err != nil {
		guild, err = s.Guild(guildID)
		if err != nil {
			return false
		}
	}
	if userID == guild.OwnerID {
		return true
	}

	return HasPrivilegedPermissions(mem)
}

// Checks if member has admin permissions
func MemberIsAdmin(s *discordgo.Session, guildID string, mem *discordgo.Member, permission int) (bool, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		guild, err = s.Guild(guildID)
		if err != nil {
			return false, nil
		}
	}
	if mem.User.ID == guild.OwnerID {
		return true, nil
	}

	// Iterate through the role IDs stored in member.Roles
	// to check permissions
	for _, roleID := range mem.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			roles, err := s.GuildRoles(guildID)
			if err != nil {
				return false, err
			}
			for _, guildRole := range roles {
				if guildRole.ID == roleID {
					role = guildRole
					break
				}
			}
		}

		if role == nil {
			return false, nil
		}
		if role.Permissions&permission != 0 {
			return true, nil
		}
	}

	return false, nil
}

// Checks if a user has a privileged role in a given server
func HasPrivilegedPermissions(m *discordgo.Member) bool {
	if m.User.ID == config.OwnerID {
		return true
	}

	guildSettings := db.GetGuildSettings(m.GuildID)

	for _, privilegedRole := range guildSettings.GetCommandRoles() {
		for _, role := range m.Roles {
			if role == privilegedRole.GetID() {
				return true
			}
		}
	}

	return false
}

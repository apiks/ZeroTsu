package functionality

import (
	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

const (
	User permission = iota
	Mod
	Admin
	Owner
)

type permission int

// Checks if a user has admin permissions or is a privileged role
func HasElevatedPermissions(s discordgo.Session, userID string, guildID string) bool {
	mem, err := s.State.Member(guildID, userID)
	if err != nil {
		mem, err = s.GuildMember(guildID, userID)
		if err != nil {
			return false
		}
	}

	if isAdmin, _ := MemberIsAdmin(&s, guildID, mem, discordgo.PermissionAdministrator); isAdmin {
		return true
	}

	return HasPrivilegedPermissions(mem, guildID)
}

// Checks if member has admin permissions
func MemberIsAdmin(s *discordgo.Session, guildID string, mem *discordgo.Member, permission int) (bool, error) {
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
func HasPrivilegedPermissions(m *discordgo.Member, guildID string) bool {
	Mutex.RLock()
	defer Mutex.RUnlock()
	if m.User.ID == config.OwnerID {
		return true
	}

	for _, role := range m.Roles {
		for _, privilegedRole := range GuildMap[guildID].GuildConfig.CommandRoles {
			if role == privilegedRole.ID {
				return true
			}
		}
	}
	return false
}

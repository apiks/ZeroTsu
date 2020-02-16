package functionality

import (
	"strconv"
	"strings"

	"github.com/mediocregopher/radix/v3"

	"github.com/r-anime/ZeroTsu/config"
)

// GetRedisGuildSettings fetches a guild's existing settings from redis and returns them
func GetRedisGuildSettings(guildID string) (*GuildSettings, error) {
	var (
		guildSettings = &GuildSettings{}

		base  map[string]string

		botLog  *Cha
		optInUnder  *Role
		optInAbove  *Role
		mutedRole  *Role
		commandRoles []*Role
		voiceChannels []*VoiceCha

		query strings.Builder
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)

	redis, err := radix.NewPool("tcp", config.RedisAddress, 10)
	if err != nil {
		return nil, err
	}

	// Base
	err = redis.Do(radix.Cmd(&base, "HGETALL", query.String()))
	if err != nil {
		redis.Close()
		return nil, err
	}

	// Botlog
	botLog, err = GetBotLog(guildID)
	if err != nil {
		redis.Close()
		return nil, err
	}

	// OptInUnder
	optInAbove, err = GetOptInUnder(guildID)
	if err != nil {
		redis.Close()
		return nil, err
	}

	// OptInAbove
	optInAbove, err = GetOptInAbove(guildID)
	if err != nil {
		redis.Close()
		return nil, err
	}

	// MutedRole
	mutedRole, err = GetMutedRole(guildID)
	if err != nil {
		redis.Close()
		return nil, err
	}

	// CommandRoles
	commandRoleIds, err := GetCommandRoleIds(guildID)
	if err != nil {
		redis.Close()
		return nil, err
	}
	for _, id := range commandRoleIds {
		commandRole, err := GetCommandRole(guildID, id)
		if err != nil {
			redis.Close()
			return nil, err
		}

		commandRoles = append(commandRoles, commandRole)
	}

	// VoiceChannels
	voiceChannelIds, err := GetVoiceChannelIds(guildID)
	if err != nil {
		redis.Close()
		return nil, err
	}
	for _, id := range voiceChannelIds {
		voiceChannel, err  := GetVoiceChannel(guildID, id)
		if err != nil {
			redis.Close()
			return nil, err
		}

		voiceChannels = append(voiceChannels, voiceChannel)
	}

	redis.Close()

	guildSettings.Prefix = base["prefix"]
	guildSettings.PingMessage = base["pingMessage"]
	guildSettings.ModOnly, _ = strconv.ParseBool(base["modOnly"])
	guildSettings.VoteModule, _ = strconv.ParseBool(base["voteModule"])
	guildSettings.WaifuModule, _ = strconv.ParseBool(base["waifuModule"])
	guildSettings.WhitelistFileFilter, _ = strconv.ParseBool(base["whitelistFileFilter"])
	guildSettings.ReactsModule, _ = strconv.ParseBool(base["reactsModule"])
	guildSettings.Premium, _ = strconv.ParseBool(base["premium"])
	guildSettings.BotLog = botLog
	guildSettings.OptInUnder = optInUnder
	guildSettings.OptInAbove = optInAbove
	guildSettings.MutedRole = mutedRole
	guildSettings.CommandRoles = commandRoles
	guildSettings.VoiceChas = voiceChannels

	return guildSettings, err
}

// SetGuildSettings sets a guild's settings in the Redis instance
func SetGuildSettings(guildID string, settings *GuildSettings) error {
	redis, err := radix.NewPool("tcp", config.RedisAddress, 10)
	if err != nil {
		return err
	}
	defer redis.Close()

	// Base
	err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID,
		"prefix", settings.Prefix,
		"pingMessage", settings.PingMessage,
		"modOnly", strconv.FormatBool(settings.ModOnly),
		"voteModule", strconv.FormatBool(settings.VoteModule),
		"waifuModule", strconv.FormatBool(settings.WaifuModule),
		"whitelistFileFilter", strconv.FormatBool(settings.WhitelistFileFilter),
		"reactsModule", strconv.FormatBool(settings.ReactsModule),
		"premium", strconv.FormatBool(settings.Premium)))
	if err != nil {
		return err
	}

	// BotLog
	if settings.BotLog != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":botLog"))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":botLog",
			"id", settings.BotLog.ID,
			"name", settings.BotLog.Name))
		if err != nil {
			return err
		}
	}

	// OptInUnder
	if settings.OptInUnder != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":optInUnder"))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":optInUnder",
			"id", settings.OptInUnder.ID,
			"name", settings.OptInUnder.Name,
			"position", strconv.Itoa(settings.OptInUnder.Position)))
		if err != nil {
			return err
		}
	}

	// OptInAbove
	if settings.OptInAbove != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":optInAbove"))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":optInAbove",
			"id", settings.OptInAbove.ID,
			"name", settings.OptInAbove.Name,
			"position", strconv.Itoa(settings.OptInAbove.Position)))
		if err != nil {
			return err
		}
	}

	// MutedRole
	if settings.MutedRole != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":mutedRole"))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":mutedRole",
			"id", settings.MutedRole.ID,
			"name", settings.MutedRole.Name))
		if err != nil {
			return err
		}
	}

	// CommandRoles
	if settings.CommandRoles != nil {
		for _, commandRole := range settings.CommandRoles {
			err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":commandRoles:"+commandRole.ID))
			if err != nil {
				return err
			}

			err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":commandRoles:"+commandRole.ID,
				"id", commandRole.ID,
				"name", commandRole.Name))
			if err != nil {
				return err
			}
		}
	}

	// VoiceChas
	if settings.VoiceChas != nil {
		for _, voiceCha := range settings.VoiceChas {
			err := redis.Do(radix.Cmd(nil, "DEL", "guild:"+guildID+":voiceChannels:"+voiceCha.ID))
			if err != nil {
				return err
			}

			err = redis.Do(radix.Cmd(nil, "HSET", "guild:"+guildID+":voiceChannels:"+voiceCha.ID,
				"id", voiceCha.ID,
				"name", voiceCha.Name))
			if err != nil {
				return err
			}

			// VoiceRoles for VoiceCha
			for _, voiceRole := range voiceCha.Roles {
				err = redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":voiceChannels:"+voiceCha.ID+":roles:"+voiceRole.ID))
				if err != nil {
					return err
				}
				err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":voiceChannels:"+voiceCha.ID+":roles:"+voiceRole.ID,
					"id", voiceRole.ID,
					"name", voiceRole.Name))
				if err != nil {
					return err
				}
			}
		}
	}

	// VoteChannelCategory
	if settings.VoteChannelCategory != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":voteChannelCategory"))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":voteChannelCategory",
			"id", settings.VoteChannelCategory.ID,
			"name", settings.VoteChannelCategory.Name))
		if err != nil {
			return err
		}
	}

	return nil
}

// GetBotLog returns a target guild's botLog channel object
func GetBotLog(guildID string) (*Cha, error) {
	var (
		channel  = &Cha{}
		query strings.Builder
		m     map[string]string
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":botLog")

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	err = redis.Do(radix.Cmd(&m, "HGETALL", query.String()))
	if err != nil {
		return nil, err
	}

	channel.ID = m["id"]
	channel.Name = m["name"]

	return channel, nil
}

// GetOptInUnder returns a target guild's optInUnder role object
func GetOptInUnder(guildID string) (*Role, error) {
	var (
		role  = &Role{}
		query strings.Builder
		m     map[string]string
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":optInUnder")

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	err = redis.Do(radix.Cmd(&m, "HGETALL", query.String()))
	if err != nil {
		return nil, err
	}

	role.ID = m["id"]
	role.Name = m["name"]
	role.Position, _ = strconv.Atoi(m["position"])

	return role, nil
}

// GetOptInAbove returns a target guild's optInAbove role object
func GetOptInAbove(guildID string) (*Role, error) {
	var (
		role  = &Role{}
		query strings.Builder
		m     map[string]string
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":optInAbove")

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	err = redis.Do(radix.Cmd(&m, "HGETALL", query.String()))
	if err != nil {
		return nil, err
	}

	role.ID = m["id"]
	role.Name = m["name"]
	role.Position, _ = strconv.Atoi(m["position"])

	return role, nil
}

// GetMutedRole returns a target guild's muted role object
func GetMutedRole(guildID string) (*Role, error) {
	var (
		role  = &Role{}
		query strings.Builder
		m     map[string]string
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":mutedRole")

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	err = redis.Do(radix.Cmd(&m, "HGETALL", query.String()))
	if err != nil {
		return nil, err
	}

	role.ID = m["id"]
	role.Name = m["name"]

	return role, nil
}

// GetCommandRole returns a target guild's command role object
func GetCommandRole(guildID string, roleID string) (*Role, error) {
	var (
		role  = &Role{}
		query strings.Builder
		m     map[string]string
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":commandRoles:")
	query.WriteString(roleID)

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	err = redis.Do(radix.Cmd(&m, "HGETALL", query.String()))
	if err != nil {
		return nil, err
	}

	role.ID = m["id"]
	role.Name = m["name"]

	return role, nil
}

// GetCommandRoleIds returns the ids of a guild's command role ids
func GetCommandRoleIds(guildID string) ([]string, error) {
	var (
		id string
		ids []string
		query strings.Builder
		prefix strings.Builder
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":commandRoles:*")

	prefix.WriteString("guilds:")
	prefix.WriteString(guildID)
	prefix.WriteString(":commandRoles:")

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	s := radix.NewScanner(redis, radix.ScanOpts{Command: "SCAN", Pattern: query.String()})
	defer s.Close()

	for s.Next(&id) {
		ids = append(ids, strings.TrimPrefix(id, prefix.String()))
	}

	return ids, nil
}

// GetVoiceChannelIds returns the ids of a guild's voice channel ids
func GetVoiceChannelIds(guildID string) ([]string, error) {
	var (
		id string
		ids []string
		query strings.Builder
		prefix strings.Builder
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":voiceChannels:*")

	prefix.WriteString("guilds:")
	prefix.WriteString(guildID)
	prefix.WriteString(":voiceChannels:")

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	s := radix.NewScanner(redis, radix.ScanOpts{Command: "SCAN", Pattern: query.String()})
	defer s.Close()

	for s.Next(&id) {
		ids = append(ids, strings.TrimPrefix(id, prefix.String()))
	}

	return ids, nil
}

// GetVoiceChannel returns a target guild's voice channel object
func GetVoiceChannel(guildID string, channelID string) (*VoiceCha, error) {
	var (
		voiceChannel = &VoiceCha{}
		query strings.Builder
		m  map[string]string
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":voiceChannels:")
	query.WriteString(channelID)

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	err = redis.Do(radix.Cmd(&m, "HGETALL", query.String()))
	if err != nil {
		return nil, err
	}

	voiceChannel.ID = m["id"]
	voiceChannel.Name = m["name"]

	roleIds, err := GetVoiceChannelRoleIds(guildID, channelID)
	if err != nil {
		return nil, err
	}

	for _, roleId := range roleIds {
		role, err := GetVoiceChannelRole(guildID, channelID, roleId)
		if err != nil {
			return nil, err
		}

		voiceChannel.Roles = append(voiceChannel.Roles, role)
	}

	return voiceChannel, nil
}

// GetVoiceChannelRole returns a target guild's voice channel role object
func GetVoiceChannelRole(guildID string, channelID string, roleID string) (*Role, error) {
	var (
		role  = &Role{}
		query strings.Builder
		m     map[string]string
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":voiceChannels:")
	query.WriteString(channelID)
	query.WriteString(":roles:")
	query.WriteString(roleID)

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	err = redis.Do(radix.Cmd(&m, "HGETALL", query.String()))
	if err != nil {
		return nil, err
	}

	role.ID = m["id"]
	role.Name = m["name"]

	return role, nil
}


// GetVoiceChannelRoleIds returns the ids of a guild's voice channel role ids
func GetVoiceChannelRoleIds(guildID string, channelID string) ([]string, error) {
	var (
		id string
		ids []string
		query strings.Builder
		prefix strings.Builder
	)

	query.WriteString("guilds:")
	query.WriteString(guildID)
	query.WriteString(":voiceChannels:")
	query.WriteString(channelID)
	query.WriteString(":roles:*")

	prefix.WriteString("guilds:")
	prefix.WriteString(guildID)
	prefix.WriteString(":voiceChannels:")
	prefix.WriteString(channelID)
	prefix.WriteString(":roles:")

	redis, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		return nil, err
	}
	defer redis.Close()

	s := radix.NewScanner(redis, radix.ScanOpts{Command: "SCAN", Pattern: query.String()})
	defer s.Close()

	for s.Next(&id) {
		ids = append(ids, strings.TrimPrefix(id, prefix.String()))
	}

	return ids, nil
}

package functionality

import (
	"github.com/mediocregopher/radix/v3"
	"log"
	"strconv"

	"ZeroTsu/config"
)

// TransferGuildToRedis transfers a guild's json database values to the Redis instance
func TransferGuildToRedis(guildID string) error {
	redis, err := radix.NewPool("tcp", config.RedisIP+":"+config.RedisPort, 10)
	if err != nil {
		return err
	}
	defer redis.Close()

	// Sets the base guild object
	err = redis.Do(radix.Cmd(nil, "DEL", "guild:"+guildID))
	if err != nil {
		return err
	}

	log.Println("Starting transfer to redis instance. . .")

	Mutex.RLock()
	err = redis.Do(radix.Cmd(nil, "HSET", "guild:"+guildID,
		"id", guildID,
		"prefix", GuildMap[guildID].GuildConfig.Prefix,
		"pingMessage", GuildMap[guildID].GuildConfig.PingMessage,
		"modOnly", strconv.FormatBool(GuildMap[guildID].GuildConfig.ModOnly),
		"voteModule", strconv.FormatBool(GuildMap[guildID].GuildConfig.VoteModule),
		"waifuModule", strconv.FormatBool(GuildMap[guildID].GuildConfig.WaifuModule),
		"whitelistFileFilter", strconv.FormatBool(GuildMap[guildID].GuildConfig.WhitelistFileFilter),
		"reactsModule", strconv.FormatBool(GuildMap[guildID].GuildConfig.ReactsModule),
		"premium", strconv.FormatBool(GuildMap[guildID].GuildConfig.Premium)))
	if err != nil {
		Mutex.RUnlock()
		return err
	}
	Mutex.RUnlock()

	err = transferGuildSettings(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildPunishedUsersToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildFiltersToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildSpoilerRolesToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildMessageRequirementsToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildFeedsToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildFeedCheksToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildRafflesToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildWaifusToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildWaifuTradesToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildMemberInfoToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildSpoilerMapToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildEmojiStatsToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildChannelStatsToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildUserChangeStatsToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildVerifiedStatsToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildVoteInfoToRedis(redis, guildID)
	if err != nil {
		return err
	}
	err = transferGuildTempChaToRedis(redis, guildID)
	if err != nil {
		return err
	}

	log.Println("Transfer to redis instance has been completed!")

	return nil
}

// transferGuildSettingsToRedis transfers a guild's settings to the Redis instance
func transferGuildSettings(redis *radix.Pool, guildID string) error {
	Mutex.RLock()
	defer Mutex.RUnlock()

	// BotLog
	if GuildMap[guildID].GuildConfig.BotLog != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":botLog"))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":botLog",
			"botLogName", GuildMap[guildID].GuildConfig.BotLog.Name,
			"botLogID", GuildMap[guildID].GuildConfig.BotLog.ID))
		if err != nil {
			return err
		}
	}

	// OptInUnder
	if GuildMap[guildID].GuildConfig.OptInUnder != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":optInUnder"))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":optInUnder",
			"optInUnderName", GuildMap[guildID].GuildConfig.OptInUnder.Name,
			"optInUnderID", GuildMap[guildID].GuildConfig.OptInUnder.ID,
			"optInUnderPosition", strconv.Itoa(GuildMap[guildID].GuildConfig.OptInUnder.Position)))
		if err != nil {
			return err
		}
	}

	// OptInAbove
	if GuildMap[guildID].GuildConfig.OptInAbove != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":optInAbove"))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":optInAbove",
			"optInAboveName", GuildMap[guildID].GuildConfig.OptInAbove.Name,
			"optInAboveID", GuildMap[guildID].GuildConfig.OptInAbove.ID,
			"optInAbovePosition", strconv.Itoa(GuildMap[guildID].GuildConfig.OptInAbove.Position)))
		if err != nil {
			return err
		}
	}

	// MutedRole
	if GuildMap[guildID].GuildConfig.MutedRole != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":mutedRole:"+GuildMap[guildID].GuildConfig.MutedRole.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":mutedRole:"+GuildMap[guildID].GuildConfig.MutedRole.ID,
			"id", GuildMap[guildID].GuildConfig.MutedRole.ID,
			"name", GuildMap[guildID].GuildConfig.MutedRole.Name,
			"position", strconv.Itoa(GuildMap[guildID].GuildConfig.MutedRole.Position)))
		if err != nil {
			return err
		}
	}

	// CommandRoles
	if GuildMap[guildID].GuildConfig.CommandRoles != nil {
		for _, commandRole := range GuildMap[guildID].GuildConfig.CommandRoles {
			err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":commandRoles:"+commandRole.ID))
			if err != nil {
				return err
			}

			err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":commandRoles:"+commandRole.ID,
				"id", commandRole.ID,
				"name", commandRole.Name,
				"position", strconv.Itoa(commandRole.Position)))
			if err != nil {
				return err
			}
		}
	}

	// VoiceChas
	if GuildMap[guildID].GuildConfig.VoiceChas != nil {
		for _, voiceCha := range GuildMap[guildID].GuildConfig.VoiceChas {
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
					"name", voiceRole.Name,
					"position", strconv.Itoa(voiceRole.Position)))
				if err != nil {
					return err
				}
			}
		}
	}

	// VoteChannelCategory
	if GuildMap[guildID].GuildConfig.VoteChannelCategory != nil {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":voteChannelCategory:"+GuildMap[guildID].GuildConfig.VoteChannelCategory.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":voteChannelCategory:"+GuildMap[guildID].GuildConfig.VoteChannelCategory.ID,
			"id", GuildMap[guildID].GuildConfig.VoteChannelCategory.ID,
			"name", GuildMap[guildID].GuildConfig.VoteChannelCategory.Name))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildPunishedUsersToRedis transfers a guild's punished users to the Redis instance
func transferGuildPunishedUsersToRedis(redis *radix.Pool, guildID string) error {
	Mutex.RLock()
	defer Mutex.RUnlock()

	for _, punishedUser := range GuildMap[guildID].PunishedUsers {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":punishedUsers:"+punishedUser.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":punishedUsers:"+punishedUser.ID,
			"id", punishedUser.ID,
			"username", punishedUser.User,
			"unbanDate", punishedUser.UnbanDate.String(),
			"unmuteDate", punishedUser.UnmuteDate.String()))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildFiltersToRedis transfers a guild's filters to the Redis instance
func transferGuildFiltersToRedis(redis *radix.Pool, guildID string) error {
	Mutex.RLock()
	defer Mutex.RUnlock()

	err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":filters"))
	if err != nil {
		return err
	}
	for _, filter := range GuildMap[guildID].Filters {
		err = redis.Do(radix.Cmd(nil, "SADD", "guilds:"+guildID+":filters", filter.Filter))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildSpoilerRolesToRedis transfers a guild's spoiler roles to the Redis instance
func transferGuildSpoilerRolesToRedis(redis *radix.Pool, guildID string) error {
	Mutex.RLock()
	defer Mutex.RUnlock()

	for _, role := range GuildMap[guildID].SpoilerRoles {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":spoilerRoles:"+role.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":spoilerRoles:"+role.ID,
			"id", role.ID,
			"name", role.Name,
			"color", strconv.Itoa(role.Color),
			"hoist", strconv.FormatBool(role.Hoist),
			"managed", strconv.FormatBool(role.Managed),
			"mentionable", strconv.FormatBool(role.Mentionable),
			"permissions", strconv.Itoa(role.Permissions),
			"position", strconv.Itoa(role.Position)))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildMessageRequirementsToRedis transfers a guild's message requirements to the Redis instance
func transferGuildMessageRequirementsToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, messReq := range GuildMap[guildID].MessageRequirements {
		if messReq.ID == "" {
			id, err := GenerateID(guildID)
			if err != nil {
				log.Println(err)
				continue
			}
			messReq.ID = id
		}

		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":messageRequirements:"+messReq.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":messageRequirements:"+messReq.ID,
			"id", messReq.ID,
			"phrase", messReq.Phrase,
			"channelID", messReq.Channel,
			"lastUserID", messReq.LastUserID,
			"type", messReq.Type))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildFeedsToRedis transfers a guild's reddit feeds to the Redis instance
func transferGuildFeedsToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, feed := range GuildMap[guildID].Feeds {
		if feed.ID == "" {
			id, err := GenerateID(guildID)
			if err != nil {
				log.Println(err)
				continue
			}
			feed.ID = id
		}

		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":feeds:"+feed.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":feeds:"+feed.ID,
			"id", feed.ID,
			"title", feed.Title,
			"subreddit", feed.Subreddit,
			"channelID", feed.ChannelID,
			"author", feed.Author,
			"type", feed.PostType,
			"pin", strconv.FormatBool(feed.Pin)))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildFeedCheksToRedis transfers a guild's reddit feed checks to the Redis instance
func transferGuildFeedCheksToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, feedCheck := range GuildMap[guildID].FeedChecks {
		if feedCheck.ID == "" {
			id, err := GenerateID(guildID)
			if err != nil {
				log.Println(err)
				continue
			}
			feedCheck.ID = id
		}

		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":feedChecks:"+feedCheck.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":feedChecks:"+feedCheck.ID,
			"id", feedCheck.ID,
			"guid", feedCheck.GUID,
			"date", feedCheck.Date.String(),
			"title", feedCheck.Thread.Title,
			"subreddit", feedCheck.Thread.Subreddit,
			"channelID", feedCheck.Thread.ChannelID,
			"author", feedCheck.Thread.Author,
			"type", feedCheck.Thread.PostType,
			"pin", strconv.FormatBool(feedCheck.Thread.Pin)))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildRafflesToRedis transfers a guild's raffles to the Redis instance
func transferGuildRafflesToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, raffle := range GuildMap[guildID].Raffles {
		if raffle.ID == "" {
			id, err := GenerateID(guildID)
			if err != nil {
				log.Println(err)
				continue
			}
			raffle.ID = id
		}

		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":raffles:"+raffle.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":raffles:"+raffle.ID,
			"id", raffle.ID,
			"name", raffle.Name,
			"reactMessageID", raffle.ReactMessageID))
		if err != nil {
			return err
		}

		err = redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":raffles:"+raffle.ID+":participants"))
		if err != nil {
			return err
		}
		for _, participantID := range raffle.ParticipantIDs {
			err = redis.Do(radix.Cmd(nil, "SADD", "guilds:"+guildID+":raffles:"+raffle.ID+":participants", participantID))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// transferGuildWaifusToRedis transfers a guild's waifus to the Redis instance
func transferGuildWaifusToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, waifu := range GuildMap[guildID].Waifus {
		if waifu.ID == "" {
			id, err := GenerateID(guildID)
			if err != nil {
				log.Println(err)
				continue
			}
			waifu.ID = id
		}

		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":waifus:"+waifu.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":waifus:"+waifu.ID,
			"id", waifu.ID,
			"name", waifu.Name))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildWaifuTradesToRedis transfers a guild's waifu trades to the Redis instance
func transferGuildWaifuTradesToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, trade := range GuildMap[guildID].WaifuTrades {
		if trade.ID == "" {
			id, err := GenerateID(guildID)
			if err != nil {
				log.Println(err)
				continue
			}
			trade.ID = id
		}

		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":waifuTrades:"+trade.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":waifuTrades:"+trade.ID,
			"id", trade.ID,
			"tradeID", trade.TradeID,
			"initiatorID", trade.InitiatorID,
			"accepteeID", trade.AccepteeID))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildMemberInfoToRedis transfers a guild's member info to the Redis instance
func transferGuildMemberInfoToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, member := range GuildMap[guildID].MemberInfoMap {

		// Base User
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":members:"+member.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":members:"+member.ID,
			"id", member.ID,
			"discrim", member.Discrim,
			"username", member.Username,
			"nickname", member.Nickname,
			"joinDate", member.JoinDate,
			"redditUsername", member.RedditUsername,
			"verifiedDate", member.VerifiedDate,
			"unmuteDate", member.UnmuteDate,
			"unbanDate", member.UnbanDate,
			"suspectedSpambot", strconv.FormatBool(member.SuspectedSpambot)))
		if err != nil {
			return err
		}

		// Past Usernames
		if member.PastUsernames != nil {
			err = redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":members:"+member.ID+":pastUsernames"))
			if err != nil {
				return err
			}
			for _, username := range member.PastUsernames {
				err = redis.Do(radix.Cmd(nil, "SADD", "guilds:"+guildID+":members:"+member.ID+":pastUsernames", username))
				if err != nil {
					return err
				}
			}
		}

		// Past Nicknames
		if member.PastNicknames != nil {
			err = redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":members:"+member.ID+":pastNicknames"))
			if err != nil {
				return err
			}
			for _, nickname := range member.PastNicknames {
				err = redis.Do(radix.Cmd(nil, "SADD", "guilds:"+guildID+":members:"+member.ID+":pastNicknames", nickname))
				if err != nil {
					return err
				}
			}
		}

		// Warnings
		if member.Warnings != nil {
			err = redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":members:"+member.ID+":warnings"))
			if err != nil {
				return err
			}
			for _, warning := range member.Warnings {
				err = redis.Do(radix.Cmd(nil, "SADD", "guilds:"+guildID+":members:"+member.ID+":warnings", warning))
				if err != nil {
					return err
				}
			}
		}

		// Mutes
		if member.Mutes != nil {
			err = redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":members:"+member.ID+":mutes"))
			if err != nil {
				return err
			}
			for _, mute := range member.Mutes {
				err = redis.Do(radix.Cmd(nil, "SADD", "guilds:"+guildID+":members:"+member.ID+":mutes", mute))
				if err != nil {
					return err
				}
			}
		}

		// Kicks
		if member.Kicks != nil {
			err = redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":members:"+member.ID+":kicks"))
			if err != nil {
				return err
			}
			for _, kick := range member.Kicks {
				err = redis.Do(radix.Cmd(nil, "SADD", "guilds:"+guildID+":members:"+member.ID+":kicks", kick))
				if err != nil {
					return err
				}
			}
		}

		// Bans
		if member.Bans != nil {
			err = redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":members:"+member.ID+":bans"))
			if err != nil {
				return err
			}
			for _, ban := range member.Bans {
				err = redis.Do(radix.Cmd(nil, "SADD", "guilds:"+guildID+":members:"+member.ID+":bans", ban))
				if err != nil {
					return err
				}
			}
		}

		// Timestamps
		if member.Timestamps != nil {
			for _, timestamp := range member.Timestamps {
				if timestamp.ID == "" {
					id, err := GenerateID(guildID)
					if err != nil {
						log.Println(err)
						continue
					}
					timestamp.ID = id
				}

				err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":members:"+member.ID+":timestamps:"+timestamp.ID))
				if err != nil {
					return err
				}
				err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":members:"+member.ID+":timestamps:"+timestamp.ID,
					"id", timestamp.ID,
					"type", timestamp.Type,
					"punishment", timestamp.Punishment,
					"timestamp", timestamp.Timestamp.String()))
				if err != nil {
					return err
				}
			}
		}

		// Waifu
		if member.Waifu != nil {
			err = redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":members:"+member.ID+":waifu"))
			if err != nil {
				return err
			}
			err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":members:"+member.ID+":waifu",
				"id", member.Waifu.ID,
				"name", member.Waifu.Name))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// transferGuildSpoilerMapToRedis transfers a guild's spoiler map to the Redis instance
func transferGuildSpoilerMapToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, spoilerRole := range GuildMap[guildID].SpoilerMap {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":spoilerMap:"+spoilerRole.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":spoilerMap:"+spoilerRole.ID,
			"id", spoilerRole.ID,
			"name", spoilerRole.Name,
			"position", strconv.Itoa(spoilerRole.Position),
			"color", strconv.Itoa(spoilerRole.Color),
			"permissions", strconv.Itoa(spoilerRole.Permissions),
			"mentionable", strconv.FormatBool(spoilerRole.Mentionable),
			"hoist", strconv.FormatBool(spoilerRole.Hoist),
			"managed", strconv.FormatBool(spoilerRole.Managed)))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildEmojiStatsToRedis transfers a guild's emoji stats to the Redis instance
func transferGuildEmojiStatsToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, emoji := range GuildMap[guildID].EmojiStats {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":emojiStats:"+emoji.ID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":emojiStats:"+emoji.ID,
			"id", emoji.ID,
			"name", emoji.Name,
			"messageUsage", strconv.Itoa(emoji.MessageUsage),
			"uniqueMessageUsage", strconv.Itoa(emoji.UniqueMessageUsage),
			"reactions", strconv.Itoa(emoji.Reactions)))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildChannelStatsToRedis transfers a guild's channel stats to the Redis instance
func transferGuildChannelStatsToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for _, channel := range GuildMap[guildID].ChannelStats {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":channelStats:"+channel.ChannelID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":channelStats:"+channel.ChannelID,
			"id", channel.ChannelID,
			"name", channel.Name,
			"exists", strconv.FormatBool(channel.Exists),
			"optin", strconv.FormatBool(channel.Optin)))
		if err != nil {
			return err
		}

		// Messages per date
		for date, messages := range channel.Messages {
			err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":channelStats:"+channel.ChannelID+":messages:"+date))
			if err != nil {
				return err
			}
			err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":channelStats:"+channel.ChannelID+":messages:"+date,
				"messages", strconv.Itoa(messages)))
			if err != nil {
				return err
			}
		}

		// Users per date
		for date, users := range channel.RoleCount {
			err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":channelStats:"+channel.ChannelID+":roleCount:"+date))
			if err != nil {
				return err
			}
			err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":channelStats:"+channel.ChannelID+":roleCount:"+date,
				"roleCount", strconv.Itoa(users)))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// transferGuildUserChangeStatsToRedis transfers a guild's user change stats to the Redis instance
func transferGuildUserChangeStatsToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for date, userChange := range GuildMap[guildID].UserChangeStats {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":userChangeStats:"+date))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":userChangeStats:"+date,
			"change", strconv.Itoa(userChange)))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildVerifiedStatsToRedis transfers a guild's verified stats to the Redis instance
func transferGuildVerifiedStatsToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for date, verified := range GuildMap[guildID].VerifiedStats {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":verifiedStats:"+date))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":verifiedStats:"+date,
			"verified", strconv.Itoa(verified)))
		if err != nil {
			return err
		}
	}

	return nil
}

// transferGuildVoteInfoToRedis transfers a guild's vote info to the Redis instance
func transferGuildVoteInfoToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for messageID, vote := range GuildMap[guildID].VoteInfoMap {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":voteInfo:"+messageID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":voteInfo:"+messageID,
			"votesRequired", strconv.Itoa(vote.VotesReq),
			"channelName", vote.Channel,
			"date", vote.Date.String(),
			"description", vote.Description,
			"type", vote.ChannelType,
			"categoryID", vote.Category))

		if vote.MessageReact != nil {
			err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":voteInfo:"+messageID+":messageReact"))
			if err != nil {
				return err
			}
			err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":voteInfo:"+messageID+":messageReact",
				"id", vote.MessageReact.ID,
				"channelID", vote.MessageReact.ChannelID))
			if err != nil {
				return err
			}
		}

		if vote.User != nil {
			err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":voteInfo:"+messageID+":author"))
			if err != nil {
				return err
			}
			err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":voteInfo:"+messageID+":author",
				"id", vote.User.ID,
				"username", vote.User.Username,
				"discrim", vote.User.Discriminator))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// transferGuildTempChaToRedis transfers a guild's temp channels to the Redis instance
func transferGuildTempChaToRedis(redis *radix.Pool, guildID string) error {
	Mutex.Lock()
	defer Mutex.Unlock()

	for roleID, tempCha := range GuildMap[guildID].TempChaMap {
		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":tempChas:"+roleID))
		if err != nil {
			return err
		}
		err = redis.Do(radix.Cmd(nil, "HSET", "guilds:"+guildID+":tempChas:"+roleID,
			"creationDate", tempCha.CreationDate.String(),
			"roleName", tempCha.RoleName,
			"elevated", strconv.FormatBool(tempCha.Elevated)))
	}

	return nil
}

//// transferGuildReactJoinsToRedis transfers a guild's react join autoroles to the Redis instance
//func transferGuildReactJoinsToRedis(redis *radix.Pool, guildID string) error {
//	Mutex.Lock()
//	defer Mutex.Unlock()
//
//	for messageID, reactMap := range GuildMap[guildID].ReactJoinMap {
//		err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":reactJoins:"+messageID))
//		if err != nil {
//			return err
//		}
//
//		for _, t := range reactMap {
//			err := redis.Do(radix.Cmd(nil, "DEL", "guilds:"+guildID+":reactJoins:"+messageID+":"+t))
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

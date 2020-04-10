package commands

import (
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"log"
)

func CleanGuild() {
	for guildID, _ := range entities.Guilds.DB {
		memberInfo := db.GetGuildMemberInfo(guildID)
		for id, member := range memberInfo {
			if member.ID == "" ||
				member.Discrim == "" ||
				member.Username == "" {
				delete(memberInfo, id)
				continue
			}
			if member.Warnings != nil ||
				member.Mutes != nil ||
				member.Kicks != nil ||
				member.Bans != nil ||
				member.VerifiedDate != "" ||
				member.UnbanDate != "" ||
				member.SuspectedSpambot ||
				member.Timestamps != nil ||
				member.UnmuteDate != "" ||
				member.Waifu.Name != "" {
				continue
			}
			if member.PastNicknames != nil && len(member.PastNicknames) > 3 ||
				member.PastUsernames != nil && len(member.PastUsernames) > 3 {
				continue
			}
			delete(memberInfo, id)
			continue
		}
		db.SetGuildMemberInfo(guildID, memberInfo)
	}

	log.Println("CLEANED GUILDS")
}
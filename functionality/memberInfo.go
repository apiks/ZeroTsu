package functionality

import (
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"
	"time"
)

// Initializes a user in memberInfo if he doesn't exist there
func InitializeUser(u *discordgo.User, guildID string) {
	// Stores time of joining
	var joinDate strings.Builder
	t := time.Now()
	z, _ := t.Zone()
	joinDate.WriteString(t.Format("2006-01-02 15:04:05"))
	joinDate.WriteString(" ")
	joinDate.WriteString(z)

	// Creates User object and sets him in db
	user := entities.NewUserInfo(u.ID, u.Discriminator, u.Username, "",
		nil, nil, nil, nil, nil,
		nil, joinDate.String(), "", "", "",
		"", nil, entities.Waifu{}, false)

	db.SetGuildMember(guildID, user)
}

package misc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

// OnUserUpdate listens for user updates to compare usernames and discrim
func OnUserUpdate(s *discordgo.Session, e *discordgo.PresenceUpdate) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string) + "\n" + ErrorLocation(rec.(error)))
			if err != nil {
				fmt.Println(rec)
			}
		}
	}()

	s.RWMutex.RLock()
	userMember := e.User
	s.RWMutex.RUnlock()

	presenceCounter++
	fmt.Println(fmt.Sprintf("in presence update %v", presenceCounter))

	MapMutex.Lock()
	if len(MemberInfoMap) == 0 {
		MapMutex.Unlock()
		return
	}

	// Fetches user from memberInfo if possible
	user, ok := MemberInfoMap[userMember.ID]
	if !ok {
		MapMutex.Unlock()
		return
	}
	MapMutex.Unlock()

	// Checks usernames and updates if needed
	if user.Username != userMember.Username && userMember.Username != "" {
		flag := true
		lower := strings.ToLower(userMember.Username)

		for _, names := range user.PastUsernames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			user.PastUsernames = append(user.PastUsernames, userMember.Username)
			user.Username = userMember.Username
		}
	}

	// Checks if the discrim in database is the same as the discrim used by the user. If not it changes it
	if user.Discrim != userMember.Discriminator {
		user.Discrim = userMember.Discriminator
	}

	// Saves the updates to memberInfoMap and writes to disk
	MapMutex.Lock()
	MemberInfoMap[userMember.ID] = user
	MemberInfoWrite(MemberInfoMap)
	MapMutex.Unlock()
}
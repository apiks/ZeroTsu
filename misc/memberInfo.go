package misc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
)

var (
	MapMutex sync.Mutex
	Key      = []byte("VfBhgLzmD4QH3W94pjgdbH8Tyv2HPRzq")
)

// UserInfo is the in memory storage of each user's information
type UserInfo struct {
	ID               string       `json:"id"`
	Discrim          string       `json:"discrim"`
	Username         string       `json:"username"`
	Nickname         string       `json:"nickname,omitempty"`
	PastUsernames    []string     `json:"pastUsernames,omitempty"`
	PastNicknames    []string     `json:"pastNicknames,omitempty"`
	Warnings         []string     `json:"warnings,omitempty"`
	Mutes			 []string	  `json:"mutes,omitempty"`
	Kicks            []string     `json:"kicks,omitempty"`
	Bans             []string     `json:"bans,omitempty"`
	JoinDate         string       `json:"joinDate"`
	RedditUsername   string       `json:"redditUser,omitempty"`
	VerifiedDate     string       `json:"verifiedDate,omitempty"`
	UnmuteDate		 string		  `json:"unmuteDate,omitempty"`
	UnbanDate        string       `json:"unbanDate,omitempty"`
	Timestamps       []*Punishment `json:"timestamps,omitempty"`
	Waifu            *Waifu        `json:"waifu,omitempty"`
	SuspectedSpambot bool
}

// Creates a struct type in which we'll hold every banned user
type PunishedUsers struct {
	ID        	string    	`json:"id"`
	User      	string    	`json:"user"`
	UnbanDate 	time.Time 	`json:"unbanDate"`
	UnmuteDate 	time.Time 	`json:"unmuteDate"`
}

// Struct where we'll hold punishment timestamps
type Punishment struct {
	Punishment string    `json:"punishment"`
	Type       string    `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
}

// Initializes user in memberInfo if he doesn't exist there
func InitializeUser(u *discordgo.Member, guildID string) {

	// Stores time of joining
	t := time.Now()
	z, _ := t.Zone()
	join := t.Format("2006-01-02 15:04:05") + " " + z

	GuildMap[guildID].MemberInfoMap[u.User.ID] = &UserInfo{
		ID:       u.User.ID,
		Discrim:  u.User.Discriminator,
		Username: u.User.Username,
		JoinDate: join,
	}
}

// Checks if user exists in memberInfo on joining server and adds him if he doesn't
// Also updates usernames and/or nicknames
// Also updates discriminator
// Also verifies them if they're already verified in memberinfo
func OnMemberJoinGuild(s *discordgo.Session, e *discordgo.GuildMemberAdd) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMemberJoinGuild")
		}
	}()

	var (
		flag        bool
		writeFlag   bool
	)

	if e.GuildID == "" {
		return
	}

	// Pulls info on user if possible
	user, err := s.GuildMember(e.GuildID, e.User.ID)
	if err != nil {
		return
	}

	MapMutex.Lock()
	defer MapMutex.Unlock()
	if _, ok := GuildMap[e.GuildID]; !ok {
		InitDB(s, e.GuildID)
		LoadGuilds()
	}

	// If memberInfo is empty, it initializes
	if len(GuildMap[e.GuildID].MemberInfoMap) == 0 {

		// Initializes the first user of memberInfo
		InitializeUser(user, e.GuildID)

		flag = true
		writeFlag = true

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		// Sends verification message to user in DMs if possible
		if config.Website != "" && e.GuildID == "267799767843602452" {
			dm, _ := s.UserChannelCreate(user.User.ID)
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n"+
				"Please verify your reddit account at http://%v/verification?reqvalue=%v", config.Website, ciphertext))
		}

	} else {
		// Checks if user exists in memberInfo.json. If yes it changes flag to true
		if _, ok := GuildMap[e.GuildID].MemberInfoMap[user.User.ID]; ok {
			flag = true
		}
	}

	// If user still doesn't exist after check above, it initializes user
	if !flag {

		// Initializes the new user
		InitializeUser(user, e.GuildID)
		writeFlag = true

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		// Sends verification message to user in DMs if possible
		if config.Website != "" && e.GuildID == "267799767843602452" {
			dm, _ := s.UserChannelCreate(user.User.ID)
			_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n"+
				"Please verify your reddit account at http://%v/verification?reqvalue=%v", config.Website, ciphertext))
		}
	}

	// Fetches user from memberInfo if possible
	memberInfoUser, ok := GuildMap[e.GuildID].MemberInfoMap[user.User.ID]
	if !ok {
		return
	}

	// If user is already in memberInfo lacks reddit username and site is enabled, tell user to verify
	if memberInfoUser.RedditUsername == "" && config.Website != "" && e.GuildID == "267799767843602452" {

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		// Sends verification message to user in DMs if possible
		dm, _ := s.UserChannelCreate(user.User.ID)
		_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n"+
			"Please verify your reddit account at http://%v/verification?reqvalue=%v", config.Website, ciphertext))
	}

	// Checks if the user's current username is the same as the one in the database. Otherwise updates
	if user.User.Username != memberInfoUser.Username && user.User.Username != "" {
		flagTwo := true
		lower := strings.ToLower(user.User.Username)

		for _, names := range memberInfoUser.PastUsernames {
			if strings.ToLower(names) == lower {
				flagTwo = false
				break
			}
		}

		if flagTwo {
			memberInfoUser.PastUsernames = append(memberInfoUser.PastUsernames, user.User.Username)
		}
		memberInfoUser.Username = user.User.Username
		writeFlag = true
	}

	// Checks if the user's current nickname is the same as the one in the database. Otherwise updates
	if memberInfoUser.Nickname != user.Nick && user.Nick != "" {
		flagTwo := true
		lower := strings.ToLower(user.Nick)

		for _, names := range memberInfoUser.PastNicknames {
			if strings.ToLower(names) == lower {
				flagTwo = false
				break
			}
		}

		if flagTwo {
			memberInfoUser.PastNicknames = append(memberInfoUser.PastNicknames, user.Nick)
		}
		memberInfoUser.Nickname = user.Nick
		writeFlag = true
	}

	// Checks if the discrim in database is the same as the discrim used by the user. If not it changes it
	if user.User.Discriminator != memberInfoUser.Discrim && user.User.Discriminator != "" {
		memberInfoUser.Discrim = user.User.Discriminator
		writeFlag = true
	}

	// Saves the updates to memberInfoMap and writes to disk if need be
	if !writeFlag {
		return
	}

	WriteMemberInfo(GuildMap[e.GuildID].MemberInfoMap, e.GuildID)
}

// OnMemberUpdate listens for member updates to compare usernames, nicknames and discrim
func OnMemberUpdate(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnMemberUpdate")
		}
	}()

	if e.GuildID == "" {
		return
	}

	var writeFlag bool

	MapMutex.Lock()
	defer MapMutex.Unlock()
	if _, ok := GuildMap[e.GuildID]; !ok {
		InitDB(s, e.GuildID)
		LoadGuilds()
	}

	if len(GuildMap[e.GuildID].MemberInfoMap) == 0 {
		return
	}

	// Fetches user from memberInfo if possible
	memberInfoUser, ok := GuildMap[e.GuildID].MemberInfoMap[e.User.ID]
	if !ok {
		return
	}

	// Checks usernames and updates if needed
	if memberInfoUser.Username != e.User.Username && e.User.Username != "" {
		flag := true
		lower := strings.ToLower(e.User.Username)

		for _, names := range memberInfoUser.PastUsernames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			memberInfoUser.PastUsernames = append(memberInfoUser.PastUsernames, e.User.Username)
		}
		memberInfoUser.Username = e.User.Username
		writeFlag = true
	}

	// Checks nicknames and updates if needed
	if memberInfoUser.Nickname != e.Nick && e.Nick != "" {
		flag := true
		lower := strings.ToLower(e.Nick)

		for _, names := range memberInfoUser.PastNicknames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			memberInfoUser.PastNicknames = append(memberInfoUser.PastNicknames, e.Nick)
		}
		memberInfoUser.Nickname = e.Nick
		writeFlag = true
	}

	// Checks if the discrim in database is the same as the discrim used by the memberInfoUser. If not it changes it
	if memberInfoUser.Discrim != e.User.Discriminator && e.User.Discriminator != "" {
		memberInfoUser.Discrim = e.User.Discriminator
		writeFlag = true
	}

	// Checks if username or discrim were changed, else do NOT write to disk
	if !writeFlag {
		return
	}

	// Saves the updates to memberInfoMap and writes to disk
	WriteMemberInfo(GuildMap[e.GuildID].MemberInfoMap, e.GuildID)
}

// OnPresenceUpdate listens for user updates to compare usernames and discrim
func OnPresenceUpdate(s *discordgo.Session, e *discordgo.PresenceUpdate) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			log.Println("Recovery in OnPresenceUpdate")
		}
	}()

	if e.GuildID == "" {
		return
	}

	var writeFlag bool

	MapMutex.Lock()
	defer MapMutex.Unlock()
	if _, ok := GuildMap[e.GuildID]; !ok {
		InitDB(s, e.GuildID)
		LoadGuilds()
	}

	if len(GuildMap[e.GuildID].MemberInfoMap) == 0 {
		return
	}

	// Fetches user from memberInfo if possible
	memberInfoUser, ok := GuildMap[e.GuildID].MemberInfoMap[e.User.ID]
	if !ok {
		return
	}

	// Checks usernames and updates if needed
	if memberInfoUser.Username != e.User.Username && e.User.Username != "" {
		flag := true
		lower := strings.ToLower(e.User.Username)

		for _, names := range memberInfoUser.PastUsernames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			memberInfoUser.PastUsernames = append(memberInfoUser.PastUsernames, e.User.Username)
		}
		memberInfoUser.Username = e.User.Username
		writeFlag = true
	}

	// Checks nicknames and updates if needed
	if memberInfoUser.Nickname != e.Nick && e.Nick != "" {
		flag := true
		lower := strings.ToLower(e.Nick)

		for _, names := range memberInfoUser.PastNicknames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			memberInfoUser.PastNicknames = append(memberInfoUser.PastNicknames, e.Nick)
		}
		memberInfoUser.Nickname = e.Nick
		writeFlag = true
	}

	// Checks if the discrim in database is the same as the discrim used by the memberInfoUser. If not it changes it
	if memberInfoUser.Discrim != e.User.Discriminator && e.User.Discriminator != "" {
		memberInfoUser.Discrim = e.User.Discriminator
		writeFlag = true
	}

	// Checks if username or discrim were changed, else do NOT write to disk
	if !writeFlag {
		return
	}

	// Saves the updates to memberInfoMap and writes to disk
	WriteMemberInfo(GuildMap[e.GuildID].MemberInfoMap, e.GuildID)
}

// Encrypt string to base64 crypto using AES
func Encrypt(key []byte, text string) string {
	// key := []byte(keyText)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return ""
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Println(err)
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// Decrypt from base64 to decrypted string
func Decrypt(key []byte, cryptoText string) (string, bool) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println(err)
		return "", false
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		log.Println("ciphertext too short")
		return "", false
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext), true
}

// Cleans up duplicate nicknames and usernames in memberInfo.json
func DuplicateUsernamesAndNicknamesCleanup() {
	path := "database/guilds"
	folders, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err)
		return
	}
	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		MapMutex.Lock()
		DuplicateRecursion(f.Name())
		WriteMemberInfo(GuildMap[f.Name()].MemberInfoMap, f.Name())
		MapMutex.Unlock()
	}

	log.Println("FINISHED WITH DUPLICATES")
}

// Helper of above
func DuplicateRecursion(guildID string) {
	for _, value := range GuildMap[guildID].MemberInfoMap {
		// Remove duplicate usernames
		for index, username := range value.PastUsernames {
			for indexDuplicate, usernameDuplicate := range value.PastUsernames {
				if index != indexDuplicate && username == usernameDuplicate {
					value.PastUsernames = append(value.PastUsernames[:indexDuplicate], value.PastUsernames[indexDuplicate+1:]...)
					DuplicateRecursion(guildID)
					return
				}
			}
		}
		// Remove duplicate nicknames
		for index, nickname := range value.PastNicknames {
			for indexDuplicate, nicknameDuplicate := range value.PastNicknames {
				if index != indexDuplicate && nickname == nicknameDuplicate {
					value.PastNicknames = append(value.PastNicknames[:indexDuplicate], value.PastNicknames[indexDuplicate+1:]...)
					DuplicateRecursion(guildID)
					return
				}
			}
		}
	}
}

// Updates user usernames to the current ones they're using in memberInfo.json
func UsernameCleanup(s *discordgo.Session, e *discordgo.Ready) {
	var progress int
	MapMutex.Lock()
	defer MapMutex.Unlock()
	for _, guild := range e.Guilds {
		for _, mapUser := range GuildMap[guild.ID].MemberInfoMap {
			user, err := s.User(mapUser.ID)
			if err != nil {
				progress++
				continue
			}
			if mapUser.Username != user.Username {
				mapUser.Username = user.Username
			}
			if mapUser.Discrim != user.Discriminator {
				mapUser.Discrim = user.Discriminator
			}
			progress++
			log.Printf("%v out of %v \n", progress, len(GuildMap[guild.ID].MemberInfoMap))
		}

		path := "database/guilds"
		folders, err := ioutil.ReadDir(path)
		if err != nil {
			log.Println(err)
			return
		}
		for _, f := range folders {
			if !f.IsDir() {
				continue
			}
			WriteMemberInfo(GuildMap[guild.ID].MemberInfoMap, f.Name())
		}
	}

	fmt.Println("FINISHED WITH USERNAMES")
}

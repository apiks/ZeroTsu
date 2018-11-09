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

var (
	MemberInfoMap    = make(map[string]*UserInfo)
	BannedUsersSlice []BannedUsers
	MapMutex         sync.Mutex
	Key              = []byte("VfBhgLzmD4QH3W94pjgdbH8Tyv2HPRzq")
)

// UserInfo is the in memory storage of each user's information
type UserInfo struct {
	ID             string   `json:"id"`
	Discrim        string   `json:"discrim"`
	Username       string   `json:"username"`
	Nickname       string   `json:"nickname,omitempty"`
	PastUsernames  []string `json:"pastUsernames,omitempty"`
	PastNicknames  []string `json:"pastNicknames,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	Kicks          []string `json:"kicks,omitempty"`
	Bans           []string `json:"bans,omitempty"`
	JoinDate       string   `json:"joinDate"`
	RedditUsername string   `json:"redditUser,omitempty"`
	VerifiedDate   string   `json:"verifiedDate,omitempty"`
	UnbanDate      string   `json:"unbanDate,omitempty"`
}

// Creates a struct type in which we'll hold every banned user
type BannedUsers struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	UnbanDate time.Time `json:"unbanDate"`
}

// Reads member info from memberInfo.json
func MemberInfoRead() {

	// Reads all the member users from the memberInfo.json file and puts them in memberInfoByte as bytes
	memberInfoByte, err := ioutil.ReadFile("database/memberInfo.json")
	if err != nil {
		return
	}

	// Takes all the users from memberInfo.json from byte and puts them into the UserInfo map
	MapMutex.Lock()
	err = json.Unmarshal(memberInfoByte, &MemberInfoMap)
	if err != nil {
		MapMutex.Unlock()
		return
	}

	// Fixes empty IDs
	for ID, user := range MemberInfoMap {
		if user.ID == "" {
			user.ID = ID
		}
	}
	MapMutex.Unlock()
}

// Writes member info to memberInfo.json
func MemberInfoWrite(info map[string]*UserInfo) {

	// Turns info slice into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		return
	}

	// Writes to file
	err = ioutil.WriteFile("database/memberInfo.json", MarshaledStruct, 0644)
	if err != nil {
		return
	}
}

// Initializes user in memberInfo if he doesn't exist there
func InitializeUser(u *discordgo.Member) {

	var temp UserInfo

	// Sets ID, username and discriminator
	temp.ID = u.User.ID
	temp.Username = u.User.Username
	temp.Discrim = u.User.Discriminator

	// Stores time of joining
	t := time.Now()
	z, _ := t.Zone()
	join := t.Format("2006-01-02 15:04:05") + " " + z

	// Sets join date
	temp.JoinDate = join

	MemberInfoMap[u.User.ID] = &temp
}

// Checks if user exists in memberInfo on joining server and adds him if he doesn't
// Also updates usernames and/or nicknames
// Also updates discriminator
// Also verifies them if they're already verified in memberinfo
func OnMemberJoinGuild(s *discordgo.Session, e *discordgo.GuildMemberAdd) {

	var (
		flag        = false
		initialized = false
		userID		  string
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string) + "\n" + ErrorLocation(rec.(error)))
			if err != nil {
				fmt.Println(rec)
			}
		}
	}()

	// Pulls info on user if possible
	s.State.RWMutex.RLock()
	user, err := s.GuildMember(config.ServerID, e.User.ID)
	if err != nil {
		s.State.RWMutex.RUnlock()
		return
	}
	userID = user.User.ID
	s.State.RWMutex.RUnlock()

	// If memberInfo is empty, it initializes
	MapMutex.Lock()
	if len(MemberInfoMap) == 0 {

		// Initializes the first user of memberInfo.
		InitializeUser(user)

		flag = true
		initialized = true

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		// Sends verification message to user in DMs if possible
		dm, _ := s.UserChannelCreate(user.User.ID)
		_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n" +
			"Please verify your reddit account at http://%v/verification?reqvalue=%v", config.Website, ciphertext))

	} else {
		// Checks if user exists in memberInfo.json. If yes it changes flag to true
		for id := range MemberInfoMap {
			if MemberInfoMap[id].ID == user.User.ID {
				flag = true
				break
			}
		}
	}
	MapMutex.Unlock()

	// If user still doesn't exist after check above, it initializes user
	if !flag  {

		// Initializes the new user
		MapMutex.Lock()
		InitializeUser(user)
		MapMutex.Unlock()
		initialized = true

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		// Sends verification message to user in DMs if possible
		dm, _ := s.UserChannelCreate(user.User.ID)
		_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n" +
			"Please verify your reddit account at http://%v/verification?reqvalue=%v", config.Website, ciphertext))
	}

	// Fetches user from memberInfo
	MapMutex.Lock()
	existingUser, ok := MemberInfoMap[userID]
	if !ok {
		MapMutex.Unlock()
		return
	}

	// If user is already in memberInfo but hasn't verified before tell him to verify now
	if MemberInfoMap[userID].RedditUsername == "" && !initialized {

		// Encrypts id
		ciphertext := Encrypt(Key, userID)

		// Sends verification message to user in DMs if possible
		dm, _ := s.UserChannelCreate(userID)
		_, _ = s.ChannelMessageSend(dm.ID, fmt.Sprintf("You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n" +
			"Please verify your reddit account at http://%v/verification?reqvalue=%v", config.Website, ciphertext))
	}
	MapMutex.Unlock()

	// Checks if the user's current username is the same as the one in the database. Otherwise updates
	if user.User.Username != existingUser.Username {
		flag := true
		lower := strings.ToLower(user.User.Username)

		for _, names := range existingUser.PastUsernames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			existingUser.PastUsernames = append(existingUser.PastUsernames, user.User.Username)
			existingUser.Username = user.User.Username
		}
	}

	// Checks if the user's current nickname is the same as the one in the database. Otherwise updates
	if existingUser.Nickname != user.Nick && user.Nick != "" {
		flag := true
		lower := strings.ToLower(user.Nick)

		for _, names := range existingUser.PastNicknames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			existingUser.PastNicknames = append(existingUser.PastNicknames, user.Nick)
			existingUser.Nickname = user.Nick
		}
	}

	// Checks if the discrim in database is the same as the discrim used by the user. If not it changes it
	if user.User.Discriminator != existingUser.Discrim {
		existingUser.Discrim = user.User.Discriminator
	}

	MapMutex.Lock()
	// Saves the updates to memberInfoMap and writes to disk
	MemberInfoMap[userID] = existingUser
	MemberInfoWrite(MemberInfoMap)
	MapMutex.Unlock()
}

// OnMemberUpdate listens for member updates to compare nicks/usernames and discrim
func OnMemberUpdate(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			_, err := s.ChannelMessageSend(config.BotLogID, rec.(string) + "\n" + ErrorLocation(rec.(error)))
			if err != nil {
				fmt.Println(rec)
			}
		}
	}()

	s.State.RWMutex.RLock()
	userMember := e
	s.State.RWMutex.RUnlock()

	MapMutex.Lock()
	if len(MemberInfoMap) == 0 {
		MapMutex.Unlock()
		return
	}

	// Fetches user from memberInfo if possible
	user, ok := MemberInfoMap[userMember.User.ID]
	if !ok {
		MapMutex.Unlock()
		return
	}
	MapMutex.Unlock()


	// Checks usernames and updates if needed
	if user.Username != userMember.User.Username {
		flag := true
		lower := strings.ToLower(userMember.User.Username)

		for _, names := range user.PastUsernames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			user.PastUsernames = append(user.PastUsernames, userMember.User.Username)
			user.Username = userMember.User.Username
		}
	}

	// Checks nicknames and updates if needed
	if user.Nickname != userMember.Nick && userMember.Nick != "" {
		flag := true
		lower := strings.ToLower(userMember.Nick)

		for _, names := range user.PastNicknames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			user.PastNicknames = append(user.PastNicknames, userMember.Nick)
			user.Nickname = userMember.Nick
		}
	}

	// Checks if the discrim in database is the same as the discrim used by the user. If not it changes it
	if user.Discrim != userMember.User.Discriminator {
		user.Discrim = userMember.User.Discriminator
	}

	// Saves the updates to memberInfoMap and writes to disk
	MapMutex.Lock()
	MemberInfoMap[userMember.User.ID] = user
	MemberInfoWrite(MemberInfoMap)
	MapMutex.Unlock()
}

// Encrypt string to base64 crypto using AES
func Encrypt(key []byte, text string) string {
	// key := []byte(keyText)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// Decrypt from base64 to decrypted string
func Decrypt(key []byte, cryptoText string) string {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}

// Cleans up duplicate nicknames and usernames in memberInfo.json
func DuplicateUsernamesAndNicknamesCleanup() {
	MapMutex.Lock()
	DuplicateRecursion()
	MapMutex.Unlock()

	MemberInfoWrite(MemberInfoMap)

	fmt.Println("FINISHED WITH DUPLICATES")
}

// Helper of above
func DuplicateRecursion() {
	for _, value := range MemberInfoMap {
		// Remove duplicate usernames
		for index, username := range value.PastUsernames {
			for indexDuplicate, usernameDuplicate := range value.PastUsernames {
				if index != indexDuplicate && username == usernameDuplicate {
					value.PastUsernames = append(value.PastUsernames[:indexDuplicate], value.PastUsernames[indexDuplicate+1:]...)
					DuplicateRecursion()
					return
				}
			}
		}
		// Remove duplicate nicknames
		for index, nickname := range value.PastNicknames {
			for indexDuplicate, nicknameDuplicate := range value.PastNicknames {
				if index != indexDuplicate && nickname == nicknameDuplicate {
					value.PastNicknames = append(value.PastNicknames[:indexDuplicate], value.PastNicknames[indexDuplicate+1:]...)
					DuplicateRecursion()
					return
				}
			}
		}

	}
}
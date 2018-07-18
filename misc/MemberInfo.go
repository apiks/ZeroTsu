package misc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"
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

//Creates a struct type in which we'll hold every banned user
type BannedUsers struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	UnbanDate time.Time `json:"unbanDate"`
}

//Reads member info from memberInfo.json
func MemberInfoRead() {

	MapMutex.Lock()

	//Reads all the member users from the memberInfo.json file and puts them in MemberInfoMap as bytes
	memberInfoByte, err := ioutil.ReadFile("database/memberInfo.json")
	if err != nil {

		fmt.Println(err)
	}

	//Takes all the users from memberInfo.json from byte and puts them into the UserInfo map
	err = json.Unmarshal(memberInfoByte, &MemberInfoMap)
	if err != nil {

		fmt.Println(err)
	}

	MapMutex.Unlock()
}

//Writes member info to memberInfo.json
func MemberInfoWrite(info map[string]*UserInfo) {

	MapMutex.Lock()

	//Turns info slice into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/memberInfo.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}

	MapMutex.Unlock()
}

//Reads banned users info from bannedUsers.json
func BannedUsersRead() {

	//Reads all the banned users from the bannedUsers.json file and puts them in bannedusersByte as bytes
	bannedUsersByte, err := ioutil.ReadFile("database/bannedUsers.json")
	if err != nil {

		fmt.Println(err)
	}

	//Takes all the banned users from bannedUsers.json from byte and puts them into the BannedUsers struct slice
	err = json.Unmarshal(bannedUsersByte, &BannedUsersSlice)
	if err != nil {

		fmt.Println(err)
	}

}

//Writes banned users info to bannedUsers.json
func BannedUsersWrite(info []BannedUsers) {

	//Turns info into byte ready to be pushed to file
	MarshaledStruct, err := json.MarshalIndent(info, "", "    ")
	if err != nil {

		fmt.Println(err)
	}

	//Writes to file
	err = ioutil.WriteFile("database/bannedUsers.json", MarshaledStruct, 0644)
	if err != nil {

		fmt.Println(err)
	}
}

//Initializes user in memberInfo if he doesn't exist there
func InitializeUser(u *discordgo.Member) {

	var temp UserInfo

	MapMutex.Lock()

	//Sets ID, username and discriminator
	temp.ID = u.User.ID
	temp.Username = u.User.Username
	temp.Discrim = u.User.Discriminator

	//Stores time of joining
	t := time.Now()
	z, _ := t.Zone()
	join := t.Format("2006-01-02 15:04:05") + " " + z

	//Sets join date
	temp.JoinDate = join

	MemberInfoMap[u.User.ID] = &temp

	MapMutex.Unlock()
}

//Checks if user exists in memberInfo on joining server and adds him if he doesn't
//Also updates usernames and/or nicknames
//Also updates discriminator
func OnMemberJoinGuild(s *discordgo.Session, e *discordgo.GuildMemberAdd) {

	var (
		flag        = false
		initialized = false
	)

	//Pulls info on user
	user, err := s.State.Member(config.ServerID, e.User.ID)
	if err != nil {
		user, err = s.GuildMember(config.ServerID, e.User.ID)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	//Reads memberInfo.json
	MemberInfoRead()

	//If memberInfo is empty, it initializes
	if len(MemberInfoMap) == 0 {

		fmt.Println("Empty memberInfo. Initializing user.	")

		// Initializes the first user of memberInfo.
		InitializeUser(user)

		//Writes to memberInfo.json
		MemberInfoWrite(MemberInfoMap)

		// Set flags to true
		flag = true
		initialized = true

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		//Assigns success print string for user
		success := "You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n" +
			"Please verify your reddit account at http://localhost:3000//?reqvalue=" + ciphertext

		//Creates a DM connection and assigns it to dm
		dm, err := s.UserChannelCreate(user.User.ID)
		if err != nil {

			fmt.Println("Error: ", err)
		}

		//Sends a message to that DM connection
		_, err = s.ChannelMessageSend(dm.ID, success)
		if err != nil {

			fmt.Println("Error: ", err)
		}

	} else {

		//Checks if user exists in memberInfo.json. If yes it changes flag to true
		for id := range MemberInfoMap {
			if MemberInfoMap[id].ID == user.User.ID {
				flag = true
				break
			}
		}
	}

	//If user still doesn't exist after check above, it initializes user
	if flag == false {

		// Initializes the new user
		InitializeUser(user)

		// Set Initialized flag to true
		initialized = true

		//Writes to memberInfo.json
		MemberInfoWrite(MemberInfoMap)

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		//Assigns success print string for user
		success := "You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n" +
			"Please verify your reddit account at http://localhost:3000//?reqvalue=" + ciphertext

		//Creates a DM connection and assigns it to dm
		dm, err := s.UserChannelCreate(user.User.ID)
		if err != nil {

			fmt.Println("Error: ", err)
		}

		//Sends a message to that DM connection
		_, err = s.ChannelMessageSend(dm.ID, success)
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	//Writes to memberInfo.json
	MemberInfoWrite(MemberInfoMap)

	existingUser, ok := MemberInfoMap[e.User.ID]
	if !ok {
		fmt.Print("User: " + e.User.Username + " not found in memberInfo")
		return
	}

	if MemberInfoMap[e.User.ID].RedditUsername == "" && initialized == false {

		// Encrypts id
		ciphertext := Encrypt(Key, user.User.ID)

		//Assigns success print string for user
		success := "You have joined the /r/anime discord. We require a reddit account verification with an at least 1 week old account. \n" +
			"Please verify your reddit account at http://localhost:3000//?reqvalue=" + ciphertext

		//Creates a DM connection and assigns it to dm
		dm, err := s.UserChannelCreate(user.User.ID)
		if err != nil {

			fmt.Println("Error: ", err)
		}

		//Sends a message to that DM connection
		_, err = s.ChannelMessageSend(dm.ID, success)
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	//Checks if the user's current username is the same as the one in the database. Otherwise updates
	if user.User.Username != existingUser.Username {

		MapMutex.Lock()

		flag := true
		lower := strings.ToLower(e.User.Username)

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

		MapMutex.Unlock()

		//Writes to memberInfo.json
		MemberInfoWrite(MemberInfoMap)
	}

	//Checks if the user's current nickname is the same as the one in the database. Otherwise updates
	if existingUser.Nickname != user.Nick && user.Nick != "" {

		MapMutex.Lock()

		flag := true

		lower := strings.ToLower(user.Nick)

		for _, names := range existingUser.PastNicknames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}
		if flag {
			existingUser.PastNicknames = append(existingUser.PastNicknames, e.Nick)
			existingUser.Nickname = user.Nick
		}

		MapMutex.Unlock()

		//Writes to memberInfo.json
		MemberInfoWrite(MemberInfoMap)
	}

	//Checks if the discrim in database is the same as the discrim used by the user. If not it changes it
	if user.User.Discriminator != existingUser.Discrim {

		MapMutex.Lock()

		existingUser.Discrim = user.User.Discriminator

		MapMutex.Unlock()

		//Writes to memberInfo.json
		MemberInfoWrite(MemberInfoMap)

	}
}

// OnMemberUpdate listens for member updates to compare nicks/usernames and discrim
func OnMemberUpdate(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {

	//Reads memberInfo.json
	MemberInfoRead()

	if len(MemberInfoMap) == 0 {
		return
	}

	user, ok := MemberInfoMap[e.User.ID]
	if !ok {
		fmt.Print("User: " + e.User.Username + " not found in memberInfo")
		return
	}

	//Checks usernames and updates if needed
	if user.Username != e.User.Username {

		MapMutex.Lock()

		flag := true

		lower := strings.ToLower(e.User.Username)

		for _, names := range user.PastUsernames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			user.PastUsernames = append(user.PastUsernames, e.User.Username)
			user.Username = e.User.Username
		}

		MapMutex.Unlock()

		//Writes to memberInfo.json
		MemberInfoWrite(MemberInfoMap)
	}

	//Checks nicknames and updates if needed
	if user.Nickname != e.Nick && e.Nick != "" {

		MapMutex.Lock()

		flag := true

		lower := strings.ToLower(e.Nick)

		for _, names := range user.PastNicknames {
			if strings.ToLower(names) == lower {
				flag = false
				break
			}
		}

		if flag {
			user.PastNicknames = append(user.PastNicknames, e.Nick)
			user.Nickname = e.Nick
		}
		MapMutex.Unlock()

		//Writes to memberInfo.json
		MemberInfoWrite(MemberInfoMap)
	}

	//Checks if the discrim in database is the same as the discrim used by the user. If not it changes it
	if user.Discrim != e.User.Discriminator {

		MapMutex.Lock()

		user.Discrim = e.User.Discriminator

		MapMutex.Unlock()

		//Writes to memberInfo.json
		MemberInfoWrite(MemberInfoMap)

	}
}

// encrypt string to base64 crypto using AES
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

// decrypt from base64 to decrypted string
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

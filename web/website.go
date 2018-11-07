package web

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

var (
	Name          string
	DateUnix      float64
	UserCookieMap = make(map[string]*User)
)

type Access struct {
	RedditAccessToken string `json:"access_token"`
	TokenType         string `json:"token_type"`
	ExpiresIn         int    `json:"expires_in"`
	Scope             string `json:"scope"`
	RefreshToken      string `json:"refresh_token,omitempty"`
}

type User struct {
	Cookie                string    `json:"cookie"`
	Expiry                time.Time `json:"expiry"`
	RedditName            string    `json:"name"`
	AccCreation           float64   `json:"created_utc"`
	ID                    string    `json:"id"`
	UsernameDiscrim       string    `json:"usernamediscrim"`
	RedditVerifiedStatus  bool      `json:"redditverifiedstatus"`
	DiscordVerifiedStatus bool      `json:"redditverifiedstatus"`
	Error                 string    `json:"error"`
	Username              string    `json:"username"`
	Discriminator         string    `json:"discriminator"`
	AccOldEnough          bool      `json:"accoldenough"`
	Code                  string    `json:"code"`
	AltCheck              bool      `json:"altcheck"`
}

type Stats struct {
	Name string
	Dates []string
	Messages []int
	TotalMessages int
	DailyAverage int
}

type ChannelPick struct {
	ChannelStats map[string]*misc.Channel
	Flag bool
	Stats Stats
	Error bool
}

// Sorting by date. By Kagumi
type byDate []string

func (d byDate) Len() int {
	return len(d)
}

func (d byDate) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d byDate) Less(i, j int) bool {
	t1, _ := time.Parse(misc.DateFormat, d[i])
	t2, _ := time.Parse(misc.DateFormat, d[j])
	return t1.Before(t2)
}

// Generates a random string. By Kagumi
func randString(n int) (string, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func HomepageHandler(w http.ResponseWriter, r *http.Request) {
	// Loads the html & css homepage files
	t, err := template.ParseFiles("./web/assets/index.html")
	if err != nil {
		fmt.Print(err.Error())
	}
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func StatsPageHandler(w http.ResponseWriter, r *http.Request) {

	var (
		dateLabels []string
		messageCount []int
		stats Stats
		totalMessages int
		id string
		pick ChannelPick
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
		}
	}()

	// Fetches channel ID from url query
	queryValues := r.URL.Query()
	id = queryValues.Get("channelid")
	pick.Error = true

	// Checks for nil entry assignment error and saves from that (could be abused to stop bot)
	if id != "" {
		misc.MapMutex.Lock()
		if misc.ChannelStats[id] == nil {
			pick.Error = false
			// Loads the html & css stats files
			t, err := template.ParseFiles("./web/assets/channelstats.html")
			if err != nil {
				fmt.Print(err.Error())
				misc.MapMutex.Unlock()
				return
			}
			err = t.Execute(w, pick)
			if err != nil {
				fmt.Println(err.Error())
				misc.MapMutex.Unlock()
				return
			}
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
	}

	if id == "" {
		pick.ChannelStats = make(map[string]*misc.Channel)
		misc.MapMutex.Lock()
		pick.ChannelStats = misc.ChannelStats
		// Loads the html & css stats files
		t, err := template.ParseFiles("./web/assets/channelstats.html")
		if err != nil {
			fmt.Print(err.Error())
			misc.MapMutex.Unlock()
			return
		}
		err = t.Execute(w, pick)
		if err != nil {
			fmt.Println(err.Error())
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	} else {
		pick.Flag = true
	}

	// Save dates, sort them and then assign messages in order of the dates
	misc.MapMutex.Lock()
	for date := range misc.ChannelStats[id].Messages {
		dateLabels = append(dateLabels, date)
	}
	sort.Sort(byDate(dateLabels))
	for i := 0; i < len(dateLabels); i++ {
		messageCount = append(messageCount, misc.ChannelStats[id].Messages[dateLabels[i]])
		totalMessages += misc.ChannelStats[id].Messages[dateLabels[i]]
	}

	stats.Name = misc.ChannelStats["267799767843602452"].Name
	misc.MapMutex.Unlock()
	stats.Dates = dateLabels
	stats.Messages = messageCount
	stats.TotalMessages = totalMessages
	stats.DailyAverage = totalMessages / len(dateLabels)
	pick.Stats = stats


	// Loads the html & css stats files
	t, err := template.ParseFiles("./web/assets/channelstats.html")
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	err = t.Execute(w, pick)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Handles the verification
func VerificationHandler(w http.ResponseWriter, r *http.Request) {

	// Pulls cookie if it exists, else it creates a new one and assigns it
	cookieValue, err := r.Cookie("session-id")
	if err != nil {
		randNum, _ := randString(64)
		expire := time.Now().Add(10 * time.Minute)
		cookie := http.Cookie{Name: "session-id", Value: randNum, Expires: expire}
		http.SetCookie(w, &cookie)
		cookieValue = &cookie
	}

	var (
		errorVar string
		state    string
		code     string
		id       string
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
		}
	}()

	// Blurb fetches query from link
	queryValues := r.URL.Query()
	id = queryValues.Get("reqvalue")
	state = queryValues.Get("state")
	code = queryValues.Get("code")
	errorVar = queryValues.Get("error")

	// Saves the id in the user map if it exists
	if id != "" {
		var temp User

		// Decrypts encrypted id from url
		trueid := misc.Decrypt(misc.Key, id)

		misc.MapMutex.Lock()

		// Make it copy the current cookie map if it exists, otherwise make a new one
		if UserCookieMap[cookieValue.Value] != nil {
			temp = *UserCookieMap[cookieValue.Value]

			// If the user is verifying to another account with this cookie reset the old cookie values
			if temp.ID != trueid {

				temp.RedditVerifiedStatus = false
				temp.DiscordVerifiedStatus = false
				temp.RedditName = ""
				temp.AccOldEnough = false
				temp.UsernameDiscrim = ""
				temp.Username = ""
				temp.AltCheck = false
			}
		}

		temp.ID = trueid
		temp.Cookie = cookieValue.Value
		UserCookieMap[cookieValue.Value] = &temp
		misc.MapMutex.Unlock()
	}

	if code != "" {
		misc.MapMutex.Lock()
		if UserCookieMap[cookieValue.Value] != nil {
			// Sets code
			var temp User
			temp = *UserCookieMap[cookieValue.Value]
			temp.Code = code
			UserCookieMap[cookieValue.Value] = &temp
		}
		misc.MapMutex.Unlock()
	}

	if errorVar != "" {
		misc.MapMutex.Lock()
		if UserCookieMap[cookieValue.Value] != nil {
			// Sets error message
			var temp User
			temp = *UserCookieMap[cookieValue.Value]
			temp.Error = "Error: Permission not given in verification. If this was a mistake please try to verify again."
			UserCookieMap[cookieValue.Value] = &temp
		}
		misc.MapMutex.Unlock()
	}

	// Fetches user username and discriminator combo for showing in website. Also checks if user is verified already
	misc.MapMutex.Lock()
	if UserCookieMap[cookieValue.Value] != nil {
		var temp User
		temp = *UserCookieMap[cookieValue.Value]

		// Sets the username + discrim combo if it exists, also sorts out the verified status
		if misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID] != nil {

			username := misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID].Username + "#" + misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID].Discrim
			temp.UsernameDiscrim = username

			if misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID].RedditUsername != "" {
				temp.RedditVerifiedStatus = true
			}
		}

		UserCookieMap[cookieValue.Value] = &temp
	}

	if cookieValue != nil && errorVar == "" {
		if code == "" && id == "" && state == "" {

			// Sets default state
			var temp User
			temp.UsernameDiscrim = ""
			UserCookieMap[cookieValue.Value] = &temp

		} else if _, ok := UserCookieMap[cookieValue.Value]; ok {
			if state == "overlordconfirmsdiscord" && UserCookieMap[cookieValue.Value].Code != "" {

				uname, udiscrim, uid, err := getDiscordUsernameDiscrim(UserCookieMap[cookieValue.Value].Code)
				if err != nil {
					// Sets error message
					var temp User
					temp = *UserCookieMap[cookieValue.Value]
					temp.Error = "Error: User is not in memberInfo or cookie has expired. Please rejoin the server and try again."
					UserCookieMap[cookieValue.Value] = &temp

					// Loads the html & css verification files
					t, err := template.ParseFiles("web/assets/verification.html")
					if err != nil {
						fmt.Print(err.Error())
					}
					err = t.Execute(w, UserCookieMap[cookieValue.Value])
					if err != nil {
						fmt.Println(err.Error())
					}

					// Resets assigned Error Message
					if cookieValue != nil {
						var temp User
						temp = *UserCookieMap[cookieValue.Value]
						temp.Error = ""
						UserCookieMap[cookieValue.Value] = &temp
					}
					misc.MapMutex.Unlock()
					return
				}

				var temp User

				temp = *UserCookieMap[cookieValue.Value]
				temp.ID = uid
				temp.Username = uname
				temp.Discriminator = udiscrim
				temp.UsernameDiscrim = uname + "#" + udiscrim
				temp.DiscordVerifiedStatus = true
				UserCookieMap[cookieValue.Value] = &temp

				if UserCookieMap[cookieValue.Value].AccOldEnough && UserCookieMap[cookieValue.Value].ID != "" &&
					UserCookieMap[cookieValue.Value].RedditVerifiedStatus && UserCookieMap[cookieValue.Value].RedditName != "" {
					// Verifies user
					Verify(cookieValue, r)
				}

				if misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID] != nil {
					if misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID].RedditUsername != "" {

						var temp User
						temp = *UserCookieMap[cookieValue.Value]
						temp.RedditVerifiedStatus = true
						temp.DiscordVerifiedStatus = true
						UserCookieMap[cookieValue.Value] = &temp
					}
				} else {
					// Sets error message
					var temp User
					temp = *UserCookieMap[cookieValue.Value]
					temp.Error = "Error: User is not in memberInfo or cookie has expired. Please rejoin the server and try again."
					UserCookieMap[cookieValue.Value] = &temp
				}

			} else if state == "overlordconfirmsstring" && UserCookieMap[cookieValue.Value].Code != "" {

				// Fetches reddit username and checks whether account is at least 1 week old
				Name, DateUnix, err = getRedditUsername(UserCookieMap[cookieValue.Value].Code)
				if err != nil {
					// Sets error message
					var temp User
					temp = *UserCookieMap[cookieValue.Value]
					temp.Error = "Error: User is not in memberInfo or cookie has expired. Please rejoin the server and try again."
					UserCookieMap[cookieValue.Value] = &temp

					// Loads the html & css verification files
					t, err := template.ParseFiles("web/assets/verification.html")
					if err != nil {
						fmt.Print(err.Error())
					}
					err = t.Execute(w, UserCookieMap[cookieValue.Value])
					if err != nil {
						fmt.Println(err.Error())
					}

					// Resets assigned Error Message
					if cookieValue != nil {
						var temp User
						temp = *UserCookieMap[cookieValue.Value]
						temp.Error = ""
						UserCookieMap[cookieValue.Value] = &temp
					}
					misc.MapMutex.Unlock()
					return
				}

				epochT := time.Unix(int64(DateUnix), 0)
				prevWeek := time.Now().AddDate(0, 0, -7)
				accOldEnough := epochT.Before(prevWeek)

				// If account is old enough continue, else show error message
				if accOldEnough != true {

					// Sets error message
					var temp User
					temp = *UserCookieMap[cookieValue.Value]
					temp.Error = "Error: Reddit account is not old enough. Please try again once it is one week old."
					UserCookieMap[cookieValue.Value] = &temp

				} else if accOldEnough && UserCookieMap[cookieValue.Value].ID != "" &&
					UserCookieMap[cookieValue.Value].DiscordVerifiedStatus &&
					UserCookieMap[cookieValue.Value].RedditName == "" {

					// Saves the reddit username and acc age bool
					var temp User
					temp = *UserCookieMap[cookieValue.Value]
					temp.RedditName = Name
					temp.RedditVerifiedStatus = true
					temp.AccOldEnough = true
					UserCookieMap[cookieValue.Value] = &temp

					// Verifies user
					Verify(cookieValue, r)

				} else if accOldEnough == true && UserCookieMap[cookieValue.Value].RedditName == "" {

					// Saves the reddit username and acc age bool
					var temp User
					temp = *UserCookieMap[cookieValue.Value]
					temp.RedditName = Name
					temp.RedditVerifiedStatus = true
					temp.AccOldEnough = true
					UserCookieMap[cookieValue.Value] = &temp
				}
			}
		} else {
			// Sets error message
			var temp User
			temp.Error = "Error: Cookie has expired. Please try the bot link again."
			UserCookieMap[cookieValue.Value] = &temp
		}
	}
	misc.MapMutex.Unlock()

	// Loads the html & css verification files
	t, err := template.ParseFiles("web/assets/verification.html")
	if err != nil {
		fmt.Print(err.Error())
	}
	misc.MapMutex.Lock()
	err = t.Execute(w, UserCookieMap[cookieValue.Value])
	if err != nil {
		misc.MapMutex.Unlock()
		fmt.Println(err.Error())
	}
	misc.MapMutex.Unlock()

	// Resets assigned Error Message
	if cookieValue != nil {
		var temp User
		misc.MapMutex.Lock()
		temp = *UserCookieMap[cookieValue.Value]
		temp.Error = ""
		UserCookieMap[cookieValue.Value] = &temp
		misc.MapMutex.Unlock()
	}
}

// Verifies user on reddit and returns their reddit username
func getRedditUsername(code string) (string, float64, error) {

	// Initializes client
	client := &http.Client{Timeout: time.Second * 2}

	// Sets reddit required post info
	POSTinfo := "grant_type=authorization_code&code=" + code + fmt.Sprintf("&redirect_uri=http://%v/verification", config.Website)

	// Starts request to reddit
	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", bytes.NewBuffer([]byte(POSTinfo)))
	if err != nil {
		return "", 0, err
	}

	// Sets needed request parameters User Agent and Basic Auth
	req.Header.Set("User-Agent", misc.UserAgent)
	req.SetBasicAuth(config.RedditAppName, config.RedditAppSecret)
	resp, err := client.Do(req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	// Initializes Access type variable to hold data
	access := Access{}

	// Unmarshals json info into the above access variable to hold
	jsonErr := json.Unmarshal(body, &access)
	if jsonErr != nil {
		return "", 0, err
	}

	// Makes a GET request to reddit in reqAPI
	reqAPI, err := http.NewRequest("GET", "https://oauth.reddit.com/api/v1/me", nil)
	if err != nil {
		return "", 0, err
	}

	// Sets needed reqAPI paraemeters
	reqAPI.Header.Add("Authorization", "Bearer "+access.RedditAccessToken)
	reqAPI.Header.Add("User-Agent", misc.UserAgent)

	// Does the GET request and puts it into the respAPI
	respAPI, err := client.Do(reqAPI)
	if err != nil {
		return "", 0, err
	}

	// Reads the byte respAPI body into bodyAPI
	bodyAPI, err := ioutil.ReadAll(respAPI.Body)
	if err != nil {
		return "", 0, err
	}

	// Initializes user variable of type User to hold reddit json in
	user := User{}

	// Unmarshals all the required json fields in the above user variable
	jsonErr = json.Unmarshal(bodyAPI, &user)
	if jsonErr != nil {
		return "", 0, err
	}

	// Returns user reddit username and date of account creation in epoch time
	return user.RedditName, user.AccCreation, err
}

// Verifies user on discord and returns their discord username and discrim
func getDiscordUsernameDiscrim(code string) (string, string, string, error) {

	discordConf := oauth2.Config{
		ClientID:     config.BotID,
		ClientSecret: config.DiscordAppSecret,
		Scopes:       []string{"identity"},
		RedirectURL:  fmt.Sprintf("http://%v/verification", config.Website),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}

	token, err := discordConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return "", "", "", err
	}

	// Initializes client
	client := &http.Client{Timeout: time.Second * 2}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/@me", "https://discordapp.com/api"), nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	res, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}

	// Reads the byte respAPI body into bodyAPI
	bodyAPI, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", "", err
	}

	// Initializes user variable of type User to hold reddit json in
	user := User{}

	// Unmarshals all the required json fields in the above user variable
	jsonErr := json.Unmarshal(bodyAPI, &user)
	if jsonErr != nil {
		return "", "", "", err
	}

	return user.Username, user.Discriminator, user.ID, err
}

// Verifies user by assigning the necessary values
func Verify(cookieValue *http.Cookie, r *http.Request) {

	// Confirms that the map is not empty
	if len(misc.MemberInfoMap) == 0 {
		return
	}
	// Checks if cookie has expired while doing this
	if cookieValue == nil {
		return
	}

	//Stores time of verification
	t := time.Now()
	z, _ := t.Zone()
	join := t.Format("2006-01-02 15:04:05") + " " + z

	// Assigns needed values to temp
	var temp misc.UserInfo
	temp = *misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID]
	temp.RedditUsername = UserCookieMap[cookieValue.Value].RedditName
	temp.VerifiedDate = join
	misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID] = &temp

	// Writes the username to memberInfo.json
	misc.MemberInfoWrite(misc.MemberInfoMap)
}

// Checks if a user in the cookie map has the role and if they're verified it gives it to them, also deletes expired map fields
func VerifiedRoleAdd(s *discordgo.Session, e *discordgo.Ready) {

	var roleID string

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
		}
	}()

	// Checks every 2 seconds if a user in the UserCookieMap needs to be given the role
	for range time.NewTicker(3 * time.Second).C {

		misc.MapMutex.Lock()
		if len(UserCookieMap) != 0 {
			for key := range UserCookieMap {
				if UserCookieMap[key].RedditName != "" &&
					UserCookieMap[key].DiscordVerifiedStatus &&
					UserCookieMap[key].RedditVerifiedStatus {

					// Puts all server roles in roles variable
					roles, err := s.GuildRoles(config.ServerID)
					if err != nil {
						_, err := s.ChannelMessageSend(config.BotLogID, err.Error())
						if err != nil {
						}
					}

					// Fetches ID of Verified role
					for i := 0; i < len(roles); i++ {
						if roles[i].Name == "Verified" {
							roleID = roles[i].ID
						}
					}

					// Assigns role
					err = s.GuildMemberRoleAdd(config.ServerID, UserCookieMap[key].ID, roleID)
					if err != nil {
						_, err := s.ChannelMessageSend(config.BotLogID, err.Error())
						if err != nil {
						}
					}

					if !UserCookieMap[key].AltCheck {
						CheckAltAccount(s, UserCookieMap[key].ID)
						UserCookieMap[key].AltCheck = true
					}

					delete(UserCookieMap, key)
				}
			}
		}
		misc.MapMutex.Unlock()
	}
}

// Checks if a user is already verified when they join the server and if they are directly assigns them the verified role
func VerifiedAlready(s *discordgo.Session, u *discordgo.GuildMemberAdd) {

	var (
		roleID string
		userID string
	)

	// Pulls info on user if possible
	misc.MapMutex.Lock()
	user, err := s.GuildMember(config.ServerID, u.User.ID)
	if err != nil {
		return
	}
	userID = user.User.ID

	// Checks if the user is an already verified one
	if len(misc.MemberInfoMap) == 0 {
		misc.MapMutex.Unlock()
		return
	}
	if misc.MemberInfoMap[userID] == nil {
		misc.MapMutex.Unlock()
		return
	}
	if misc.MemberInfoMap[userID].RedditUsername == "" {
		misc.MapMutex.Unlock()
		return
	}
	misc.MapMutex.Unlock()

	// Puts all server roles in roles
	roles, err := s.GuildRoles(config.ServerID)
	if err != nil {
		_, err := s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {
			return
		}
		return
	}

	// Fetches ID of Verified role
	for i := 0; i < len(roles); i++ {
		if roles[i].Name == "Verified" {
			roleID = roles[i].ID
		}
	}

	// Assigns role
	err = s.GuildMemberRoleAdd(config.ServerID, userID, roleID)
	if err != nil {
		_, err := s.ChannelMessageSend(config.BotLogID, err.Error())
		if err != nil {
			misc.MapMutex.Unlock()
			return
		}
		misc.MapMutex.Unlock()
		return
	}

	CheckAltAccount(s, userID)
}

// Function that iterates through memberInfo.json and checks for any alt accounts for that ID. Verification version
func CheckAltAccount(s *discordgo.Session, id string) {

	var alts []string

	if len(misc.MemberInfoMap) == 0 {
		return
	}

	// Iterates through all users in memberInfo.json
	for _, userOne := range misc.MemberInfoMap {
		// Checks if the current user has the same reddit username as userCookieMap user
		if userOne.RedditUsername == misc.MemberInfoMap[id].RedditUsername {
			alts = append(alts, userOne.ID)
		}
	}

	// If there's more than one account with that reddit username print a message
	if len(alts) > 1 {
		success := "**Alternate Account Verified:** \n"
		for i := 0; i < len(alts); i++ {
			success = success + "<@" + alts[i] + "> \n"
		}
		// Prints the alts in bot-log channel
		_, _ = s.ChannelMessageSend(config.BotLogID, success)
	}
}
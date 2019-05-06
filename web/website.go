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
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

var (
	SafeCookieMap = SafeUserCookieMap{userCookieMap: make(map[string]*User)}

	// Map that keeps all user IDs that have successfuly verified but have not been given the role
	verifyMap     = make(map[string]string)
)

type Access struct {
	RedditAccessToken string `json:"access_token"`
	TokenType         string `json:"token_type"`
	ExpiresIn         int    `json:"expires_in"`
	Scope             string `json:"scope"`
	RefreshToken      string `json:"refresh_token,omitempty"`
}

type User struct {
	Cookie                	string    		`json:"cookie"`
	Expiry                	time.Time 		`json:"expiry"`
	RedditName            	string    		`json:"name"`
	AccCreation           	float64   		`json:"created_utc"`
	ID                    	string			`json:"id"`
	UsernameDiscrim       	string			`json:"usernamediscrim"`
	RedditVerifiedStatus	bool			`json:"redditverifiedstatus"`
	DiscordVerifiedStatus 	bool			`json:"redditverifiedstatus"`
	Error                 	string			`json:"error"`
	Username              	string			`json:"username"`
	Discriminator         	string			`json:"discriminator"`
	AccOldEnough          	bool			`json:"accoldenough"`
	Code                  	string			`json:"code"`
}

// Mutex safe userCookieMap
type SafeUserCookieMap struct {
	userCookieMap	map[string]*User
	mux				sync.Mutex
}

type UserBan struct {
	IsBanned	bool	`json:"user_is_banned"`
}

type RAnimeJson struct {
	Data struct {
		UserIsBanned              bool          `json:"user_is_banned"`
	} `json:"data"`
}

type ChannelStats struct {
	Name string
	Dates []string
	Messages []int
	TotalMessages int
	DailyAverage int
}

type ChannelPick struct {
	ChannelStats map[string]*misc.Channel
	Flag         bool
	Stats        ChannelStats
	Error        bool
}

type UserChangeStats struct {
	Dates			[]string
	DailyAverage 	int
	Change			[]int
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
	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("Error is in HomepageHandler")
		}
	}()

	// Loads the html & css homepage files
	t, err := template.ParseFiles("./web/assets/index.html")
	if err == nil {
		err = t.Execute(w, nil)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func ChannelStatsPageHandler(w http.ResponseWriter, r *http.Request) {

	var (
		dateLabels    []string
		messageCount  []int
		stats         ChannelStats
		totalMessages int
		id            string
		pick          ChannelPick
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("Error is in ChannelStatsPageHandler")
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
			if err == nil {
				err = t.Execute(w, pick)
				if err != nil {
					fmt.Println(err.Error())
				}
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
		if err == nil {
			err = t.Execute(w, pick)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		misc.MapMutex.Unlock()
		return
	} else {
		misc.MapMutex.Lock()
		pick.Flag = true
		misc.MapMutex.Unlock()
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

	stats.Name = misc.ChannelStats[id].Name
	stats.Dates = dateLabels
	stats.Messages = messageCount
	stats.TotalMessages = totalMessages
	if len(dateLabels) != 0 {
		stats.DailyAverage = totalMessages / len(dateLabels)
	}
	pick.Stats = stats


	// Loads the html & css stats files
	t, err := template.ParseFiles("./web/assets/channelstats.html")
	if err == nil {
		err = t.Execute(w, pick)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	misc.MapMutex.Unlock()
}

func UserChangeStatsPageHandler(w http.ResponseWriter, r *http.Request) {

	var (
		dateLabels    	[]string
		changeCount  	[]int
		stats         	UserChangeStats
		totalChange		int
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("Error is in UserChangeStatsPageHandler")
		}
	}()

	// Save dates, sort them and then assign user change int in order of the dates
	misc.MapMutex.Lock()
	for date := range misc.UserStats {
		dateLabels = append(dateLabels, date)
		totalChange += misc.UserStats[date]
	}
	sort.Sort(byDate(dateLabels))
	for i := 0; i < len(dateLabels); i++ {
		changeCount = append(changeCount, misc.UserStats[dateLabels[i]])
	}

	stats.Dates = dateLabels
	stats.Change = changeCount
	if len(dateLabels) != 0 {
		stats.DailyAverage = totalChange / len(dateLabels)
	}


	// Loads the html & css stats files
	t, err := template.ParseFiles("./web/assets/userchangestats.html")
	if err == nil {
		err = t.Execute(w, stats)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	misc.MapMutex.Unlock()
}

// Handles the verification
func VerificationHandler(w http.ResponseWriter, r *http.Request) {

	// Pulls cookie if it exists, else it creates a new one and assigns it
	cookie, _ := r.Cookie("sessid")
	expire := time.Now().Add(15 * time.Second)
	if cookie == nil {
		randNum, _ := randString(64)
		cookieSet := http.Cookie{Name: "sessid", Value: randNum, Expires: expire, HttpOnly: true}
		http.SetCookie(w, &cookieSet)
		cookie = &cookieSet
	}

	var (
		errorVar 			string
		state    			string
		code     			string
		id       			string
		tempUser 			User
		verified 			bool
	)

	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("Error is in VerificationHandler")
		}
	}()

	// Create entry in UserCookieMap if it doesn't exist. Otherwise just update tempUser with map value
	SafeCookieMap.mux.Lock()
	if _, ok := SafeCookieMap.userCookieMap[cookie.Value]; !ok {
		tempUser.Cookie = cookie.Value
		tempUser.Expiry = expire
		tempUser.UsernameDiscrim = ""
		SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
	} else {
		tempUser = *SafeCookieMap.userCookieMap[cookie.Value]
	}
	SafeCookieMap.mux.Unlock()

	// Fetches queries from link if they exist
	queryValues := r.URL.Query()
	id = queryValues.Get("reqvalue")
	state = queryValues.Get("state")
	code = queryValues.Get("code")
	errorVar = queryValues.Get("error")

	// If errorVar exists, stop execution and print error on page
	if errorVar != "" {
		// Set error
		tempUser.Error = "Error: Permission not given in verification. If this was a mistake please try to verify again."
		SafeCookieMap.mux.Lock()
		SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
		SafeCookieMap.mux.Unlock()

		// Loads the html & css verification files
		t, err := template.ParseFiles("web/assets/verification.html")
		if err == nil {
			err = t.Execute(w, SafeCookieMap.userCookieMap[cookie.Value])
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		SafeCookieMap.mux.Lock()
		// Resets assigned Error Message
		if cookie != nil {
			tempUser.Error = ""
			SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
		}
		SafeCookieMap.mux.Unlock()
		return
	}

	// Saves the ID in the user cookie map if it exists
	if id != "" {
		// Decrypts encrypted id from url
		trueid, validid := misc.Decrypt(misc.Key, id)

		if validid {
			// If the user is verifying to another account with this cookie reset the old cookie values
			if tempUser.ID != trueid {
				tempUser.RedditVerifiedStatus = false
				tempUser.DiscordVerifiedStatus = false
				tempUser.RedditName = ""
				tempUser.AccOldEnough = false
				tempUser.UsernameDiscrim = ""
				tempUser.Username = ""
			}

			// Set new decrypted user ID to verify
			tempUser.ID = trueid
			SafeCookieMap.mux.Lock()
			SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
			SafeCookieMap.mux.Unlock()
		}
	}

	// Saves the code in the user cookie map if it exists
	if code != "" {
		tempUser.Code = code
		SafeCookieMap.mux.Lock()
		SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
		SafeCookieMap.mux.Unlock()
	}

	// Sets the username + discrim combo if it exists in memberinfo via ID, also sorts out the reddit verified status
	misc.MapMutex.Lock()
	if _, ok := misc.MemberInfoMap[SafeCookieMap.userCookieMap[cookie.Value].ID]; ok {
		usernameDiscrim := misc.MemberInfoMap[SafeCookieMap.userCookieMap[cookie.Value].ID].Username + "#" + misc.MemberInfoMap[SafeCookieMap.userCookieMap[cookie.Value].ID].Discrim
		tempUser.UsernameDiscrim = usernameDiscrim

		if misc.MemberInfoMap[SafeCookieMap.userCookieMap[cookie.Value].ID].RedditUsername != "" {
			tempUser.RedditVerifiedStatus = true
		}
		SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
	}

	// Verifies user if they have a reddit account linked in memberInfo already, skipping the entire verification process
	if SafeCookieMap.userCookieMap[cookie.Value].ID != "" {
		if _, ok := misc.MemberInfoMap[SafeCookieMap.userCookieMap[cookie.Value].ID]; ok {
			if misc.MemberInfoMap[SafeCookieMap.userCookieMap[cookie.Value].ID].RedditUsername != "" {
				// Overwrites userCookieMap redditName value with the memberinfo one to avoid abuse in changing their reddit usernames
				SafeCookieMap.userCookieMap[cookie.Value].RedditName = misc.MemberInfoMap[SafeCookieMap.userCookieMap[cookie.Value].ID].RedditUsername
				// Verifies user
				err := Verify(cookie, r)
				if err != nil {
					// Sets error message
					tempUser.Error = err.Error()
					SafeCookieMap.userCookieMap[cookie.Value] = &tempUser

					// Loads the html & css verification files
					t, err := template.ParseFiles("web/assets/verification.html")
					if err == nil {
						err = t.Execute(w, SafeCookieMap.userCookieMap[cookie.Value])
						if err != nil {
							fmt.Println(err.Error())
						}
					}
					// Resets assigned Error Message
					if cookie != nil {
						tempUser.Error = ""
						SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
					}
					misc.MapMutex.Unlock()
					return
				}
				verified = true
			}
		}
	}
	misc.MapMutex.Unlock()

	// Verifies Discord and Reddit
	misc.MapMutex.Lock()
	if SafeCookieMap.userCookieMap[cookie.Value].Code != "" && !verified {

		// Discord verification
		if state == "overlordconfirmsdiscord" {
			uname, udiscrim, uid, err := getDiscordUsernameDiscrim(SafeCookieMap.userCookieMap[cookie.Value].Code)
			if err != nil {
				// Sets error message
				tempUser.Error = "Error: Bad discord verification occurred. Please try to verify again."
				SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
			} else if uname != "" && udiscrim != "" && uid != "" {
				// Sets username#discrim for website use
				tempUser.ID = uid
				tempUser.Username = uname
				tempUser.Discriminator = udiscrim
				tempUser.UsernameDiscrim = uname + "#" + udiscrim
				tempUser.DiscordVerifiedStatus = true
				SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
				fmt.Println("discord ver")

				// Verifies user if reddit verification was already completed succesfully
				if SafeCookieMap.userCookieMap[cookie.Value].AccOldEnough && SafeCookieMap.userCookieMap[cookie.Value].ID != "" &&
					SafeCookieMap.userCookieMap[cookie.Value].RedditVerifiedStatus && SafeCookieMap.userCookieMap[cookie.Value].RedditName != "" {
					if _, ok := misc.MemberInfoMap[SafeCookieMap.userCookieMap[cookie.Value].ID]; ok {
						err := Verify(cookie, r)
						if err != nil {
							// Sets error message
							tempUser.Error = err.Error()
							SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
						}
					} else {
						tempUser.Error = "Error: Are you sure you verified with the correct Discord account? It uses the browser Discord account so please go back and check if it is correct. If it is please notify a mod with the following: User not found in memberInfo with the UserCookieMap UserID."
						SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
					}
				}
			} else {
				tempUser.Error = "Error: Needed discord verification values are missing. Please verify again or message a mod."
				SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
			}

			// Prints error if it exists
			if tempUser.Error != "" {
				// Loads the html & css verification files
				t, err := template.ParseFiles("web/assets/verification.html")
				if err == nil {
					err = t.Execute(w, SafeCookieMap.userCookieMap[cookie.Value])
					if err != nil {
						fmt.Println(err.Error())
					}
				}
				// Resets assigned Error Message
				if cookie != nil {
					tempUser.Error = ""
					SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
				}
				misc.MapMutex.Unlock()
				return
			}
		}

		// Reddit verification
		if state == "overlordconfirmsreddit" {
			// Fetches reddit username and checks whether account is at least 1 week old
			Name, DateUnix, err := getRedditUsername(SafeCookieMap.userCookieMap[cookie.Value].Code)
			if err != nil {
				// Sets error message
				tempUser.Error = "Error: Bad reddit verification occurred. Please try to verify again."
				SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
			} else if Name != "" && DateUnix != 0 {
				// Calculate if account is older than a week
				epochT := time.Unix(int64(DateUnix), 0)
				prevWeek := time.Now().AddDate(0, 0, -7)
				accOldEnough := epochT.Before(prevWeek)

				// Print error if acc is not old enough
				if !accOldEnough {
					// Sets error message
					tempUser.Error = "Error: Reddit account is not old enough. Please try again once it is one week old."
					SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
				} else {
					// Either only saves reddit info or verifies if Discord verification was completed successfully
					// Saves the reddit username and acc age bool
					tempUser.RedditName = Name
					tempUser.AccOldEnough = true
					tempUser.RedditVerifiedStatus = true
					SafeCookieMap.userCookieMap[cookie.Value] = &tempUser

					// Verifies user if Discord was verified already
					if SafeCookieMap.userCookieMap[cookie.Value].ID != "" &&
						SafeCookieMap.userCookieMap[cookie.Value].DiscordVerifiedStatus &&
						SafeCookieMap.userCookieMap[cookie.Value].RedditName != "" {
						if _, ok := misc.MemberInfoMap[SafeCookieMap.userCookieMap[cookie.Value].ID]; ok {
							err := Verify(cookie, r)
							if err != nil {
								// Sets error message
								tempUser.Error = err.Error()
								SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
							}
						} else {
							tempUser.Error = "Error: Are you sure you verified with the correct Discord account? It uses the browser Discord account so please go back and check if it is correct. If it is please notify a mod with the following: User not found in memberInfo with the UserCookieMap UserID."
							SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
						}
					}
				}
			} else {
				tempUser.Error = "Error: Needed reddit values are missing. Please verify again or message a mod."
				SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
			}
		}
	}

	// Loads the html & css verification files
	t, err := template.ParseFiles("web/assets/verification.html")
	if err == nil {
		err = t.Execute(w, SafeCookieMap.userCookieMap[cookie.Value])
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	// Resets assigned Error Message
	if cookie != nil {
		tempUser.Error = ""
		SafeCookieMap.userCookieMap[cookie.Value] = &tempUser
	}
	misc.MapMutex.Unlock()
}

// Verifies user on reddit and returns their reddit username
func getRedditUsername(code string) (string, float64, error) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("getRedditUsername func")
		}
	}()

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
	if err != nil {
		return "", 0, err
	}

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

	// Sets needed reqAPI parameters
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

	// Makes a GET request to reddit in reqAPIBan
	reqAPIBan, err := http.NewRequest("GET", "https://oauth.reddit.com/r/anime/about.json", nil)
	if err != nil {
		return "", 0, err
	}

	// Sets needed reqAPIBan parameters
	reqAPIBan.Header.Add("Authorization", "Bearer "+access.RedditAccessToken)
	reqAPIBan.Header.Add("User-Agent", misc.UserAgent)

	// Does the GET request and puts it into the respAPI
	respAPIBan, err := client.Do(reqAPIBan)
	if err != nil {
		return "", 0, err
	}

	// Reads the byte respAPIBan body into bodyAPIBan
	bodyAPIBan, err := ioutil.ReadAll(respAPIBan.Body)
	if err != nil {
		return "", 0, err
	}

	// Initializes user variable of type UserBan to hold /r/anime reddit ban json in
	userBan := RAnimeJson{}

	// Unmarshals all the required json fields in the above user variable
	jsonErr = json.Unmarshal(bodyAPIBan, &userBan)
	if jsonErr != nil {
		return "", 0, err
	}

	// Gives an error if the user is banned on the sub
	if userBan.Data.UserIsBanned {
		return "", 0, fmt.Errorf("Error: Banned users from the subreddit are not allowed on the Discord server.")
	}

	// Returns user reddit username and date of account creation in epoch time
	return user.RedditName, user.AccCreation, err
}

// Verifies user on discord and returns their discord username and discrim
func getDiscordUsernameDiscrim(code string) (string, string, string, error) {

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("Error is in getDiscordUsernameDiscrim func")
		}
	}()

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
		return "", "", "", jsonErr
	}

	return user.Username, user.Discriminator, user.ID, nil
}

// Verifies user by assigning the necessary values
func Verify(cookieValue *http.Cookie, r *http.Request) error {

	var (
		temp 	misc.UserInfo
		userID 	string
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("Error is in Verify func")
		}
	}()

	// Confirms that the map is not empty
	if len(misc.MemberInfoMap) == 0 {
		return fmt.Errorf("Critical Error: MemberInfo is empty. Please notify a mod.")
	}
	// Checks if cookie has expired while doing this
	if cookieValue == nil {
		return fmt.Errorf("Minor Error: Cookie has expired. Please refresh and try again.")
	}
	if _, ok := SafeCookieMap.userCookieMap[cookieValue.Value]; !ok {
		return fmt.Errorf("Rare Error: CookieValue is not in UserCookieMap. Please notify a mod.")
	}
	userID = SafeCookieMap.userCookieMap[cookieValue.Value].ID
	if _, ok := misc.MemberInfoMap[userID]; !ok {
		return fmt.Errorf("Critical Error: Either user does not exist in MemberInfo or the user ID does not exist. Please notify a mod.")
	}

	// Stores time of verification
	t := time.Now()
	z, _ := t.Zone()
	joinDate := t.Format("2006-01-02 15:04:05") + " " + z

	// Assigns needed values to temp
	temp = *misc.MemberInfoMap[userID]
	temp.RedditUsername = SafeCookieMap.userCookieMap[cookieValue.Value].RedditName
	temp.VerifiedDate = joinDate
	misc.MemberInfoMap[userID] = &temp

	// Saves the userID for verified timer
	verifyMap[userID] = userID

	// Confirms that the above happened (possible bug safety net)
	if _, ok := verifyMap[userID]; !ok {
		return fmt.Errorf("Critical Error: User is not in verifyMap. Please notify a mod.")
	}

	// Writes the username to memberInfo.json
	misc.MemberInfoWrite(misc.MemberInfoMap)

	// Adds to verified stats
	misc.VerifiedStats[t.Format(misc.DateFormat)]++
	return nil
}

// Checks if a user in the verify map has the role and if they're verified it gives it to them
func VerifiedRoleAdd(s *discordgo.Session, e *discordgo.Ready) {

	var (
		roleID string
		userInGuild bool
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("Error is in VerifiedRoleAdd func")
		}
	}()

	// Checks every 5 seconds if a user in the verifyMap needs to be given the role, also clears UserCookieMap if expiry date has passed
	for range time.NewTicker(5 * time.Second).C {

		misc.MapMutex.Lock()
		if len(verifyMap) != 0 {
			for userID := range verifyMap {

				// Checks if banned suspected spambot accounts verified
				if misc.MemberInfoMap[userID].SuspectedSpambot {
					for i, banUser := range misc.BannedUsersSlice {
						if banUser.ID == userID {
							misc.BannedUsersSlice = append(misc.BannedUsersSlice[:i], misc.BannedUsersSlice[i+1:]...)
							break
						}
					}
					// Removes the ban
					err := s.GuildBanDelete(config.ServerID, userID)
					if err != nil {
						_, err := s.ChannelMessageSend(config.BotLogID, err.Error())
						if err != nil {
							continue
						}
						continue
					}
					misc.MemberInfoMap[userID].SuspectedSpambot = false
					misc.MemberInfoWrite(misc.MemberInfoMap)
				}

				// Checks if the user is in the server before continuing. Very important to avoid bugs
				userInGuild = isUserInGuild(s, userID)
				if !userInGuild {
					continue
				}

				// Puts all server roles in roles
				roles, err := s.GuildRoles(config.ServerID)
				if err != nil {
					_, err := s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
					if err != nil {
						continue
					}
					continue
				}

				// Fetches ID of Verified role
				for i := 0; i < len(roles); i++ {
					if roles[i].Name == "Verified" {
						roleID = roles[i].ID
						break
					}
				}

				// Assigns Verified role to user
				err = s.GuildMemberRoleAdd(config.ServerID, userID, roleID)
				if err != nil {
					_, err := s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
					if err != nil {
						continue
					}
					continue
				}

				// Alt check
				check := CheckAltAccount(s, userID)
				if !check {
					user, err := s.GuildMember(config.ServerID, userID)
					if err != nil {
						_, err := s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
						if err != nil {
							delete(verifyMap, userID)
							continue
						}
						delete(verifyMap, userID)
						continue
					}
					misc.InitializeUser(user)
					misc.MemberInfoWrite(misc.MemberInfoMap)
				}
				delete(verifyMap, userID)
			}
		}
		misc.MapMutex.Unlock()

		//Clears userCookieMap based on expiry date
		SafeCookieMap.mux.Lock()
		if len(SafeCookieMap.userCookieMap) != 0 {
			now := time.Now()
			for key, cookie := range SafeCookieMap.userCookieMap {
				if now.Sub(cookie.Expiry) > 0 {
					fmt.Println("delet")
					delete(SafeCookieMap.userCookieMap, key)
				}
			}
		}
		SafeCookieMap.mux.Unlock()
	}
}

// Checks if a user is already verified when they join the server and if they are directly assigns them the verified role
func VerifiedAlready(s *discordgo.Session, u *discordgo.GuildMemberAdd) {

	var (
		roleID string
		userID string
	)

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("Error is in VerifiedAlready func")
		}
	}()

	// Pulls info on user if possible
	s.RWMutex.Lock()
	user, err := s.GuildMember(config.ServerID, u.User.ID)
	if err != nil {
		s.RWMutex.Unlock()
		return
	}
	userID = user.User.ID
	s.RWMutex.Unlock()

	// Checks if the user is an already verified one
	misc.MapMutex.Lock()
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
		_, err := s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
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
		_, err := s.ChannelMessageSend(config.BotLogID, err.Error()+"\n"+misc.ErrorLocation(err))
		if err != nil {
			return
		}
		return
	}

	misc.MapMutex.Lock()
	_ = CheckAltAccount(s, userID)
	misc.MapMutex.Unlock()
}

// Function that iterates through memberInfo.json and checks for any alt accounts for that ID. Verification version
func CheckAltAccount(s *discordgo.Session, id string) bool {

	var alts []string

	// Saves program from panic and continues running normally without executing the command if it happens
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Println(rec)
			fmt.Println("Error is in CheckAltAccount func")
		}
	}()

	if len(misc.MemberInfoMap) == 0 {
		return false
	}
	if misc.MemberInfoMap[id] == nil {
		return false
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
	return true
}

// Checks if the user is in the server
func isUserInGuild(s *discordgo.Session, userID string) bool {
	_, err := s.GuildMember(config.ServerID, userID)
	if err != nil {
		return false
	}
	return true
}
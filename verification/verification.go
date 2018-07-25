package verification

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
	"log"
	"net/http"
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

func randString(n int) (string, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	// Pulls cookie if it exists, else it creates a new one and assigns it
	cookieValue, err := r.Cookie("session-id")
	if err != nil {
		randNum, _ := randString(64)
		expire := time.Now().Add(10 * time.Minute)
		cookie := http.Cookie{Name: "session-id", Value: randNum, Expires: expire}
		http.SetCookie(w, &cookie)
		cookieValue = &cookie
	}

	// Initializes needed variables
	var (
		errorVar string
		state    string
		code     string
		id       string
	)

	// Blurb fetches query from link
	queryValues := r.URL.Query()
	id = queryValues.Get("reqvalue")
	state = queryValues.Get("state")
	code = queryValues.Get("code")
	errorVar = queryValues.Get("error")

	// Saves the id in the user map
	if cookieValue != nil {
		if id != "" {

			var temp User

			misc.MapMutex.Lock()

			// Decrypts encrypted id from url
			trueid := misc.Decrypt(misc.Key, id)

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
	}

	if code != "" {
		if UserCookieMap[cookieValue.Value] != nil {

			// Sets code
			var temp User
			misc.MapMutex.Lock()
			temp = *UserCookieMap[cookieValue.Value]
			temp.Code = code
			UserCookieMap[cookieValue.Value] = &temp
			misc.MapMutex.Unlock()
		}
	}

	if errorVar != "" {
		if UserCookieMap[cookieValue.Value] != nil {

			// Sets error message
			var temp User
			misc.MapMutex.Lock()
			temp = *UserCookieMap[cookieValue.Value]
			temp.Error = "Error: Permission not given in verification. If this was a mistake please try to verify again."
			UserCookieMap[cookieValue.Value] = &temp
			misc.MapMutex.Unlock()
		}
	}

	// Reads memberInfo.json
	misc.MemberInfoRead()

	// Fetches user username and discriminator combo for showing in website. Also checks if user is verified already
	if cookieValue != nil {
		if UserCookieMap[cookieValue.Value] != nil {

			var temp User
			temp = *UserCookieMap[cookieValue.Value]

			misc.MapMutex.Lock()

			// Sets the username + discrim combo if it exists, also sorts out the verified status
			if misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID] != nil {

				username := misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID].Username + "#" + misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID].Discrim
				temp.UsernameDiscrim = username

				if misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID].RedditUsername != "" {
					temp.RedditVerifiedStatus = true
				}

			} else {

				temp.UsernameDiscrim = "Invalid User"
			}

			UserCookieMap[cookieValue.Value] = &temp

			misc.MapMutex.Unlock()
		}
	}

	// Reads memberInfo.json
	misc.MemberInfoRead()

	if cookieValue != nil && errorVar == "" {
		if code == "" && id == "" && state == "" {

			// Sets error message
			var temp User
			misc.MapMutex.Lock()
			temp = *UserCookieMap[cookieValue.Value]
			temp.UsernameDiscrim = ""
			UserCookieMap[cookieValue.Value] = &temp
			misc.MapMutex.Unlock()

		} else if UserCookieMap[cookieValue.Value] != nil {
			if misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID] != nil {
				if misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID].RedditUsername == "" {

					if state == "overlordconfirmsdiscord" && UserCookieMap[cookieValue.Value].Code != "" {

						uname, udiscrim, uid := getDiscordUsernameDiscrim(UserCookieMap[cookieValue.Value].Code)

						var temp User
						misc.MapMutex.Lock()

						temp = *UserCookieMap[cookieValue.Value]
						temp.ID = uid
						temp.Username = uname
						temp.Discriminator = udiscrim
						temp.UsernameDiscrim = uname + "#" + udiscrim
						temp.DiscordVerifiedStatus = true
						UserCookieMap[cookieValue.Value] = &temp

						misc.MapMutex.Unlock()

						if UserCookieMap[cookieValue.Value].AccOldEnough == true && UserCookieMap[cookieValue.Value].ID != "" &&
							UserCookieMap[cookieValue.Value].RedditVerifiedStatus == true && UserCookieMap[cookieValue.Value].RedditName != "" {

							// Verifies user
							Verify(cookieValue, r)
						}

					} else if state == "overlordconfirmsstring" && UserCookieMap[cookieValue.Value].Code != "" {

						// Fetches reddit username and checks whether account is at least 1 week old
						Name, DateUnix = getRedditUsername(UserCookieMap[cookieValue.Value].Code)

						epochT := time.Unix(int64(DateUnix), 0)
						prevWeek := time.Now().AddDate(0, 0, -7)
						accOldEnough := epochT.Before(prevWeek)

						// If account is old enough continue, else show error message
						if accOldEnough != true {

							// Sets error message
							var temp User
							misc.MapMutex.Lock()
							temp = *UserCookieMap[cookieValue.Value]
							temp.Error = "Error: Reddit account is not old enough. Please try again once it is one week old."
							UserCookieMap[cookieValue.Value] = &temp
							misc.MapMutex.Unlock()

						} else if accOldEnough == true && UserCookieMap[cookieValue.Value].ID != "" &&
							UserCookieMap[cookieValue.Value].DiscordVerifiedStatus == true && UserCookieMap[cookieValue.Value].RedditName == "" {

							// Saves the reddit username and acc age bool
							var temp User
							misc.MapMutex.Lock()
							temp = *UserCookieMap[cookieValue.Value]
							temp.RedditName = Name
							temp.RedditVerifiedStatus = true
							temp.AccOldEnough = true
							UserCookieMap[cookieValue.Value] = &temp
							misc.MapMutex.Unlock()

							// Verifies user
							Verify(cookieValue, r)

						} else if accOldEnough == true && UserCookieMap[cookieValue.Value].RedditName == "" {

							// Saves the reddit username and acc age bool
							var temp User
							misc.MapMutex.Lock()
							temp = *UserCookieMap[cookieValue.Value]
							temp.RedditName = Name
							temp.RedditVerifiedStatus = true
							temp.AccOldEnough = true
							UserCookieMap[cookieValue.Value] = &temp
							misc.MapMutex.Unlock()
						}
					}
				} else {

					var temp User
					misc.MapMutex.Lock()
					temp = *UserCookieMap[cookieValue.Value]
					temp.RedditVerifiedStatus = true
					temp.DiscordVerifiedStatus = true
					UserCookieMap[cookieValue.Value] = &temp
					misc.MapMutex.Unlock()
				}
			} else {

				// Sets error message
				var temp User
				misc.MapMutex.Lock()
				temp = *UserCookieMap[cookieValue.Value]
				temp.Error = "Error: User is not in memberInfo, cookie has expired or wrong url. Please rejoin the server and try again."
				UserCookieMap[cookieValue.Value] = &temp
				misc.MapMutex.Unlock()
			}
		} else {

			// Sets error message
			var temp User
			misc.MapMutex.Lock()
			temp.Error = "Error: Cookie has expired. Please try the bot link again."
			UserCookieMap[cookieValue.Value] = &temp
			misc.MapMutex.Unlock()
		}
	}

	// Loads the html index file
	t, err := template.ParseFiles("verification/web/index.html")
	if err != nil {

		fmt.Print("Error:", err)
	}
	err = t.Execute(w, UserCookieMap[cookieValue.Value])
	if err != nil {

		fmt.Println("Error:", err)
	}

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
func getRedditUsername(code string) (string, float64) {

	// Initializes client
	client := &http.Client{Timeout: time.Second * 2}

	// Sets reddit required post info
	POSTinfo := "grant_type=authorization_code&code=" + code + "&redirect_uri=http://localhost:3000/"

	// Starts request to reddit
	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", bytes.NewBuffer([]byte(POSTinfo)))
	if err != nil {
		log.Fatalln(err)
	}

	// Sets needed request parameters User Agent and Basic Auth
	req.Header.Set("User-Agent", "Discord-Reddit verification (by /u/thechosenapiks)")
	req.SetBasicAuth("6qhzdUSgF6185A", "eZJjs6_oHBDREUGyK-ktt60d_xs")
	resp, err := client.Do(req)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// Initializes Access type variable to hold data
	access := Access{}

	// Unmarshals json info into the above access variable to hold
	jsonErr := json.Unmarshal(body, &access)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// Makes a GET request to reddit in reqAPI
	reqAPI, err := http.NewRequest("GET", "https://oauth.reddit.com/api/v1/me", nil)
	if err != nil {
		log.Fatalln(err)
	}

	// Sets needed reqAPI paraemeters
	reqAPI.Header.Add("Authorization", "Bearer "+access.RedditAccessToken)
	reqAPI.Header.Add("User-Agent", "Discord-Reddit verification (by /u/thechosenapiks)")

	// Does the GET request and puts it into the respAPI
	respAPI, err := client.Do(reqAPI)
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Reads the byte respAPI body into bodyAPI
	bodyAPI, err := ioutil.ReadAll(respAPI.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// Initializes user variable of type User to hold reddit json in
	user := User{}

	// Unmarshals all the required json fields in the above user variable
	jsonErr = json.Unmarshal(bodyAPI, &user)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// Returns user reddit username and date of account creation in epoch time
	return user.RedditName, user.AccCreation
}

// Verifies user on discord and returns their discord username and discrim
func getDiscordUsernameDiscrim(code string) (string, string, string) {

	// Sets discord verification variables
	discordConf := oauth2.Config{
		ClientID:     "431328912090464266",
		ClientSecret: "BNdGM_YqEgPU9h3mXHG0OFd5FB8_Msum",
		Scopes:       []string{"identity"},
		RedirectURL:  "http://localhost:3000/",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}

	token, err := discordConf.Exchange(oauth2.NoContext, code)
	if err != nil {

		fmt.Println("Error:", err)
	}

	// Initializes client
	client := &http.Client{Timeout: time.Second * 2}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/@me", "https://discordapp.com/api"), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	// Reads the byte respAPI body into bodyAPI
	bodyAPI, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// Initializes user variable of type User to hold reddit json in
	user := User{}

	// Unmarshals all the required json fields in the above user variable
	jsonErr := json.Unmarshal(bodyAPI, &user)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return user.Username, user.Discriminator, user.ID
}

// Verifies user by assigning the necessary values
func Verify(cookieValue *http.Cookie, r *http.Request) {

	// Reads memberInfo.json
	misc.MemberInfoRead()

	// Confirms that the map is not empty
	if misc.MemberInfoMap != nil {

		// Checks if cookie has expired while doing this
		if cookieValue != nil {

			//Stores time of verification
			t := time.Now()
			z, _ := t.Zone()
			join := t.Format("2006-01-02 15:04:05") + " " + z

			// Assigns needed values to temp
			var temp misc.UserInfo
			misc.MapMutex.Lock()
			temp = *misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID]
			temp.RedditUsername = UserCookieMap[cookieValue.Value].RedditName
			temp.VerifiedDate = join
			misc.MemberInfoMap[UserCookieMap[cookieValue.Value].ID] = &temp
			misc.MapMutex.Unlock()

		}

		// Writes the username to memberInfo.json
		misc.MemberInfoWrite(misc.MemberInfoMap)

	} else {

		fmt.Println("Error: MemberInfo Map is empty")
	}
}

// Checks if a user in the cookie map has the role and if they're verified it gives it to them, also deletes expired map fields
func VerifiedRoleAdd(s *discordgo.Session, e *discordgo.Ready) {

	//Checks every 10 seconds if a user in the UserCookieMap needs to be given the role
	for range time.NewTicker(10 * time.Second).C {
		if UserCookieMap != nil {
			for key := range UserCookieMap {

				if UserCookieMap[key].RedditName != "" && UserCookieMap[key].DiscordVerifiedStatus == true &&
					UserCookieMap[key].RedditVerifiedStatus == true {

					// Initializes var roleID which will keep the Verified role ID
					var roleID string

					// Puts all server roles in roles
					roles, err := s.GuildRoles(config.ServerID)
					if err != nil {

						fmt.Println("Error:", err)
					}

					// Fetches ID of Verified role
					for i := 0; i < len(roles); i++ {
						if roles[i].Name == "Verified" {

							roleID = roles[i].ID
						}
					}

					// Assigns role
					s.GuildMemberRoleAdd(config.ServerID, UserCookieMap[key].ID, roleID)

					if UserCookieMap[key].AltCheck == false {

						CheckAltAccount(s, UserCookieMap[key].ID)

						misc.MapMutex.Lock()
						UserCookieMap[key].AltCheck = true
						misc.MapMutex.Unlock()
					}
				}
			}
		}
	}
}

// Checks if a user is already verified when they join the server and if they are directly assigns them the verified role
func VerifiedAlready(s *discordgo.Session, u *discordgo.GuildMemberAdd) {

	misc.MemberInfoRead()

	// Checks if the user is an already verified one
	if misc.MemberInfoMap != nil {
		if misc.MemberInfoMap[u.User.ID] != nil {
			if misc.MemberInfoMap[u.User.ID].RedditUsername != "" {

				// Initializes var roleID which will keep the Verified role ID
				var roleID string

				// Puts all server roles in roles
				roles, err := s.GuildRoles(config.ServerID)
				if err != nil {

					fmt.Println("Error:", err)
				}

				// Fetches ID of Verified role
				for i := 0; i < len(roles); i++ {
					if roles[i].Name == "Verified" {

						roleID = roles[i].ID
					}
				}

				// Assigns role
				s.GuildMemberRoleAdd(config.ServerID, u.User.ID, roleID)

				CheckAltAccount(s, u.User.ID)
			}
		}
	}
}

// Function that iterates through memberInfo.json and checks for any alt accounts for that ID. Verification version
func CheckAltAccount(s *discordgo.Session, id string) {

	// Initializes alts string slice to hold IDs of alts of that reddit username
	var alts []string

	// Reads memberInfo
	misc.MemberInfoRead()

	// Iterates through all users in memberInfo.json
	for userOne := range misc.MemberInfoMap {

		// Checks if the current user has the same reddit username as userCookieMap user
		if misc.MemberInfoMap[userOne].RedditUsername == misc.MemberInfoMap[id].RedditUsername {

			alts = append(alts, misc.MemberInfoMap[userOne].ID)
		}
	}

	// If there's more than one account with that reddit username print a message
	if len(alts) > 1 {

		// Forms the success string
		success := "**Alternate Account Verified:** \n\n"
		for i := 0; i < len(alts); i++ {

			success = success + "<@" + alts[i] + "> \n"
		}

		// Prints the alts in bot-log channel
		_, err := s.ChannelMessageSend(config.BotLogID, success)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}
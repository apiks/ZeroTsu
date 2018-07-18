package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/r-anime/ZeroTsu/config"
	"github.com/r-anime/ZeroTsu/misc"
)

//Creates a named channel, named role and checks for mod perms
func CreateChannelHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Checks if it's within the /r/anime server
	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		ch, err = s.Channel(m.ChannelID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
	if ch.GuildID == config.ServerID {

		if strings.HasPrefix(m.Content, config.BotPrefix) {

			//Fetches info on message author
			mem, err := s.State.Member(config.ServerID, m.Author.ID)
			if err != nil {
				mem, err = s.GuildMember(config.ServerID, m.Author.ID)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}

			//Checks for if user has permissions
			if misc.HasPermissions(mem) {

				//Puts the command to lowercase
				messageLowercase := strings.ToLower(m.Content)

				//Checks if the prefix + "create " specifically is said
				if strings.HasPrefix(messageLowercase, config.BotPrefix+"create ") && (messageLowercase != (config.BotPrefix + "create")) {

					//Initializes needed variables
					var (
						roleID            string
						category          string
						channelEdit       discordgo.ChannelEdit
						muted             string
						airing            string
						roleName          string
						isAiring          bool
						isGeneral         bool
						categoryExists    bool
						typeExists        bool
						descriptionExists bool
						categoryIndex     int
						typeIndex         int
						descriptionEdit   discordgo.ChannelEdit
						description       string
						descriptionSlice  []string
						typeParameter     string
					)

					//Initializes name variable without "create "
					name := strings.Replace(messageLowercase, config.BotPrefix+"create ", "", -1)
					roleName = name //Fixes naming issue

					//Separates every word in name and puts it in a slice
					commandStrings := strings.Split(name, " ")

					//Assigns the length of commandStrings slice to commandStringsLen
					commandStringsLen := len(commandStrings)

					//Checks if [category] or [type] exist and assigns their slice index if they do. Done because of possible [description]
					for i := 0; i < commandStringsLen; i++ {
						_, err = strconv.ParseInt(commandStrings[i], 10, 64)
						if len(commandStrings[i]) == 18 && err == nil {

							categoryExists = true
							categoryIndex = i
						}

						if commandStrings[i] == "airing" ||
							commandStrings[i] == "general" ||
							commandStrings[i] == "opt-in" ||
							commandStrings[i] == "optin" {

							typeExists = true
							typeIndex = i
						}
					}

					//Saves a copy of original commandStrings slice
					commandStringsCopy := commandStrings

					//Sets commandStringsLen to - 1 because we need to use it as index further ahead
					commandStringsLen = len(commandStrings) - 1

					//If there is a category it sets it.
					if categoryExists == true {

						//Removes category from name and sets category variable for later use
						category = commandStrings[categoryIndex]
						name = strings.Replace(name, commandStrings[categoryIndex], "", -1)
						roleName = name
						commandStrings = append(commandStrings[:categoryIndex], commandStrings[categoryIndex+1:]...)

						//Updates commandStringsLen as index
						commandStringsLen = len(commandStrings) - 1
					}

					//If there is a type it sets it.
					if typeExists == true {

						if commandStrings[typeIndex] == "airing" {

							isAiring = true

							//Removes airing from name and saves typeParameter variable for later use
							typeParameter = commandStrings[typeIndex]
							name = strings.Replace(name, commandStrings[typeIndex], "", -1)
							roleName = name
							commandStrings = append(commandStrings[:typeIndex], commandStrings[typeIndex+1:]...)

							//Updates commandStringsLen
							commandStringsLen = len(commandStrings) - 1

						} else if commandStrings[typeIndex] == "general" {

							isGeneral = true

							//Removes general from name
							typeParameter = commandStrings[typeIndex]
							name = strings.Replace(name, commandStrings[typeIndex], "", -1)
							roleName = name
							commandStrings = append(commandStrings[:typeIndex], commandStrings[typeIndex+1:]...)

							//Updates commandStringsLen
							commandStringsLen = len(commandStrings) - 1

						} else if commandStrings[typeIndex] == "opt-in" ||
							commandStrings[typeIndex] == "optin" {

							//Removes general from name
							typeParameter = commandStrings[typeIndex]
							name = strings.Replace(name, commandStrings[typeIndex], "", -1)
							roleName = name
							commandStrings = append(commandStrings[:typeIndex], commandStrings[typeIndex+1:]...)

							//Updates commandStringsLen
							commandStringsLen = len(commandStrings) - 1
						}
					}

					//Checks for description and if so, saves the description in its variable
					if (categoryExists == true && commandStringsCopy[categoryIndex] != "") ||
						(typeExists == true && commandStringsCopy[typeIndex] != "") {

						descriptionExists = true

						if categoryExists == true {

							descriptionSlice = strings.SplitAfter(m.Content, category)
						} else {

							descriptionSlice = strings.SplitAfter(m.Content, typeParameter)
						}

						//Makes the description the second element of the slice above
						description = descriptionSlice[1]

						//Makes a copy of description that it puts to lowercase
						descriptionLowercase := strings.ToLower(description)

						//Removes description from name
						name = strings.Replace(name, descriptionLowercase, "", -1)
						roleName = name
					}

					//On role creation it edits role and adds to SpoilerMap map. Then it writes to spoiler roles database
					s.AddHandlerOnce(func(s *discordgo.Session, g *discordgo.GuildRoleCreate) {

						roleID = g.Role.ID
						tempRole := discordgo.Role{

							ID:   roleID,
							Name: roleName,
						}

						s.GuildRoleEdit(config.ServerID, roleID, roleName, 0, false, 0, false)

						misc.SpoilerMap[roleID] = &tempRole
						misc.SpoilerRolesWrite(misc.SpoilerMap)
					})

					//Creates the new role
					s.GuildRoleCreate(config.ServerID)

					//Prepares to add and edit the new role to the new channel and fixes mod perms
					s.AddHandlerOnce(func(s *discordgo.Session, g *discordgo.ChannelCreate) {

						//Assigns permission overwrites
						for _, goodRole := range config.CommandRoles {

							//Mod perms
							err = s.ChannelPermissionSet(g.Channel.ID, goodRole, "role", misc.SpoilerPerms, 0)
							if err != nil {

								fmt.Println("Error: ", err)
							}
						}

						if isGeneral == false {

							//Everyone perms
							err = s.ChannelPermissionSet(g.Channel.ID, config.ServerID, "role", 0, misc.SpoilerPerms)
							if err != nil {

								fmt.Println("Error: ", err)
							}
						}
					})

					//Pulls info on server channel
					deb, err := s.GuildRoles(config.ServerID)
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Finds ID of Muted role
					for i := 0; i < len(deb); i++ {
						if deb[i].Name == "Muted" {

							muted = deb[i].ID
						} else if isAiring == true && deb[i].Name == "airing" {

							airing = deb[i].ID
						}
					}

					if isGeneral == false {

						//Adds role perms
						s.AddHandlerOnce(func(s *discordgo.Session, g *discordgo.ChannelCreate) {

							err = s.ChannelPermissionSet(g.Channel.ID, roleID, "role", misc.SpoilerPerms, 0)
							if err != nil {

								fmt.Println("Error: ", err)
							}
						})
					}

					//Adds muted perms
					s.AddHandlerOnce(func(s *discordgo.Session, g *discordgo.ChannelCreate) {

						err = s.ChannelPermissionSet(g.Channel.ID, muted, "role", 0, discordgo.PermissionSendMessages)
						if err != nil {

							fmt.Println("Error: ", err)
						}
					})

					//Adds airing perms if parameter set
					if isAiring == true {
						s.AddHandlerOnce(func(s *discordgo.Session, g *discordgo.ChannelCreate) {

							err = s.ChannelPermissionSet(g.Channel.ID, airing, "role", misc.SpoilerPerms, 0)
							if err != nil {

								fmt.Println("Error: ", err)
							}
						})
					}

					if descriptionExists == true {
						s.AddHandlerOnce(func(s *discordgo.Session, g *discordgo.ChannelCreate) {

							descriptionEdit.Topic = description

							s.ChannelEditComplex(g.Channel.ID, &descriptionEdit)
						})
					}

					//Sets variable text to channel type text for usage in GuildChannelCreate third parameter below
					// text := discordgo.ChannelTypeGuildText

					//Creates the new channel of type text
					cha, err := s.GuildChannelCreate(config.ServerID, roleName, "text")
					if err != nil {

						fmt.Println("Error: ", err)
					}

					//Changes role name to hyphenated form
					roleName = cha.Name

					//Pushes hyphenated form
					s.GuildRoleEdit(config.ServerID, roleID, roleName, 0, false, 0, false)

					//If category input exists it sets that category ID to the new channel parentID(category ID)
					if category != "" {

						//Pulls info on server channel
						chaAll, err := s.GuildChannels(config.ServerID)
						if err != nil {

							fmt.Println("Error: ", err)
						}

						for i := 0; i < len(chaAll); i++ {

							//Puts channel name to lowercase
							nameLowercase := strings.ToLower(chaAll[i].Name)

							//Compares if the categoryString is either a valid category name or ID
							if nameLowercase == category || chaAll[i].ID == category {
								if chaAll[i].Type == discordgo.ChannelTypeGuildCategory {

									category = chaAll[i].ID
								}
							}
						}

						channelEdit.ParentID = category

						//Pushes new parentID
						s.ChannelEditComplex(cha.ID, &channelEdit)
					}

					// If the message was called from StartVote, prints it for all to see, else mod-only message
					if m.Author.ID == config.BotID {

						//Prints success for users in channel
						success := "Channel `" + roleName + "` has been created. Use `" + config.BotPrefix + "join " + roleName +
							"` in #bot-commands until reaction join has been set."
						_, err = s.ChannelMessageSend(m.ChannelID, success)
						if err != nil {

							fmt.Println("Error:", err)
						}
					} else {

						//Prints success for mod
						success := "Channel and role `" + roleName + "` created. If opt-in please sort in the roles list."
						_, err = s.ChannelMessageSend(m.ChannelID, success)
						if err != nil {

							fmt.Println("Error:", err)
						}
					}
				}
			}
		}
	}
}

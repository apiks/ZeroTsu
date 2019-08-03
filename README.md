### ZeroTsu is a Discord all-purpose BOT. Its functionalities are the following:

<br/>

* Channel lock for non-mods via permission change

* Show avatar for a target user. Works for people not in the server

* Extensive member system that tracks past member usernames, nicknames, mod-issued warnings, kicks, bans, reddit verification, verification date, whether in the server, timestamps for punishments, server join date

* Punishment system where you can issue warnings, kick or ban people for a set period of time with the bot and log that using the member system, or remove those punishments and unban. Also shows timestamps for all of those

* Website verification system for reddit account linking and confirmation before being able to use the server

* Automated channel creation with various parameters you can give it to make them hidden (opt-in), temporary (auto deletes after a set period of time) and other

* Automated channel creation via non-mod started vote for a temp (auto deleted) channel with a minimum requirement of votes and hard cap of 3 at a time

* Tracks server emoji stats as well as message stats for general and opt-in channels. User gain/loss per day

* Regex-facilitated filter for phrases.

* Full spoiler/opt-in/hidden channel support with reaction based role-giving or just join/leave commands. Tracks hidden channels between two dummy roles.

* Sort all BOT created optin/spoiler roles between the two dummy roles alphabetically

* Sort a category's channels alphabetically

* Subreddit RSS system that will post a thread if it sees it containing specific phrases and is of a set author. Set for /r/anime but can be changed in the rss.go file

* BOT say/edit commands that any mod can use to send or edit important messages with the BOT, or pretend they're a ROBOT

* Automatically give roles to a user when they join a voice channel, and remove them when they leave it. Fully customizable with multiple roles per voice channel and vice versa.

* RemindMe feature where it either messages you or pings you with a message you've set after a period of time you've set

* Optional Waifu system where you can add names to a list, and each user can roll for a name only once. Users can trade them.

<br/>

How to install:
1. Download in a folder.
2. Edit config.json with your own values. This is mostly if you're using the Website Verification which requires more setup

       BotID is the ID of the BOT you are using the website. Needed if using Website, otherwise Optional

       ServerID is the ID of the server the BOT is going to using the website on. Needed if using Website, otherwise Optional
       
       BotLogID is the ID of the server on which you're using website's bot log channel. Needed if using Website, otherwise Optional

       Website is the address/ip+:port of the website. Optional
	   
	   PlayingMsg is whatever "Playing" message you want the BOT to display. Owner can change it with the playingmsg command. Optional
	   
	   OwnerID is the user ID of the person with Owner level BOT permissions. Optional

3. Make a file called configsecrets.json in the folder config.json is in and set this up the following way. It's for Website/Verification. Skip it if not using that, or use it if you receive an error about it missing
```javascript
{
  "RedditName": "redditAppName",
  "RedditSecret": "RedditAppSecret",
  "DiscordSecret": "DiscordBOTSecret"
}
```
4. Set your "ZeroTsuToken" environment variable to the BOT token (either hidden on the system env or in config.go ReadConfig func with os.Setenv("ZeroTsuToken", "TOKEN"))
5. Compile in your favorite IDE or compiler with "go build" (or type "set GOOS=[Preferred OS]" to change OS first (like linux) and then "go build".)
6. Invite BOT to server and give it an admin role or equivalent
7. Start the BOT as admin and use
8. Use the .hset command to set up the bot for your server

<br/>

If you have discovered any bugs or have questions, please message Apiks or raise an issue.

If you use the BOT successfuly, please also let Apiks know

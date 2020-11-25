## ZeroTsu is a Discord all-purpose BOT with a focus on Moderation
<p align="center">
	<img src="https://images-wixmp-ed30a86b8c4ca887773594c2.wixmp.com/f/6e4868e2-f52b-4c7d-a984-d5027576b221/dch684c-818cbf96-b76b-4e75-8445-75d1497195b7.png?token=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1cm46YXBwOjdlMGQxODg5ODIyNjQzNzNhNWYwZDQxNWVhMGQyNmUwIiwiaXNzIjoidXJuOmFwcDo3ZTBkMTg4OTgyMjY0MzczYTVmMGQ0MTVlYTBkMjZlMCIsIm9iaiI6W1t7InBhdGgiOiJcL2ZcLzZlNDg2OGUyLWY1MmItNGM3ZC1hOTg0LWQ1MDI3NTc2YjIyMVwvZGNoNjg0Yy04MThjYmY5Ni1iNzZiLTRlNzUtODQ0NS03NWQxNDk3MTk1YjcucG5nIn1dXSwiYXVkIjpbInVybjpzZXJ2aWNlOmZpbGUuZG93bmxvYWQiXX0.w_Pmn6zmDv4NcB9h-lPko3-7qnvGmLqVD7862q59XR8" alt="zero two" width="300" height="300">
</p

<br/>

* Channel lock via dynamic permission change that remembers older permissions

* Anime Schedule that prints what anime are airing when SUBBED. Source: https://AnimeSchedule.net

* Extensive member system that tracks past member usernames, nicknames, mod-issued warnings, kicks, bans, whether in the server, timestamps for punishments, server join date and account creation date

* Punishment system where you can issue warnings, mute, kick or ban people for a set period of time with the bot and log that using the member system, or remove those punishments and unban. Also shows timestamps for all of those. Automatically unbans for temp bans.

* Automated channel creation with various parameters you can give it to make them hidden (opt-in), temporary (auto deletes after a set period of time) and other

* Optional automated channel creation via non-mod started vote for a temp (auto deleted) channel with a minimum requirement of votes and hard cap of 3 at a time

* Tracks server emoji and message stats for normal and opt-in channels. User gain/loss per day

* Regex-facilitated filter for phrases.

* Give roles using reactions or just join/leave commands. Tracks opt-in roles between two dummy roles.

* Sort all BOT created opt-in roles between the two dummy roles alphabetically

* Sort a category's channels alphabetically

* Customizable Reddit RSS system that will post a reddit thread based on its settings. Can set for a specific author, sub, post type (rising, hot, new) and title. Can also auto pin/unpin that message in a channel

* BOT say/edit commands that any mod can use to send or edit important messages with the BOT, or pretend they're a ROBOT

* Automatically give roles to a user when they join a voice channel, and remove them when they leave it. Fully customizable with multiple roles per voice channel and vice versa.

* RemindMe feature where it either messages you or pings you with a message you've set after a period of time you've set

* Optional Waifu system where you can add names to a list, and each user can roll for a name only once. Users can trade them.

<br/>

How to install:
1. Download in a folder.
2. Edit config.json with your own values if you want to.
	   
	   PlayingMsg are whatever "Playing" messages you want the BOT to display. Separate each quoted message with a comma. Owner can change it with the playingmsg command. Optional
	   
	   OwnerID is the user ID of the person with Owner level BOT permissions. Optional

3. Set your "ZeroTsuToken" environment variable to the BOT token (either hidden on the system env or in config.go ReadConfig func with os.Setenv("ZeroTsuToken", "TOKEN"))
4. Compile in your favorite IDE or compiler with "go build" (or type "set GOOS=[Preferred OS]" to change OS first (like linux) and then "go build".)
5. Invite BOT to server and give it an admin role or equivalent
6. Start the BOT as admin and use
7. Use the .hset command to set up the bot for your server

<br/>

If you have discovered any bugs or have questions, please message Apiks or raise an issue.

If you use the BOT successfuly, please also let Apiks know.

<br/>

Official Discord Support Server: https://discord.gg/BDT8Twv

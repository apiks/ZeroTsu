package entities

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GuildSettingsMongo represents how GuildSettings are stored in MongoDB
type GuildSettingsMongo struct {
	Prefix       string              `bson:"prefix"`
	BotLog       ChannelMongo        `bson:"bot_log_id"`
	CommandRoles []RoleMongo         `bson:"command_roles"`
	MutedRole    RoleMongo           `bson:"muted_role"`
	VoiceChas    []VoiceChannelMongo `bson:"voice_chas"`
	ModOnly      bool                `bson:"mod_only"`
	Donghua      bool                `bson:"donghua"`
	ReactsModule bool                `bson:"reacts_module"`
	PingMessage  string              `bson:"ping_message"`
	Premium      bool                `bson:"premium"`
}

// LoadGuildSettings retrieves guild settings from MongoDB
func LoadGuildSettings(guildID string) (GuildSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		GuildSettings GuildSettingsMongo `bson:"guild_settings"`
	}

	opts := options.FindOne().SetProjection(bson.M{"guild_settings": 1})
	err := GuildCollection.FindOne(ctx, bson.M{"_id": guildID}, opts).Decode(&result)
	if err != nil {
		return GuildSettings{}, fmt.Errorf("failed to load guild settings for guild %s: %v", guildID, err)
	}

	return ConvertMongoToGuildSettings(result.GuildSettings), nil
}

// SaveGuildSettings updates guild settings in MongoDB
func SaveGuildSettings(guildID string, settings GuildSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mongoSettings := ConvertGuildSettings(settings)

	filter := bson.M{"_id": guildID}
	update := bson.M{"$set": bson.M{"guild_settings": mongoSettings}}

	_, err := GuildCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to save guild settings for guild %s: %v", guildID, err)
	}

	return nil
}

// ConvertGuildSettings removes RWMutex and prepares for MongoDB
func ConvertGuildSettings(gs GuildSettings) GuildSettingsMongo {
	// Convert command roles
	commandRolesMongo := make([]RoleMongo, len(gs.CommandRoles))
	for i, role := range gs.CommandRoles {
		commandRolesMongo[i] = RoleMongo{
			Name:     role.Name,
			ID:       role.ID,
			Position: role.Position,
		}
	}

	// Convert voice channels
	voiceChasMongo := make([]VoiceChannelMongo, len(gs.VoiceChas))
	for i, vc := range gs.VoiceChas {
		rolesMongo := make([]RoleMongo, len(vc.Roles))
		for j, role := range vc.Roles {
			rolesMongo[j] = RoleMongo{
				Name:     role.Name,
				ID:       role.ID,
				Position: role.Position,
			}
		}

		voiceChasMongo[i] = VoiceChannelMongo{
			Name:  vc.Name,
			ID:    vc.ID,
			Roles: rolesMongo,
		}
	}

	return GuildSettingsMongo{
		Prefix:       gs.Prefix,
		BotLog:       ConvertChannelToMongo(gs.BotLog),
		CommandRoles: commandRolesMongo,
		MutedRole:    ConvertRoleToMongo(gs.MutedRole),
		VoiceChas:    voiceChasMongo,
		ModOnly:      gs.ModOnly,
		Donghua:      gs.Donghua,
		ReactsModule: gs.ReactsModule,
		PingMessage:  gs.PingMessage,
		Premium:      gs.Premium,
	}
}

// ConvertMongoToGuildSettings converts GuildSettingsMongo back to entities.GuildSettings
func ConvertMongoToGuildSettings(gs GuildSettingsMongo) GuildSettings {
	// Convert command roles
	commandRoles := make([]Role, len(gs.CommandRoles))
	for i, role := range gs.CommandRoles {
		commandRoles[i] = Role{
			Name:     role.Name,
			ID:       role.ID,
			Position: role.Position,
		}
	}

	// Convert voice channels
	voiceChas := make([]VoiceCha, len(gs.VoiceChas))
	for i, vc := range gs.VoiceChas {
		roles := make([]Role, len(vc.Roles))
		for j, role := range vc.Roles {
			roles[j] = Role{
				Name:     role.Name,
				ID:       role.ID,
				Position: role.Position,
			}
		}

		voiceChas[i] = VoiceCha{
			Name:  vc.Name,
			ID:    vc.ID,
			Roles: roles,
		}
	}

	// Return converted GuildSettings
	return GuildSettings{
		Prefix:       gs.Prefix,
		BotLog:       ConvertMongoToChannel(gs.BotLog),
		CommandRoles: commandRoles,
		MutedRole:    ConvertMongoToRole(gs.MutedRole),
		VoiceChas:    voiceChas,
		ModOnly:      gs.ModOnly,
		Donghua:      gs.Donghua,
		ReactsModule: gs.ReactsModule,
		PingMessage:  gs.PingMessage,
		Premium:      gs.Premium,
	}
}

// ConvertChannelToMongo converts entities.Cha to ChannelMongo
func ConvertChannelToMongo(ch Cha) ChannelMongo {
	return ChannelMongo{
		Name:   ch.Name,
		ID:     ch.ID,
		RoleID: ch.RoleID,
	}
}

// ConvertMongoToChannel converts ChannelMongo to entities.Cha
func ConvertMongoToChannel(ch ChannelMongo) Cha {
	return Cha{
		Name:   ch.Name,
		ID:     ch.ID,
		RoleID: ch.RoleID,
	}
}

// ConvertRoleToMongo converts entities.Role to RoleMongo
func ConvertRoleToMongo(role Role) RoleMongo {
	return RoleMongo{
		Name:     role.Name,
		ID:       role.ID,
		Position: role.Position,
	}
}

// ConvertMongoToRole converts RoleMongo to entities.Role
func ConvertMongoToRole(role RoleMongo) Role {
	return Role{
		Name:     role.Name,
		ID:       role.ID,
		Position: role.Position,
	}
}

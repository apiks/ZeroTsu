package entities

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
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

	err := GuildCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return GuildSettings{}, nil
	}
	if err != nil {
		return GuildSettings{}, fmt.Errorf("failed to load guild settings for guild %s: %v", guildID, err)
	}

	return ConvertMongoToGuildSettings(result.GuildSettings), nil
}

// SaveGuildSettings updates guild settings in MongoDB
func SaveGuildSettings(guildID string, settings GuildSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Sanitize the settings before storing
	mongoSettings := ConvertGuildSettings(settings)

	filter := bson.M{"_id": guildID}
	update := bson.M{
		"$set":         bson.M{"guild_settings": mongoSettings},
		"$setOnInsert": bson.M{"_id": guildID},
	}

	_, err := GuildCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to save guild settings for guild %s: %v", guildID, err)
	}

	return nil
}

// ConvertGuildSettings removes RWMutex and prepares for MongoDB
func ConvertGuildSettings(gs GuildSettings) GuildSettingsMongo {
	commandRolesMongo := make([]RoleMongo, len(gs.CommandRoles))
	for i := range gs.CommandRoles {
		commandRolesMongo[i] = ConvertRoleToMongo(gs.CommandRoles[i])
	}

	voiceChasMongo := make([]VoiceChannelMongo, len(gs.VoiceChas))
	for i := range gs.VoiceChas {
		voiceChasMongo[i] = VoiceChannelMongo{
			Name:  gs.VoiceChas[i].Name,
			ID:    gs.VoiceChas[i].ID,
			Roles: make([]RoleMongo, len(gs.VoiceChas[i].Roles)),
		}
		for j := range gs.VoiceChas[i].Roles {
			voiceChasMongo[i].Roles[j] = ConvertRoleToMongo(gs.VoiceChas[i].Roles[j])
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
	commandRoles := make([]Role, len(gs.CommandRoles))
	for i := range gs.CommandRoles {
		commandRoles[i] = ConvertMongoToRole(gs.CommandRoles[i])
	}

	voiceChas := make([]VoiceCha, len(gs.VoiceChas))
	for i := range gs.VoiceChas {
		voiceChas[i] = VoiceCha{
			Name:  gs.VoiceChas[i].Name,
			ID:    gs.VoiceChas[i].ID,
			Roles: make([]Role, len(gs.VoiceChas[i].Roles)),
		}
		for j := range gs.VoiceChas[i].Roles {
			voiceChas[i].Roles[j] = ConvertMongoToRole(gs.VoiceChas[i].Roles[j])
		}
	}

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

func EnsureGuildSettingsIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.M{"guild_settings.prefix": 1},
			Options: options.Index(),
		},
	}

	_, err := GuildCollection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		fmt.Println("Failed to create index for guild settings:", err)
	}
}

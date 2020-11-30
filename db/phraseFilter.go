package db

import (
	"fmt"
	"github.com/r-anime/ZeroTsu/entities"
	"strings"
)

// GetGuildFilters returns the guild's in-memory filters
func GetGuildFilters(guildID string) []entities.Filter {
	entities.HandleNewGuild(guildID)

	entities.Guilds.RLock()
	defer entities.Guilds.RUnlock()

	return entities.Guilds.DB[guildID].GetFilters()
}

// SetGuildFilter sets a target guild's filter in-memory
func SetGuildFilter(guildID string, filter entities.Filter, delete ...bool) error {
	entities.HandleNewGuild(guildID)

	entities.Guilds.Lock()

	if len(delete) == 0 {
		if entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetFilters()) >= 300 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the filter limit (300) for this premium server.")
		} else if !entities.Guilds.DB[guildID].GetGuildSettings().GetPremium() && len(entities.Guilds.DB[guildID].GetFilters()) >= 50 {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: You have reached the filter limit (50) for this server. Please remove some or increase them to 300 by upgrading to a premium server at <https://patreon.com/animeschedule>")
		}
	}

	filter = filter.SetFilter(strings.ToLower(filter.GetFilter()))

	if len(delete) == 0 {
		var exists bool
		for _, guildFilter := range entities.Guilds.DB[guildID].GetFilters() {
			if strings.ToLower(guildFilter.GetFilter()) == filter.GetFilter() {
				exists = true
				break
			}
		}

		if !exists {
			entities.Guilds.DB[guildID].AppendToFilters(filter)
		} else {
			entities.Guilds.Unlock()
			return fmt.Errorf("Error: That phrase filter already exists.")
		}
	} else {
		err := deleteGuildFilter(guildID, filter)
		if err != nil {
			entities.Guilds.Unlock()
			return err
		}
	}

	entities.Guilds.Unlock()

	entities.Guilds.DB[guildID].WriteData("filters", entities.Guilds.DB[guildID].GetFilters())

	return nil
}

// deleteGuildFilter safely deletes a phrase filter from the filters slice
func deleteGuildFilter(guildID string, filter entities.Filter) error {
	var exists bool

	for i, guildFilter := range entities.Guilds.DB[guildID].GetFilters() {
		if strings.ToLower(guildFilter.Filter) == filter.Filter {
			entities.Guilds.DB[guildID].RemoveFromFilters(i)
			exists = true
			break
		}
	}

	if !exists {
		return fmt.Errorf("Error: No such phrase filter exists.")
	}

	return nil
}

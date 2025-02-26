package db

import (
	"github.com/r-anime/ZeroTsu/entities"
	"log"
)

// GetReminders retrieves a user's or guild's reminders from MongoDB
func GetReminders(id string) *entities.RemindMeSlice {
	reminders, err := entities.GetReminders(id)
	if err != nil {
		log.Println("Error loading reminders:", err)
		return nil
	}
	return reminders
}

// GetDueReminders retrieves reminders that are due for sending
func GetDueReminders() map[string]*entities.RemindMeSlice {
	reminders, err := entities.GetDueReminders()
	if err != nil {
		log.Println("Error fetching due reminders:", err)
		return nil
	}
	return reminders
}

// SetReminder saves a reminder for a user or guild in MongoDB
func SetReminder(id string, remindMe *entities.RemindMe, isGuild bool, premium bool) {
	reminders := GetReminders(id)

	if reminders == nil {
		reminders = &entities.RemindMeSlice{
			RemindMeSlice: []*entities.RemindMe{},
			Guild:         isGuild,
			Premium:       premium,
		}
	}

	// Append new reminder only if remindMe is not nil
	if remindMe != nil {
		reminders.AppendToRemindMeSlice(remindMe)
	}

	// Save updated reminders to MongoDB
	err := entities.SaveReminders(id, reminders)
	if err != nil {
		log.Printf("Error saving reminder for %s: %v\n", id, err)
	}
}

// RemoveReminder deletes a specific reminder for a user or guild in MongoDB
func RemoveReminder(id string, remindID int) {
	reminders := GetReminders(id)
	if reminders == nil {
		return
	}

	// Find and remove reminder by ID
	for index, remindMe := range reminders.GetRemindMeSlice() {
		if remindMe.GetRemindID() == remindID {
			reminders.RemoveFromRemindMeSlice(index)
			break
		}
	}

	// Save updated reminders to MongoDB
	err := entities.SaveReminders(id, reminders)
	if err != nil {
		log.Printf("Error saving reminders after deletion for %s: %v\n", id, err)
	}
}

// RemoveReminders deletes specific reminders for a user or guild in MongoDB
func RemoveReminders(id string, indexes []int) {
	reminders := GetReminders(id)
	if reminders == nil {
		return
	}

	var newReminders []*entities.RemindMe
	for i, remindMe := range reminders.GetRemindMeSlice() {
		if !contains(indexes, i) {
			newReminders = append(newReminders, remindMe)
		}
	}

	// If all reminders are gone, delete from MongoDB
	if len(newReminders) == 0 {
		RemoveAllReminders(id)
	} else {
		reminders.SetRemindMeSlice(newReminders)
		err := entities.SaveReminders(id, reminders)
		if err != nil {
			log.Printf("Error saving reminders after deletion for %s: %v\n", id, err)
		}
	}
}

// Helper function to check if a value is in an array
func contains(arr []int, val int) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}

// RemoveAllReminders deletes all reminders for a user or guild in MongoDB
func RemoveAllReminders(id string) {
	err := entities.SaveReminders(id, &entities.RemindMeSlice{})
	if err != nil {
		log.Printf("Error saving reminders after deleting all for %s: %v\n", id, err)
	}
}

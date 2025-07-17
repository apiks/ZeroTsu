package events

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/r-anime/ZeroTsu/common"
	"github.com/r-anime/ZeroTsu/db"
	"github.com/r-anime/ZeroTsu/embeds"
	"github.com/r-anime/ZeroTsu/entities"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"

	"github.com/r-anime/ZeroTsu/config"
)

const (
	feedCheckLifespanHours = 720 // 30 days
	maxEmbedBatchSize      = 10
)

var (
	redditFeedBlock        Block
	redditFeedWebhookBlock Block
	cleanupCounter         int

	// Object pools for memory reuse
	embedPool = sync.Pool{
		New: func() any {
			return &[]*discordgo.MessageEmbed{}
		},
	}

	feedCheckPool = sync.Pool{
		New: func() any {
			return &[]entities.FeedCheck{}
		},
	}
)

// feedProcessor handles feed processing for a single guild with memory optimization
type feedProcessor struct {
	guildID    string
	session    *discordgo.Session
	webhook    *discordgo.Webhook
	feeds      []entities.Feed
	feedChecks map[string]bool
}

func newFeedProcessor(guildID string, session *discordgo.Session, webhook *discordgo.Webhook, feeds []entities.Feed) *feedProcessor {
	return &feedProcessor{
		guildID:    guildID,
		session:    session,
		webhook:    webhook,
		feeds:      feeds,
		feedChecks: make(map[string]bool, len(feeds)*10),
	}
}

func (fp *feedProcessor) processFeeds(parsedFeeds map[string]*gofeed.Feed) error {
	// Get reusable slices from pools
	embedsSlicePtr := embedPool.Get().(*[]*discordgo.MessageEmbed)
	embedsSlice := *embedsSlicePtr
	defer func() {
		*embedsSlicePtr = embedsSlice[:0]
		embedPool.Put(embedsSlicePtr)
	}()

	newFeedChecksPtr := feedCheckPool.Get().(*[]entities.FeedCheck)
	newFeedChecks := *newFeedChecksPtr
	defer func() {
		*newFeedChecksPtr = newFeedChecks[:0]
		feedCheckPool.Put(newFeedChecksPtr)
	}()

	fp.loadFeedChecks()

	for _, feed := range fp.feeds {
		feedKey := fmt.Sprintf("%s/%s", feed.GetSubreddit(), feed.GetPostType())
		parsedFeed, exists := parsedFeeds[feedKey]
		if !exists {
			continue
		}

		for _, item := range parsedFeed.Items {
			checkKey := fmt.Sprintf("%s_%s", item.GUID, feed.GetChannelID())
			if fp.feedChecks[checkKey] {
				continue
			}

			if !fp.validateFeedItem(feed, item) {
				continue
			}

			embed := embeds.FeedEmbed(&feed, item)
			embedsSlice = append(embedsSlice, embed)

			feedCheck := entities.NewFeedCheck(feed, time.Now(), item.GUID)
			newFeedChecks = append(newFeedChecks, feedCheck)

			fp.feedChecks[checkKey] = true

			if len(embedsSlice) >= maxEmbedBatchSize {
				if err := fp.sendEmbeds(embedsSlice); err != nil {
					return err
				}

				if len(newFeedChecks) > 0 {
					db.SetGuildFeedChecks(fp.guildID, newFeedChecks)
				}

				embedsSlice = embedsSlice[:0]
				newFeedChecks = newFeedChecks[:0]
			}
		}
	}

	if len(embedsSlice) > 0 {
		if err := fp.sendEmbeds(embedsSlice); err != nil {
			return err
		}

		if len(newFeedChecks) > 0 {
			db.SetGuildFeedChecks(fp.guildID, newFeedChecks)
		}
	}

	return nil
}

func (fp *feedProcessor) loadFeedChecks() {
	feedChecks := db.GetGuildFeedChecks(fp.guildID, -1)

	if len(feedChecks) > 0 {
		if fp.feedChecks == nil {
			fp.feedChecks = make(map[string]bool, len(feedChecks))
		}

		for _, check := range feedChecks {
			key := fmt.Sprintf("%s_%s", check.GetGUID(), check.GetFeed().GetChannelID())
			fp.feedChecks[key] = true
		}
	}
}

func (fp *feedProcessor) validateFeedItem(feed entities.Feed, item *gofeed.Item) bool {
	if feed.GetAuthor() != "" && item.Author != nil {
		expectedAuthor := fmt.Sprintf("/u/%s", feed.GetAuthor())
		if strings.ToLower(item.Author.Name) != expectedAuthor {
			return false
		}
	}

	if feed.GetTitle() != "" && !strings.HasPrefix(strings.ToLower(item.Title), feed.GetTitle()) {
		return false
	}

	return true
}

func (fp *feedProcessor) sendEmbeds(embeds []*discordgo.MessageEmbed) error {
	_, err := fp.session.WebhookExecute(fp.webhook.ID, fp.webhook.Token, false, &discordgo.WebhookParams{
		Embeds: embeds,
	})
	return err
}

// FeedWebhookHandler processes Reddit feeds for guilds with webhooks
func FeedWebhookHandler(guildIds []string) {
	redditFeedWebhookBlock.Lock()
	if redditFeedWebhookBlock.Block {
		redditFeedWebhookBlock.Unlock()
		return
	}
	redditFeedWebhookBlock.Block = true
	redditFeedWebhookBlock.Unlock()

	defer func() {
		redditFeedWebhookBlock.Lock()
		redditFeedWebhookBlock.Block = false
		redditFeedWebhookBlock.Unlock()
	}()

	processFeedsWithWebhooks(guildIds)

	// Run cleanup every 10 cycles to avoid interference
	cleanupCounter++
	if cleanupCounter%10 == 0 {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in cleanup goroutine: %v", r)
				}
			}()
			cleanupExpiredFeedChecks(guildIds)
		}()
	}
}

func cleanupExpiredFeedChecks(guildIds []string) {
	const batchSize = 20
	expiredCount := 0

	for i := 0; i < len(guildIds); i += batchSize {
		end := min(i+batchSize, len(guildIds))
		batch := guildIds[i:end]

		for _, guildID := range batch {
			guildFeedChecks := db.GetGuildFeedChecks(guildID, -1)
			if len(guildFeedChecks) == 0 {
				continue
			}

			for _, feedCheck := range guildFeedChecks {
				dateRemoval := feedCheck.GetDate().Add(feedCheckLifespanHours)
				if time.Since(dateRemoval) > 0 {
					db.SetGuildFeedCheck(guildID, feedCheck, true)
					expiredCount++
				}
			}
		}
	}

	if expiredCount > 0 {
		log.Printf("Cleaned up %d expired feed checks", expiredCount)
	}
}

func processFeedsWithWebhooks(guildIds []string) {
	feedsMap := make(map[string][]entities.Feed)
	webhooksMap := make(map[string]*discordgo.Webhook)
	parsedFeedsMap := make(map[string]*gofeed.Feed)

	const guildBatchSize = 10
	for i := 0; i < len(guildIds); i += guildBatchSize {
		end := min(i+guildBatchSize, len(guildIds))
		batch := guildIds[i:end]
		collectFeedsAndWebhooks(batch, feedsMap, webhooksMap)
	}

	if len(webhooksMap) == 0 {
		return
	}

	processFeedsConcurrently(feedsMap, webhooksMap, parsedFeedsMap, guildIds)
}

func collectFeedsAndWebhooks(guildIds []string, feedsMap map[string][]entities.Feed, webhooksMap map[string]*discordgo.Webhook) {
	for _, guildID := range guildIds {
		guildFeeds := db.GetGuildFeeds(guildID)
		if len(guildFeeds) == 0 {
			continue
		}

		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}
		s := config.Mgr.SessionForGuild(guildIDInt)

		for _, feed := range guildFeeds {
			perms, err := s.State.UserChannelPermissions(s.State.User.ID, feed.GetChannelID())
			if err != nil || perms&discordgo.PermissionManageWebhooks != discordgo.PermissionManageWebhooks || perms&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel || perms&discordgo.PermissionSendMessages != discordgo.PermissionSendMessages {
				continue
			}

			feedsMap[feed.GetChannelID()] = append(feedsMap[feed.GetChannelID()], feed)

			if _, ok := webhooksMap[feed.GetChannelID()]; ok {
				continue
			}

			webhook := getOrCreateWebhook(s, feed.GetChannelID())
			if webhook != nil {
				webhooksMap[feed.GetChannelID()] = webhook
			}
		}
	}
}

func getOrCreateWebhook(s *discordgo.Session, channelID string) *discordgo.Webhook {
	ws, err := s.ChannelWebhooks(channelID)
	if err != nil {
		return nil
	}

	for _, w := range ws {
		if w.User.ID == s.State.User.ID && w.ChannelID == channelID {
			return w
		}
	}

	avatar, err := s.UserAvatarDecode(s.State.User)
	if err != nil {
		return nil
	}

	out := new(bytes.Buffer)
	defer out.Reset()

	err = png.Encode(out, avatar)
	if err != nil {
		return nil
	}

	base64Img := base64.StdEncoding.EncodeToString(out.Bytes())
	out.Reset()

	w, err := s.WebhookCreate(channelID, s.State.User.Username, fmt.Sprintf("data:image/png;base64,%s", base64Img))
	if err != nil {
		return nil
	}
	return w
}

func processFeedsConcurrently(feedsMap map[string][]entities.Feed, webhooksMap map[string]*discordgo.Webhook, parsedFeedsMap map[string]*gofeed.Feed, guildIds []string) {
	feedParseMap := make(map[string]bool)
	for k, feeds := range feedsMap {
		if _, ok := webhooksMap[k]; !ok {
			continue
		}
		for _, feed := range feeds {
			feedParseMap[fmt.Sprintf("%s/%s", feed.GetSubreddit(), feed.GetPostType())] = true
		}
	}

	parseFeedsWithRateLimit(feedParseMap, parsedFeedsMap)

	maxWorkers := runtime.NumCPU()
	if maxWorkers > 4 {
		maxWorkers = 4
	}
	if maxWorkers < 2 {
		maxWorkers = 2
	}

	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	const batchSize = 5
	for i := 0; i < len(guildIds); i += batchSize {
		end := i + batchSize
		if end > len(guildIds) {
			end = len(guildIds)
		}

		batch := guildIds[i:end]
		for _, guildID := range batch {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(guid string) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Panic in feed processing goroutine for guild %s: %v", guid, r)
					}
					wg.Done()
					<-semaphore
				}()

				processGuildFeeds(guid, feedsMap, webhooksMap, parsedFeedsMap)
			}(guildID)
		}

		wg.Wait()
	}
}

func parseFeedsWithRateLimit(feedParseMap map[string]bool, parsedFeedsMap map[string]*gofeed.Feed) {
	const maxConcurrentParses = 5
	semaphore := make(chan struct{}, maxConcurrentParses)
	var wg sync.WaitGroup

	for k := range feedParseMap {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(key string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in feed parsing goroutine for key %s: %v", key, r)
				}
				wg.Done()
				<-semaphore
			}()

			time.Sleep(time.Second * 2)

			feedKey := key
			feedKey = strings.TrimSuffix(feedKey, "/")
			feedKey = strings.TrimPrefix(feedKey, "https://www.reddit.com/r/")
			feedKey = strings.TrimPrefix(feedKey, "/")

			feedParser, statusCode, err := common.GetRedditRSSFeed(fmt.Sprintf("https://www.reddit.com/r/%s/.rss", feedKey), 1)
			if err != nil {
				if statusCode == 429 {
					log.Println("HIT REDDIT RATE LIMIT feedWebhookHandler!")
					time.Sleep(10 * time.Minute)
				}
				return
			}

			parsedFeedsMap[feedKey] = feedParser
		}(k)
	}

	wg.Wait()
}

func processGuildFeeds(guildID string, feedsMap map[string][]entities.Feed, webhooksMap map[string]*discordgo.Webhook, parsedFeedsMap map[string]*gofeed.Feed) {
	guildFeeds := db.GetGuildFeeds(guildID)
	if len(guildFeeds) == 0 {
		return
	}

	guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
	if err != nil {
		return
	}
	s := config.Mgr.SessionForGuild(guildIDInt)

	for channelID, webhook := range webhooksMap {
		if webhook.GuildID != guildID {
			continue
		}

		feeds, exists := feedsMap[channelID]
		if !exists {
			continue
		}

		processor := newFeedProcessor(guildID, s, webhook, feeds)
		if err := processor.processFeeds(parsedFeedsMap); err != nil {
			log.Printf("Error processing feeds for guild %s: %v", guildID, err)
		}
	}
}

// FeedHandler processes Reddit feeds for guilds without webhooks
func FeedHandler(guildIds []string) {
	redditFeedBlock.Lock()
	if redditFeedBlock.Block {
		redditFeedBlock.Unlock()
		return
	}
	redditFeedBlock.Block = true
	redditFeedBlock.Unlock()

	defer func() {
		redditFeedBlock.Lock()
		redditFeedBlock.Block = false
		redditFeedBlock.Unlock()
	}()

	const batchSize = 5
	for i := 0; i < len(guildIds); i += batchSize {
		end := min(i+batchSize, len(guildIds))
		batch := guildIds[i:end]
		processGuildFeedsWithoutWebhooks(batch)
	}
}

func processGuildFeedsWithoutWebhooks(guildIds []string) {
	for _, guildID := range guildIds {
		guildFeeds := db.GetGuildFeeds(guildID)
		if len(guildFeeds) == 0 {
			continue
		}

		guildFeedChecks := db.GetGuildFeedChecks(guildID, -1)

		feedCheckMap := make(map[string]bool, len(guildFeedChecks))
		for _, check := range guildFeedChecks {
			key := fmt.Sprintf("%s_%s", check.GetGUID(), check.GetFeed().GetChannelID())
			feedCheckMap[key] = true
		}

		guildIDInt, err := strconv.ParseInt(guildID, 10, 64)
		if err != nil {
			continue
		}
		s := config.Mgr.SessionForGuild(guildIDInt)

		for _, feed := range guildFeeds {
			perms, err := s.State.UserChannelPermissions(s.State.User.ID, feed.GetChannelID())
			if err != nil || perms&discordgo.PermissionManageWebhooks == discordgo.PermissionManageWebhooks || perms&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel || perms&discordgo.PermissionSendMessages != discordgo.PermissionSendMessages {
				continue
			}

			time.Sleep(4 * time.Second)
			feedParser, statusCode, err := common.GetRedditRSSFeed(fmt.Sprintf("https://www.reddit.com/r/%s/%s/.rss", feed.GetSubreddit(), feed.GetPostType()), 1)
			if err != nil {
				if statusCode == 429 {
					log.Println("HIT REDDIT RATE LIMIT feedHandler!")
					time.Sleep(10 * time.Minute)
				}
				continue
			}

			processFeedItems(s, feed, feedParser, feedCheckMap, guildID)
		}
	}
}

func processFeedItems(s *discordgo.Session, feed entities.Feed, feedParser *gofeed.Feed, feedCheckMap map[string]bool, guildID string) {
	for _, item := range feedParser.Items {
		checkKey := fmt.Sprintf("%s_%s", item.GUID, feed.GetChannelID())
		if feedCheckMap[checkKey] {
			continue
		}

		if !validateFeedItem(feed, item) {
			continue
		}

		time.Sleep(time.Millisecond * 250)

		message, err := embeds.Feed(s, &feed, item)
		if err != nil {
			continue
		}

		db.SetGuildFeedCheck(guildID, entities.NewFeedCheck(feed, time.Now(), item.GUID))

		if feed.GetPin() {
			handleFeedPinning(s, feed, message)
		}
	}
}

func validateFeedItem(feed entities.Feed, item *gofeed.Item) bool {
	// Check author
	if feed.GetAuthor() != "" && item.Author != nil {
		expectedAuthor := fmt.Sprintf("/u/%s", feed.GetAuthor())
		if strings.ToLower(item.Author.Name) != expectedAuthor {
			return false
		}
	}

	// Check title prefix
	if feed.GetTitle() != "" && !strings.HasPrefix(strings.ToLower(item.Title), feed.GetTitle()) {
		return false
	}

	return true
}

func handleFeedPinning(s *discordgo.Session, feed entities.Feed, message *discordgo.Message) {
	perms, err := s.State.UserChannelPermissions(s.State.User.ID, message.ChannelID)
	if err != nil || perms&discordgo.PermissionManageMessages != discordgo.PermissionManageMessages {
		return
	}

	pins, err := s.ChannelMessagesPinned(message.ChannelID)
	if err != nil {
		return
	}

	// Unpin old pins from this subreddit
	for _, pin := range pins {
		if pin.Author.ID != s.State.User.ID || len(pin.Embeds) == 0 || pin.Embeds[0].Author == nil {
			continue
		}

		if strings.HasPrefix(strings.ToLower(pin.Embeds[0].Author.URL), fmt.Sprintf("https://www.reddit.com/r/%s/comments/", feed.GetSubreddit())) {
			_ = s.ChannelMessageUnpin(pin.ChannelID, pin.ID)
		}
	}

	// Pin new message
	_ = s.ChannelMessagePin(message.ChannelID, message.ID)
}

package command

import (
	"fmt"
	"log"
	"strings"
	"trup/db"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

const warnUsage = "warn <@user> <reason>"

func warn(ctx *Context, args []string) {
	if len(args) < 3 {
		ctx.Reply("not enough arguments.")
		return
	}

	user := parseMention(args[1])
	if user == "" {
		user = parseSnowflake(args[1])
	}
	if user == "" {
		ctx.Reply("The first argument must be a user mention.")
		return
	}

	reason := strings.Join(args[2:], " ")

	warningMessageLink := fmt.Sprintf("[(warning)](%s)", makeMessageLink(ctx.Message.GuildID, ctx.Message))

	w := db.NewWarn(ctx.Message.Author.ID, user, reason)
	err := w.Save()
	if err != nil {
		ctx.ReportError("Failed to save your warning", err)
		return
	}

	var nth string
	warnCount, err := db.CountWarns(user)
	if err != nil {
		log.Printf("Failed to count warns for user %s; Error: %s\n", user, err)
	}
	if warnCount > 0 {
		nth = " for the " + humanize.Ordinal(warnCount) + " time"
	}

	taker := ctx.Message.Author
	err = db.NewNote(taker.ID, user, fmt.Sprintf("User was warned for: %s %s", reason, warningMessageLink), db.ManualNote).Save()
	if err != nil {
		log.Printf("Failed to save warning note. Error: %s\n", err)
	}

	_, err = ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, fmt.Sprintf("<@%s> Has been warned%s with reason: %s.", user, nth, reason))
	if err != nil {
		log.Printf("Error sending warning notice: %s\n", err)
	}
	r := ""
	if reason != "" {
		r = " with reason: " + reason
	}

	_, err = ctx.Session.ChannelMessageSend(
		ctx.Env.ChannelModlog,
		fmt.Sprintf("<@%s> was warned by moderator %s%s. They've been warned%s.", user, taker.Username, r, nth),
	)
	if err != nil {
		log.Printf("Error sending warning notice into modlog: %s\n", err)
	}
}

func makeMessageLink(guildID string, m *discordgo.Message) string {
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, m.ChannelID, m.ID)
}

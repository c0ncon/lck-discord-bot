# HOWTO

1. `$ go get github.com/bwmarrin/discordgo`
2. Insert token [HERE](https://github.com/c0ncon/lck-discord-bot/blob/master/src/lck-bot/lck-bot.go#L42)
3. `$ go build lck-bot`

## schedules.json schema

```json
{
  "schedules": [{
    "date": "YYYY-MM-DD",
    "matches": [
      ["MATCH 1 TEAM 1", "MATCH 1 TEAM 2"],
      ["MATCH 2 TEAM 1", "MATCH 2 TEAM 2"],
      ...
    ]
  }, ...]
}
```

Must be ordered by date.

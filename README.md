# lck-discord-bot

## schedules.json

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

## Usage

```sh
$ lck-bot -t BOT_TOKEN
```
or make ```.token``` file
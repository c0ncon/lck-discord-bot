require 'faraday'
require 'json'
require 'date'

url = 'https://esports-api.lolesports.com/persisted/gw/getSchedule?hl=ja-JP&leagueId=98767975604431411'
header = { 'x-api-key': '0TvQnueqKa5mxJntVWt0w4LpLfEkrV1Ta8rQBb9Z' }

events = []
newer = true
until newer.nil?
  response = Faraday.get url do |req|
    req.headers = header
  end
  res = JSON.parse response.body
  events += res['data']['schedule']['events']

  newer = res['data']['schedule']['pages']['newer']
  url += "&pageToken=#{newer}"
end
events.select! { |e| e['state'] == 'unstarted' }

latest_schedule = events.map do |event|
  jst = DateTime.parse(event['startTime']).to_time + (3600 * 9)
  {
    date: jst.strftime('%F'),
    time: jst.strftime('%R'),
    home: event['match']['teams'].first['code'],
    away: event['match']['teams'].last['code']
  }

  # start_date = event['startTime'][0...10]
  # start_hour = (event['startTime'][11..12].to_i + 9).to_s.rjust(2, '0')
  # {
  #   date: start_date,
  #   time: "#{start_hour}:00",
  #   home: event['match']['teams'].first['code'],
  #   away: event['match']['teams'].last['code']
  # }
end
puts latest_schedule.to_json

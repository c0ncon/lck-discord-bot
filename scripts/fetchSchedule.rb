# frozen_string_literal: true

require 'net/http'
require 'json'
require 'time'
require 'date'

TEAM_ALIAS = {
  'Griffin': 'GRF',
  'DAMWON Gaming': 'DWG',
  'SANDBOX Gaming': 'SB',
  'SKT T1': 'SKT',
  'Afreeca Freecs': 'AF',
  'Gen.G Esports': 'GEN',
  'KING-ZONE DragonX': 'KZ',
  'KT Rolster': 'KT',
  'Hanwha Life Esports': 'HLE',
  'JIN AIR Greenwings': 'JAG'
}.freeze
WEEKDAY_KOR = %w[일 월 화 수 목 금 토].freeze

def fetch(uri_str, limit = 10)
  raise ArgumentError, 'too many HTTP redirects' if limit == 0

  response = Net::HTTP.get_response(URI(uri_str))

  case response
  when Net::HTTPSuccess
    response
  when Net::HTTPRedirection then
    location = response['location']
    warn "redirected to #{location}"
    fetch(location, limit - 1)
  else
    response.value
  end
end

URL = URI.parse 'http://www.leagueoflegends.co.kr/modules/esports_intro/ajax/esports_schedule.php?start=20190605&end=20191231'

resp = fetch URL
schedule_str = resp.body.strip

schedule_str.gsub!("\xEF\xBB\xBF".force_encoding(Encoding::BINARY), '')
schedule_str.gsub!(/(\w+\s?:)/) do |key|
  begin
    Time.parse key
    key
  rescue ArgumentError
    "\"#{key[0...-1].strip}\":"
  end
end
schedule_str = schedule_str[1...-1]

schedule = JSON.parse schedule_str
# pp schedule

sc = []
schedule.each do |s|
  s['leagueData'].select do |d|
    d['name'] == '챔피언스'
  end.each do |champions|
    sc += champions['list'].sort_by do |m|
      m['order']
    end.map do |m|
      team_a = "#{m['agencyA']} #{m['teamNameA']}".strip
      team_a = TEAM_ALIAS[team_a.to_sym] || team_a
      team_b = "#{m['agencyB']} #{m['teamNameB']}".strip
      team_b = TEAM_ALIAS[team_b.to_sym] || team_b
      {
        date: Date.parse(s['leagueDate']).strftime('%Y-%m-%d'),
        time: m['startTime'],
        home: TEAM_ALIAS[team_a.to_sym] || team_a,
        away: TEAM_ALIAS[team_b.to_sym] || team_b
      }
    end
  end
end

puts sc.to_json

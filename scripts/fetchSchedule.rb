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
}
WEEKDAY_KOR = ['일', '월', '화', '수', '목', '금', '토']

def fetch(uri_str, limit = 10)
  raise ArgumentError, 'too many HTTP redirects' if limit == 0

  response = Net::HTTP.get_response(URI(uri_str))

  case response
  when Net::HTTPSuccess then
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
schedule_str.gsub!(/(\w+\s?:)/) {|key|
  begin
    Time.parse key
    key
  rescue ArgumentError
    "\"#{key[0...-1].strip}\":"
  end
}
schedule_str = schedule_str[1...-1]

schedule = JSON.parse schedule_str
# pp schedule

sc = []
schedule.each do |s|
  s['leagueData'].select {|d|
    d['name'] == '챔피언스'
  }.each do |champions|
    sc += champions['list'].sort_by {|m|
      m['order']
    }.map{|m|
      teamA = "#{m['agencyA']} #{m['teamNameA']}".strip
      teamA = TEAM_ALIAS[teamA.to_sym] || teamA
      teamB = "#{m['agencyB']} #{m['teamNameB']}".strip
      teamB = TEAM_ALIAS[teamB.to_sym] || teamB
      {
        date: Date.parse(s['leagueDate']).strftime('%Y-%m-%d'),
        time: m['startTime'],
        home: TEAM_ALIAS[teamA.to_sym] || teamA,
        away: TEAM_ALIAS[teamB.to_sym] || teamB
      }
    }
  end
end

puts sc.to_json
require 'json'

jsonfile = File.read '../schedules.json'
data = JSON.parse(jsonfile)

puts "Subject,Start Date,Start Time"
data.each do |match|
  puts "#{match['home']} vs #{match['away']},#{match['date']},#{match['time']}"
end

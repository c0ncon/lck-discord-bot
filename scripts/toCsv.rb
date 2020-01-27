require 'json'

filename = ARGV[0]
return if filename.nil?

jsonfile = File.read filename
data = JSON.parse(jsonfile)

puts 'Subject,Start Date,Start Time'
data.each do |match|
  puts "#{match['home']} vs #{match['away']},#{match['date']},#{match['time']}"
end

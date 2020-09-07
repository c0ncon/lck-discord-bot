require 'json'
require 'date'

filename = ARGV[0]
return if filename.nil?

jsonfile = File.read filename
data = JSON.parse(jsonfile)

puts 'BEGIN:VCALENDAR'
puts 'VERSION:2.0'
puts 'PRODID:-//ME//LCK//EN'
puts 'X-WR-CALNAME:LCK'
puts 'X-WR-TIMEZONE:Asia/Seoul'
data.each do |match|
  year, month, day = match['date'].split('-').map(&:to_i)
  hour, minute = match['time'].split(':').map(&:to_i)
  dt = DateTime.new(year, month, day, hour, minute, 0, '+09:00')

  puts 'BEGIN:VEVENT'
  puts "DTSTART: #{dt.new_offset(0).strftime("%Y%m%dT%H%M%SZ")}"
  puts "SUMMARY:#{match['home']} vs #{match['away']}"
  puts 'END:VEVENT'
end

puts 'END:VCALENDAR'
const fs = require('fs');

const args = process.argv.slice(2);
const filename = args[0] || 'full_schedule.json';

const schedule = JSON.parse(fs.readFileSync(filename, 'utf8'));
const events = schedule.map((event) => {
  const date = event.date.replaceAll(/-/g, '');
  const time = event.time.replaceAll(/:/g, '');
  const home = event.home;
  const away = event.away;

  return `BEGIN:VEVENT
DTSTART;TZID=Asia/Seoul:${date}T${time}00
DTEND;TZID=Asia/Seoul:${date}T${time}00
SUMMARY:${home} vs ${away}
END:VEVENT`;
}).join('\n');

const ical = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//ME/LCK//EN
CALSCALE:GREGORIAN
METHOD:PUBLISH
X-WR-CALNAME:LCK
X-WR-TIMEZONE:Asia/Seoul
${events}
END:VCALENDAR`;

fs.writeFileSync('schedule.ics', ical);

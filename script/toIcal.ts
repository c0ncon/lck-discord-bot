import { readFile, writeFile } from 'fs/promises';
import { parseArgs } from 'util';

const {
  values: { schedulePath, outputPath }
} = parseArgs({
  args: Bun.argv.slice(2),
  options: {
    schedulePath: {
      type: 'string',
      default: 'full_schedule.json',
      short: 'p'
    },
    outputPath: {
      type: 'string',
      default: '../schedule.ics',
      short: 'o'
    }
  }
});

interface FileSchedule {
  date: string;
  time: string;
  home: string;
  away: string;
}

if (!schedulePath) {
  throw new Error('No schedule path provided');
}
const schedule: FileSchedule[] = JSON.parse(await readFile(schedulePath, 'utf-8'));
const events = schedule
  .map(({ date, time, home, away }) => {
    return `BEGIN:VEVENT
DTSTART;TZID=Asia/Seoul:${date.replaceAll(/-/g, '')}T${time.replaceAll(/:/g, '')}00
DTEND;TZID=Asia/Seoul:${date.replaceAll(/-/g, '')}T${time.replaceAll(/:/g, '')}00
SUMMARY:${home} vs ${away}
END:VEVENT`;
  })
  .join('\n');

const icalStr = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//ME/LCK//EN
CALSCALE:GREGORIAN
METHOD:PUBLISH
X-WR-CALNAME:LCK
X-WR-TIMEZONE:Asia/Seoul
${events}
END:VCALENDAR`;

if (outputPath) {
  await writeFile(outputPath, icalStr);
}

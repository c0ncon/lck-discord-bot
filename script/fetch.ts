import { readFile, writeFile } from 'fs/promises';
import { parseArgs } from 'util';
import type { Event, ScheduleData } from './event';

interface FileSchedule {
  date: string;
  time: string;
  home: string;
  away: string;
}

const {
  values: { overwrite, overtime, schedulePath, fullSchedulePath }
} = parseArgs({
  args: Bun.argv.slice(2),
  options: {
    overwrite: {
      type: 'boolean',
      default: false,
      short: 'w'
    },
    overtime: {
      type: 'boolean',
      default: false,
      short: 'o'
    },
    schedulePath: {
      type: 'string',
      default: '../schedule.json',
      short: 's'
    },
    fullSchedulePath: {
      type: 'string',
      default: 'full_schedule.json',
      short: 'f'
    }
  }
});

const API_KEY = '0TvQnueqKa5mxJntVWt0w4LpLfEkrV1Ta8rQBb9Z';
const LEAGUE_ID = {
  WORLDS: '98767975604431411',
  LCK: '98767991310872058',
  MSI: '98767991325878492'
};

const leagues = Object.values(LEAGUE_ID).join(',');
const url = `https://esports-api.lolesports.com/persisted/gw/getSchedule?hl=ko-KR&leagueId=${leagues}`;

async function getAllEvents(pageToken?: string, accumulatedEvents: Event[] = []): Promise<Event[]> {
  const response = await fetch(`${url}${pageToken ? `&pageToken=${pageToken}` : ''}`, {
    method: 'GET',
    headers: {
      'x-api-key': API_KEY
    }
  });
  const {
    data: {
      schedule: { events, pages }
    }
  } = (await response.json()) as ScheduleData;

  accumulatedEvents.push(...events);

  if (pages.newer) {
    return getAllEvents(pages.newer, accumulatedEvents);
  }

  return accumulatedEvents;
}

function eventsToSchedule(events: Event[], overtime: boolean = false) {
  const schedule: FileSchedule[] = [];

  for (const event of events) {
    if (!event.match) continue;

    const { startTime } = event;
    const startAt = new Date(startTime);
    const date = overtime
      ? `${startAt.getUTCFullYear()}-${(startAt.getUTCMonth() + 1).toString().padStart(2, '0')}-${startAt.getUTCDate().toString().padStart(2, '0')}`
      : `${startAt.getFullYear()}-${(startAt.getMonth() + 1).toString().padStart(2, '0')}-${startAt.getDate().toString().padStart(2, '0')}`;
    const time = overtime
      ? `${(startAt.getUTCHours() + 9).toString().padStart(2, '0')}:${startAt.getUTCMinutes().toString().padStart(2, '0')}`
      : `${startAt.getHours().toString().padStart(2, '0')}:${startAt.getMinutes().toString().padStart(2, '0')}`;
    const [home, away] = event.match.teams.map(({ code }) => code);

    schedule.push({ date, time, home, away });
  }

  return schedule;
}

async function mergeToFile(newSchedule: FileSchedule[], schedulePath?: string, overwrite: boolean = false) {
  if (!schedulePath) {
    console.log('No schedule path provided\n');
    return;
  }

  const scheduleFile = await readFile(schedulePath, 'utf-8');
  const fileData = JSON.parse(scheduleFile);

  newSchedule.forEach(({ date, time, home, away }) => {
    const idx = fileData.findIndex((event: FileSchedule) => event.date === date && event.time === time);

    if (idx !== -1) {
      if (fileData[idx].home === home && fileData[idx].away === away) return;

      console.log(
        `Schedule updated: ${JSON.stringify(fileData[idx])} -> ${JSON.stringify({ date, time, home, away })}`
      );
      fileData[idx] = { date, time, home, away };
    } else {
      console.log(`New schedule added: ${JSON.stringify({ date, time, home, away })}`);
      fileData.push({ date, time, home, away });
    }
  });

  if (schedulePath && overwrite) {
    console.log(`Overwriting ${schedulePath}`);
    await writeFile(schedulePath, JSON.stringify(fileData, null, 2));
  }
  console.log('Done\n');
}

console.log('Fetching events...');
const events = await getAllEvents();
console.log('Events fetched');
const sc = eventsToSchedule(events, overtime);

if (!overwrite) console.log('Dry run mode. Add -w option to overwrite\n');
console.log(`Merging new events to ${schedulePath}`);
await mergeToFile(sc, schedulePath, overwrite);

console.log(`Merging new events to ${fullSchedulePath}`);
await mergeToFile(sc, fullSchedulePath, overwrite);

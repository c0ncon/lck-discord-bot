const https = require('https');
const fs = require('fs');

const API_KEY = '0TvQnueqKa5mxJntVWt0w4LpLfEkrV1Ta8rQBb9Z'
const LEAGUE_ID = {
  'WORLDS': '98767975604431411',
  'LCK': '98767991310872058',
  'MSI': '98767991325878492',
}

const FULL_SCHEDULE_JSON = './full_schedule.json';
const CURRENT_SCHEDULE_JSON = '../schedule.json';

const leagueId = Object.values(LEAGUE_ID).join(',');
const options = {
  host: 'esports-api.lolesports.com',
  path: `/persisted/gw/getSchedule?hl=ko-KR&leagueId=${leagueId}`,
  headers: {
    'x-api-key': API_KEY
  },
  method: 'GET'
};

(async () => {
  let events = [];
  let resp = await getSchedule();
  events = events.concat(resp.data.schedule.events);

  while (!Object.is(resp.data.schedule.pages.newer, null)) {
    resp = await getSchedule(resp.data.schedule.pages.newer);
    events = events.concat(resp.data.schedule.events);
  }

  const schedule = events.filter((event) => event.state === 'unstarted').map((event) => {
    const startTime = new Date(event.startTime);
    const date = `${startTime.getFullYear()}-${(startTime.getMonth() + 1).toString().padStart(2, '0')}-${startTime.getDate().toString().padStart(2, '0')}`;
    const time = `${startTime.getHours().toString().padStart(2, '0')}:${startTime.getMinutes().toString().padStart(2, '0')}`;
    // const date = `${startTime.getUTCFullYear()}-${(startTime.getUTCMonth() + 1).toString().padStart(2, '0')}-${startTime.getUTCDate().toString().padStart(2, '0')}`;
    // const time = `${(startTime.getUTCHours() + 9).toString().padStart(2, '0')}:${startTime.getUTCMinutes().toString().padStart(2, '0')}`;
    const home = event.match.teams[0].code;
    const away = event.match.teams[1].code;

    return { date, time, home, away };
  });

  let overwrite = false;
  overwrite = (process.argv.length > 2 && process.argv[2] === 'w') ? true : false;

  if (overwrite) {
    console.log('Overwrite mode\n');
  } else {
    console.log('Dry run mode\n');
  }
  mergeSchedule(schedule, CURRENT_SCHEDULE_JSON, overwrite);
  mergeSchedule(schedule, FULL_SCHEDULE_JSON, overwrite);
})();

function getSchedule(pageToken = null) {
  return new Promise((resolve, reject) => {
    const req = https.request({ ...options, path: Object.is(pageToken, null) ? options.path : options.path + `&pageToken=${pageToken}` }, (res) => {
      let str = '';
      res.on('data', (chunk) => {
        str += chunk;
      });

      res.on('end', () => {
        resolve(JSON.parse(str));
      });
    });

    req.on('error', (err) => {
      reject(err);
    });

    req.end();
  });
}

function mergeSchedule(newSchedule, srcJson, overwrite = false) {
  const scheduleJson = fs.readFileSync(srcJson, 'utf8');
  const origSchedule = JSON.parse(scheduleJson);

  console.log('Merging ', srcJson);
  newSchedule.forEach((event) => {
    let idx = origSchedule.findIndex((event2) => {
      return (event.date == event2.date && event.time == event2.time);
    });

    if (idx === -1) {
      console.log('added', JSON.stringify(event));
      origSchedule.push(event);
    } else {
      const event2 = origSchedule[idx];
      if (event.home != event2.home || event.away != event2.away) {
        console.log('updated', `${JSON.stringify(origSchedule[idx])} -> ${JSON.stringify(event)}`);
        origSchedule[idx] = event;
      }
    }
  });
  console.log('\n');

  if (overwrite) {
    fs.writeFileSync(srcJson, JSON.stringify(origSchedule, null, 2));
  }
}

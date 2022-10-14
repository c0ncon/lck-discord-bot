const https = require('https');

const API_KEY = '0TvQnueqKa5mxJntVWt0w4LpLfEkrV1Ta8rQBb9Z'
const LEAGUE_ID = {
  'WORLDS': '98767975604431411',
  'LCK': '98767991310872058',
  'MSI': '98767991325878492',
}

const leagueId = Object.values(LEAGUE_ID).join(',');
const options = {
  host: 'esports-api.lolesports.com',
  path: `/persisted/gw/getSchedule?hl=ko-KR&leagueId=${leagueId}`,
  headers: {
    'x-api-key': API_KEY
  },
  method: 'GET'
};

https.request(options, (res) => {
  let str = '';
  res.on('data', (chunk) => {
    str += chunk;
  });

  res.on('end', () => {
    const body = JSON.parse(str);
    const events = body.data.schedule.events;

    const schedule = events.filter((event) => event.state === 'unstarted').map((event) => {
      const startTime = new Date(event.startTime);
      const date = `${startTime.getFullYear()}-${(startTime.getMonth() + 1).toString().padStart(2, '0')}-${startTime.getDate().toString().padStart(2, '0')}`;
      const time = `${startTime.getHours().toString().padStart(2, '0')}:${startTime.getMinutes().toString().padStart(2, '0')}`;
      const home = event.match.teams[0].code;
      const away = event.match.teams[1].code; 

      return { date, time, home, away };
    });

    console.log(JSON.stringify(schedule));
  });
}).end();

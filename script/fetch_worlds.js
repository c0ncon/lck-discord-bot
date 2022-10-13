const https = require('https');

const options = {
  host: 'esports-api.lolesports.com',
  path: '/persisted/gw/getSchedule?hl=ko-KR&leagueId=98767975604431411',
  headers: {
    'x-api-key': '0TvQnueqKa5mxJntVWt0w4LpLfEkrV1Ta8rQBb9Z'
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

      return {
        date, time,
        home: event.match.teams[0].code,
        away: event.match.teams[1].code
      }
    });

    console.log(JSON.stringify(schedule));
  });
}).end();

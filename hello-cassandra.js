console.log("Hello Cassandra!");

const cassandra = require('cassandra-driver');
const client = new cassandra.Client({ contactPoints: ['cassandra:9042'], localDataCenter: 'dc1'});

const query = 'SELECT * FROM videodb.videos';
client.execute(query)
  .then(result => {
    const row = result.first();
    console.log('User lives in %s, %s', row.videoname, row.description); 
  });

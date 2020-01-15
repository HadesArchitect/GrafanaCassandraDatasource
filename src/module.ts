import {CassandraDatasource} from './datasource';
import {CassandraDatasourceQueryCtrl} from './query_ctrl';

class CassandraConfigCtrl {
  static templateUrl = 'partials/config.html';
  current: any;
}

class CassandraQueryOptionsCtrl {
  static templateUrl = 'partials/query.options.html';
}

export {
  CassandraDatasource as Datasource,
  CassandraConfigCtrl as ConfigCtrl,
  CassandraDatasourceQueryCtrl as QueryCtrl,
  CassandraQueryOptionsCtrl as QueryOptionsCtrl
};

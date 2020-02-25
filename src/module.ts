import {CassandraDatasource} from './datasource';
import {CassandraQueryCtrl} from './query_ctrl';
import {CassandraConfigCtrl} from './config_ctrl';

class CassandraQueryOptionsCtrl {
  static templateUrl = 'partials/query.options.html';
}

export {
  CassandraDatasource as Datasource,
  CassandraConfigCtrl as ConfigCtrl,
  CassandraQueryCtrl as QueryCtrl,
  CassandraQueryOptionsCtrl as QueryOptionsCtrl
};

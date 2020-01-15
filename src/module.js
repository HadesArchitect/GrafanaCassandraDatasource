import {CassandraDatasource} from './datasource';
import {CassandraDatasourceQueryCtrl} from './query_ctrl';

class CassandraConfigCtrl {}
CassandraConfigCtrl.templateUrl = 'partials/config.html';

export {
  CassandraDatasource as Datasource,
  CassandraConfigCtrl as ConfigCtrl,
};

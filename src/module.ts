import {CassandraDatasource} from './datasource';

class CassandraConfigCtrl {
  static templateUrl = 'partials/config.html';
  current: any;
}

export {
  CassandraDatasource as Datasource,
  CassandraConfigCtrl as ConfigCtrl,
};

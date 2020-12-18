import {QueryCtrl} from 'grafana/app/plugins/sdk';
import { CassandraDatasource } from 'datasource';
import {TableMetadata} from './models';

export class CassandraQueryCtrl extends QueryCtrl {
  static templateUrl = 'partials/query.editor.html';

  datatsource: CassandraDatasource;
  scope: any;
  hasRawMode: false;

  /** @ngInject */
  constructor($scope, $injector) {
    super($scope, $injector);

    this.scope = $scope;
    this.target.target = this.target.target || 'select timestamp, value from keyspace.table where id=123e4567;';
    this.target.type = this.target.type || 'timeserie';
    this.target.columnTime = this.target.columnTime || ' ';
    this.target.columnValue = this.target.columnValue || ' ';
    this.target.columnId = this.target.columnId || ' ';

    // TODO if keyspace and table are set load column suggestions
  }

  getOptions(keyspace: string, table: string, type: string) {
    if (!keyspace || !table) {
      return Promise.resolve([]);
    }

    return this.datasource.metricFindQuery(keyspace, table).then(tmd => {
      return tmd.toSuggestion();
    });
  }

  toggleEditorMode(): void {
    this.target.rawQuery = !this.target.rawQuery;
  }

  onChangeInternal(): void {
    this.panelCtrl.refresh();
  }

  exampleMethod(a1: number, a2: number) {
    return a1 + a2;
  }

}

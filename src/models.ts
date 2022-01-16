import { DataQuery } from '@grafana/data';

export interface TSDBRequest {
  queries: CassandraQuery[];
  from?: string;
  to?: string;
}

export interface CassandraQuery extends DataQuery {
  target?: any;
  queryType: CassandraQueryType;
  filtering?: boolean;
  keyspace?: string;
  table?: string;
  columnTime?: string;
  columnValue?: string;
  columnId?: string;
  valueId?: string;
  hide?: boolean;
  rawQuery?: boolean;
}

type CassandraQueryType = 'query' | 'search' | 'connection';

export interface TSDBRequestOptions {
  range?: {
    from: any;
    to: any;
  };
  targets: CassandraQuery[];
}

export class TableMetadata {
  columns: ColumnMetadata[];

  constructor(rawJson?: string) {
    this.columns = [];
    if (rawJson) {
      for (const column of JSON.parse(rawJson)) {
        this.columns.push(new ColumnMetadata(column['Name'], column['Type']));
      }
    }
  }

  toSuggestion(): Record<string, any> {
    const suggestions: Array<Record<string, any>> = [];
    for (const column of this.columns) {
      suggestions.push(column.toSuggestion());
    }
    return suggestions;
  }
}

class ColumnMetadata {
  name: string;
  type: string;

  constructor(name: string, type: string) {
    this.name = name;
    this.type = type;
  }

  toSuggestion(): Record<string, any> {
    return {
      text: this.name,
      value: this.name,
    };
  }
}

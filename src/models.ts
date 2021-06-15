export interface TSDBRequest {
    queries: TSDBQuery[];
    from?: string;
    to?: string;
}
  
export interface TSDBQuery {
    datasourceId: string;
    target?: any;
    queryType: TSDBQueryType;
    refId?: string;
    filtering?: boolean;
    keyspace?: string;
    table?: string,
    type?: TSDBDataType;
    columnTime?: string;
    columnValue?: string;
    columnId?: string;
    valueId?: string;
    hide?: boolean;
    rawQuery?: boolean;
}
  
type TSDBQueryType = 'query' | 'search' | 'connection';
type TSDBDataType = 'timeserie' | 'table';

export interface TSDBRequestOptions {
    range?: {
      from: any;
      to: any;
    };
    targets: TSDBQuery[];
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
        const suggestions : Record<string, any>[] = [];
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
            value: this.name
        };
    }
}

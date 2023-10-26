package db_client

// TODO KAI does not seem to be used

// // ColumnTypeDatabaseTypeName returns the database system type name. If the name is unknown the OID is returned.
// func columnTypeDatabaseTypeName(field pgconn.FieldDescription, connection *pgx.Conn) (typeName string) {
// 	if dt, ok := connection.TypeMap().TypeForOID(field.DataTypeOID); ok {
// 		return strings.ToUpper(dt.Name)
// 	}

// 	return strconv.FormatInt(int64(field.DataTypeOID), 10)
// }

// func fieldDescriptionsToColumns(fieldDescriptions []pgconn.FieldDescription, connection *pgx.Conn) []*queryresult.ColumnDef {
// 	cols := make([]*queryresult.ColumnDef, len(fieldDescriptions))

// 	for i, f := range fieldDescriptions {
// 		typeName := columnTypeDatabaseTypeName(f, connection)

// 		cols[i] = &queryresult.ColumnDef{
// 			Name:     string(f.Name),
// 			DataType: typeName,
// 		}
// 	}
// 	return cols
// }

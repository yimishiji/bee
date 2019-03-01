// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package generate

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	beeLogger "github.com/yimishiji/bee/logger"
	"github.com/yimishiji/bee/logger/colors"
	strings2 "github.com/yimishiji/bee/pkg/strings"
	"github.com/yimishiji/bee/utils"
)

const (
	OModel byte = 1 << iota
	OController
	ORouter
	OVue
)

// DbTransformer has method to reverse engineer a database schema to restful api code
type DbTransformer interface {
	GetTableNames(conn *sql.DB) []string
	GetConstraints(conn *sql.DB, table *Table, blackList map[string]bool)
	GetColumns(conn *sql.DB, table *Table, blackList map[string]bool)
	GetGoDataType(sqlType string) (string, error)
}

// MysqlDB is the MySQL version of DbTransformer
type MysqlDB struct {
}

// PostgresDB is the PostgreSQL version of DbTransformer
type PostgresDB struct {
}

// dbDriver maps a DBMS name to its version of DbTransformer
var dbDriver = map[string]DbTransformer{
	"mysql":    &MysqlDB{},
	"postgres": &PostgresDB{},
}

type MvcPath struct {
	ModelPath      string
	ControllerPath string
	RouterPath     string
	VuePath        string
	FilterPath     string
}

// typeMapping maps SQL data type to corresponding Go data type
var typeMappingMysql = map[string]string{
	"int":                "int", // int signed
	"integer":            "int",
	"tinyint":            "int8",
	"smallint":           "int16",
	"mediumint":          "int32",
	"bigint":             "int64",
	"int unsigned":       "uint", // int unsigned
	"integer unsigned":   "uint",
	"tinyint unsigned":   "uint8",
	"smallint unsigned":  "uint16",
	"mediumint unsigned": "uint32",
	"bigint unsigned":    "uint64",
	"bit":                "uint64",
	"bool":               "bool",   // boolean
	"enum":               "string", // enum
	"set":                "string", // set
	"varchar":            "string", // string & text
	"char":               "string",
	"tinytext":           "string",
	"mediumtext":         "string",
	"text":               "string",
	"longtext":           "string",
	"blob":               "string", // blob
	"tinyblob":           "string",
	"mediumblob":         "string",
	"longblob":           "string",
	"date":               "time.Time", // time
	"datetime":           "time.Time",
	"timestamp":          "time.Time",
	"time":               "time.Time",
	"float":              "float32", // float & decimal
	"double":             "float64",
	"decimal":            "float64",
	"binary":             "string", // binary
	"varbinary":          "string",
	"year":               "int16",
	"json":               "string", // json
}

// typeMappingPostgres maps SQL data type to corresponding Go data type
var typeMappingPostgres = map[string]string{
	"serial":                      "int", // serial
	"big serial":                  "int64",
	"smallint":                    "int16", // int
	"integer":                     "int",
	"bigint":                      "int64",
	"boolean":                     "bool",   // bool
	"char":                        "string", // string
	"character":                   "string",
	"character varying":           "string",
	"varchar":                     "string",
	"text":                        "string",
	"date":                        "time.Time", // time
	"time":                        "time.Time",
	"timestamp":                   "time.Time",
	"timestamp without time zone": "time.Time",
	"timestamp with time zone":    "time.Time",
	"interval":                    "string",  // time interval, string for now
	"real":                        "float32", // float & decimal
	"double precision":            "float64",
	"decimal":                     "float64",
	"numeric":                     "float64",
	"money":                       "float64", // money
	"bytea":                       "string",  // binary
	"tsvector":                    "string",  // fulltext
	"ARRAY":                       "string",  // array
	"USER-DEFINED":                "string",  // user defined
	"uuid":                        "string",  // uuid
	"json":                        "string",  // json
	"jsonb":                       "string",  // jsonb
	"inet":                        "string",  // ip address
}

// Table represent a table in a database
type Table struct {
	Name          string
	Pk            string
	Uk            []string
	Fk            map[string]*ForeignKey
	Columns       []*Column
	ImportTimePkg bool
}

// Column reprsents a column for a table
type Column struct {
	Name string
	Type string
	Tag  *OrmTag
}

// ForeignKey represents a foreign key column for a table
type ForeignKey struct {
	Name      string
	RefSchema string
	RefTable  string
	RefColumn string
}

// OrmTag contains Beego ORM tag information for a column
type OrmTag struct {
	Auto        bool
	Pk          bool
	Null        bool
	Index       bool
	Unique      bool
	Column      string
	Size        string
	Decimals    string
	Digits      string
	AutoNow     bool
	AutoNowAdd  bool
	Type        string
	Default     string
	RelOne      bool
	ReverseOne  bool
	RelFk       bool
	ReverseMany bool
	RelM2M      bool
	Comment     string //column comment
}

// String returns the source code string for the Table struct
func (tb *Table) String() string {
	rv := fmt.Sprintf("type %s struct {\n", utils.CamelCase(tb.Name))
	for _, v := range tb.Columns {
		rv += v.String() + "\n"
	}
	rv += "}\n"
	return rv
}

// String returns the source code string of a field in Table struct
// It maps to a column in database table. e.g. Id int `orm:"column(id);auto"`
func (col *Column) String() string {
	return fmt.Sprintf("%s %s %s", col.Name, col.Type, col.Tag.String())
}

// String returns the ORM tag string for a column
func (tag *OrmTag) String() string {
	var ormOptions []string
	if tag.Column != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("column:%s", tag.Column))
	}
	if tag.Auto {
		ormOptions = append(ormOptions, "auto")
	}
	if tag.Size != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("size:%s", tag.Size))
	}
	if tag.Type != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("type:%s", tag.Type))
	}
	if tag.Null {
		ormOptions = append(ormOptions, "null")
	}
	if tag.AutoNow {
		ormOptions = append(ormOptions, "auto_now")
	}
	if tag.AutoNowAdd {
		ormOptions = append(ormOptions, "auto_now_add")
	}
	if tag.Decimals != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("digits:%s;decimals:%s", tag.Digits, tag.Decimals))
	}
	if tag.RelFk {
		ormOptions = append(ormOptions, "rel:fk")
	}
	if tag.RelOne {
		ormOptions = append(ormOptions, "rel:one")
	}
	if tag.ReverseOne {
		ormOptions = append(ormOptions, "reverse:one")
	}
	if tag.ReverseMany {
		ormOptions = append(ormOptions, "reverse:many")
	}
	if tag.RelM2M {
		ormOptions = append(ormOptions, "rel:m2m")
	}
	if tag.Pk {
		ormOptions = append(ormOptions, "pk")
	}
	if tag.Unique {
		ormOptions = append(ormOptions, "unique")
	}
	if tag.Default != "" {
		ormOptions = append(ormOptions, fmt.Sprintf("default:%s", tag.Default))
	}
	if len(ormOptions) == 0 {
		return ""
	}
	if tag.Comment != "" {
		return fmt.Sprintf("`json:\"%s\" gorm:\"%s\" description:\"%s\"`", tag.Column, strings.Join(ormOptions, ";"), tag.Comment)
	}
	return fmt.Sprintf("`json:\"%s\" gorm:\"%s\"`", tag.Column, strings.Join(ormOptions, ";"))
}

func GenerateAppcode(driver, connStr, level, tables, currpath string) {
	var mode byte
	switch level {
	case "1":
		mode = OModel
	case "2":
		mode = OModel | OController
	case "3":
		mode = OModel | OController | ORouter
	case "4":
		mode = OModel | OController | ORouter | OVue
	default:
		beeLogger.Log.Fatal("Invalid level value. Must be either \"1\", \"2\", or \"3\"")
	}
	var selectedTables map[string]bool
	if tables != "" {
		selectedTables = make(map[string]bool)
		for _, v := range strings.Split(tables, ",") {
			selectedTables[v] = true
		}
	}
	switch driver {
	case "mysql":
	case "postgres":
	case "sqlite":
		beeLogger.Log.Fatal("Generating app code from SQLite database is not supported yet.")
	default:
		beeLogger.Log.Fatal("Unknown database driver. Must be either \"mysql\", \"postgres\" or \"sqlite\"")
	}
	gen(driver, connStr, mode, selectedTables, currpath)
}

// Generate takes table, column and foreign key information from database connection
// and generate corresponding golang source files
func gen(dbms, connStr string, mode byte, selectedTableNames map[string]bool, apppath string) {
	db, err := sql.Open(dbms, connStr)
	if err != nil {
		beeLogger.Log.Fatalf("Could not connect to '%s' database using '%s': %s", dbms, connStr, err)
	}
	defer db.Close()
	if trans, ok := dbDriver[dbms]; ok {
		beeLogger.Log.Info("Analyzing database tables...")
		var tableNames []string
		if len(selectedTableNames) != 0 {
			for tableName := range selectedTableNames {
				tableNames = append(tableNames, tableName)
			}
		} else {
			tableNames = trans.GetTableNames(db)
		}
		tables := getTableObjects(tableNames, db, trans)
		mvcPath := new(MvcPath)
		mvcPath.ModelPath = path.Join(apppath, "models")
		mvcPath.ControllerPath = path.Join(apppath, "controllers")
		mvcPath.RouterPath = path.Join(apppath, "routers")
		mvcPath.VuePath = path.Join(apppath, "vue/src/components")
		mvcPath.FilterPath = path.Join(apppath, "filters")

		createPaths(mode, mvcPath)
		pkgPath := getPackagePath(apppath)
		writeSourceFiles(pkgPath, tables, mode, mvcPath)
	} else {
		beeLogger.Log.Fatalf("Generating app code from '%s' database is not supported yet.", dbms)
	}
}

// GetTableNames returns a slice of table names in the current database
func (*MysqlDB) GetTableNames(db *sql.DB) (tables []string) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		beeLogger.Log.Fatalf("Could not show tables: %s", err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			beeLogger.Log.Fatalf("Could not show tables: %s", err)
		}
		tables = append(tables, name)
	}
	return
}

// getTableObjects process each table name
func getTableObjects(tableNames []string, db *sql.DB, dbTransformer DbTransformer) (tables []*Table) {
	// if a table has a composite pk or doesn't have pk, we can't use it yet
	// these tables will be put into blacklist so that other struct will not
	// reference it.
	blackList := make(map[string]bool)
	// process constraints information for each table, also gather blacklisted table names
	for _, tableName := range tableNames {
		// create a table struct
		tb := new(Table)
		tb.Name = tableName
		tb.Fk = make(map[string]*ForeignKey)
		dbTransformer.GetConstraints(db, tb, blackList)
		tables = append(tables, tb)
	}
	// process columns, ignoring blacklisted tables
	for _, tb := range tables {
		dbTransformer.GetColumns(db, tb, blackList)
	}
	return
}

// GetConstraints gets primary key, unique key and foreign keys of a table from
// information_schema and fill in the Table struct
func (*MysqlDB) GetConstraints(db *sql.DB, table *Table, blackList map[string]bool) {
	rows, err := db.Query(
		`SELECT
			c.constraint_type, u.column_name, u.referenced_table_schema, u.referenced_table_name, referenced_column_name, u.ordinal_position
		FROM
			information_schema.table_constraints c
		INNER JOIN
			information_schema.key_column_usage u ON c.constraint_name = u.constraint_name
		WHERE
			c.table_schema = database() AND c.table_name = ? AND u.table_schema = database() AND u.table_name = ?`,
		table.Name, table.Name) //  u.position_in_unique_constraint,
	if err != nil {
		beeLogger.Log.Fatal("Could not query INFORMATION_SCHEMA for PK/UK/FK information")
	}
	for rows.Next() {
		var constraintTypeBytes, columnNameBytes, refTableSchemaBytes, refTableNameBytes, refColumnNameBytes, refOrdinalPosBytes []byte
		if err := rows.Scan(&constraintTypeBytes, &columnNameBytes, &refTableSchemaBytes, &refTableNameBytes, &refColumnNameBytes, &refOrdinalPosBytes); err != nil {
			beeLogger.Log.Fatal("Could not read INFORMATION_SCHEMA for PK/UK/FK information")
		}
		constraintType, columnName, refTableSchema, refTableName, refColumnName, refOrdinalPos :=
			string(constraintTypeBytes), string(columnNameBytes), string(refTableSchemaBytes),
			string(refTableNameBytes), string(refColumnNameBytes), string(refOrdinalPosBytes)
		if constraintType == "PRIMARY KEY" {
			if refOrdinalPos == "1" {
				table.Pk = columnName
			} else {
				table.Pk = ""
				// Add table to blacklist so that other struct will not reference it, because we are not
				// registering blacklisted tables
				blackList[table.Name] = true
			}
		} else if constraintType == "UNIQUE" {
			table.Uk = append(table.Uk, columnName)
		} else if constraintType == "FOREIGN KEY" {
			fk := new(ForeignKey)
			fk.Name = columnName
			fk.RefSchema = refTableSchema
			fk.RefTable = refTableName
			fk.RefColumn = refColumnName
			table.Fk[columnName] = fk
		}
	}
}

// GetColumns retrieves columns details from
// information_schema and fill in the Column struct
func (mysqlDB *MysqlDB) GetColumns(db *sql.DB, table *Table, blackList map[string]bool) {
	// retrieve columns
	colDefRows, err := db.Query(
		`SELECT
			column_name, data_type, column_type, is_nullable, column_default, extra, column_comment 
		FROM
			information_schema.columns
		WHERE
			table_schema = database() AND table_name = ?`,
		table.Name)
	if err != nil {
		beeLogger.Log.Fatalf("Could not query the database: %s", err)
	}
	defer colDefRows.Close()

	for colDefRows.Next() {
		// datatype as bytes so that SQL <null> values can be retrieved
		var colNameBytes, dataTypeBytes, columnTypeBytes, isNullableBytes, columnDefaultBytes, extraBytes, columnCommentBytes []byte
		if err := colDefRows.Scan(&colNameBytes, &dataTypeBytes, &columnTypeBytes, &isNullableBytes, &columnDefaultBytes, &extraBytes, &columnCommentBytes); err != nil {
			beeLogger.Log.Fatal("Could not query INFORMATION_SCHEMA for column information")
		}
		colName, dataType, columnType, isNullable, columnDefault, extra, columnComment :=
			string(colNameBytes), string(dataTypeBytes), string(columnTypeBytes), string(isNullableBytes), string(columnDefaultBytes), string(extraBytes), string(columnCommentBytes)

		// create a column
		col := new(Column)
		col.Name = utils.CamelCase(colName)
		col.Type, err = mysqlDB.GetGoDataType(dataType)
		if err != nil {
			beeLogger.Log.Fatalf("%s", err)
		}

		// Tag info
		tag := new(OrmTag)
		tag.Column = colName
		tag.Comment = columnComment
		if table.Pk == colName {
			col.Name = "Id"
			col.Type = "int"
			if extra == "auto_increment" {
				tag.Auto = true
			} else {
				tag.Pk = true
			}
		} else {
			fkCol, isFk := table.Fk[colName]
			isBl := false
			if isFk {
				_, isBl = blackList[fkCol.RefTable]
			}
			// check if the current column is a foreign key
			if isFk && !isBl {
				tag.RelFk = true
				refStructName := fkCol.RefTable
				col.Name = utils.CamelCase(colName)
				col.Type = "*" + utils.CamelCase(refStructName)
			} else {
				// if the name of column is Id, and it's not primary key
				if colName == "id" {
					col.Name = "Id_RENAME"
				}
				if isNullable == "YES" {
					tag.Null = true
				}
				if isSQLSignedIntType(dataType) {
					sign := extractIntSignness(columnType)
					if sign == "unsigned" && extra != "auto_increment" {
						col.Type, err = mysqlDB.GetGoDataType(dataType + " " + sign)
						if err != nil {
							beeLogger.Log.Fatalf("%s", err)
						}
					}
				}
				if isSQLStringType(dataType) {
					tag.Size = extractColSize(columnType)
				}
				if isSQLTemporalType(dataType) {
					tag.Type = dataType
					//check auto_now, auto_now_add
					if columnDefault == "CURRENT_TIMESTAMP" && extra == "on update CURRENT_TIMESTAMP" {
						tag.AutoNow = true
					} else if columnDefault == "CURRENT_TIMESTAMP" {
						tag.AutoNowAdd = true
					}
					// need to import time package
					table.ImportTimePkg = true
				}
				if isSQLDecimal(dataType) {
					tag.Digits, tag.Decimals = extractDecimal(columnType)
				}
				if isSQLBinaryType(dataType) {
					tag.Size = extractColSize(columnType)
				}
				if isSQLBitType(dataType) {
					tag.Size = extractColSize(columnType)
				}
			}
		}
		col.Tag = tag
		table.Columns = append(table.Columns, col)
	}
}

// GetGoDataType maps an SQL data type to Golang data type
func (*MysqlDB) GetGoDataType(sqlType string) (string, error) {
	if v, ok := typeMappingMysql[sqlType]; ok {
		return v, nil
	}
	return "", fmt.Errorf("data type '%s' not found", sqlType)
}

// GetTableNames for PostgreSQL
func (*PostgresDB) GetTableNames(db *sql.DB) (tables []string) {
	rows, err := db.Query(`
		SELECT table_name FROM information_schema.tables
		WHERE table_catalog = current_database() AND
		table_type = 'BASE TABLE' AND
		table_schema NOT IN ('pg_catalog', 'information_schema')`)
	if err != nil {
		beeLogger.Log.Fatalf("Could not show tables: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			beeLogger.Log.Fatalf("Could not show tables: %s", err)
		}
		tables = append(tables, name)
	}
	return
}

// GetConstraints for PostgreSQL
func (*PostgresDB) GetConstraints(db *sql.DB, table *Table, blackList map[string]bool) {
	rows, err := db.Query(
		`SELECT
			c.constraint_type,
			u.column_name,
			cu.table_catalog AS referenced_table_catalog,
			cu.table_name AS referenced_table_name,
			cu.column_name AS referenced_column_name,
			u.ordinal_position
		FROM
			information_schema.table_constraints c
		INNER JOIN
			information_schema.key_column_usage u ON c.constraint_name = u.constraint_name
		INNER JOIN
			information_schema.constraint_column_usage cu ON cu.constraint_name =  c.constraint_name
		WHERE
			c.table_catalog = current_database() AND c.table_schema NOT IN ('pg_catalog', 'information_schema')
			 AND c.table_name = $1
			AND u.table_catalog = current_database() AND u.table_schema NOT IN ('pg_catalog', 'information_schema')
			 AND u.table_name = $2`,
		table.Name, table.Name) //  u.position_in_unique_constraint,
	if err != nil {
		beeLogger.Log.Fatalf("Could not query INFORMATION_SCHEMA for PK/UK/FK information: %s", err)
	}

	for rows.Next() {
		var constraintTypeBytes, columnNameBytes, refTableSchemaBytes, refTableNameBytes, refColumnNameBytes, refOrdinalPosBytes []byte
		if err := rows.Scan(&constraintTypeBytes, &columnNameBytes, &refTableSchemaBytes, &refTableNameBytes, &refColumnNameBytes, &refOrdinalPosBytes); err != nil {
			beeLogger.Log.Fatalf("Could not read INFORMATION_SCHEMA for PK/UK/FK information: %s", err)
		}
		constraintType, columnName, refTableSchema, refTableName, refColumnName, refOrdinalPos :=
			string(constraintTypeBytes), string(columnNameBytes), string(refTableSchemaBytes),
			string(refTableNameBytes), string(refColumnNameBytes), string(refOrdinalPosBytes)
		if constraintType == "PRIMARY KEY" {
			if refOrdinalPos == "1" {
				table.Pk = columnName
			} else {
				table.Pk = ""
				// add table to blacklist so that other struct will not reference it, because we are not
				// registering blacklisted tables
				blackList[table.Name] = true
			}
		} else if constraintType == "UNIQUE" {
			table.Uk = append(table.Uk, columnName)
		} else if constraintType == "FOREIGN KEY" {
			fk := new(ForeignKey)
			fk.Name = columnName
			fk.RefSchema = refTableSchema
			fk.RefTable = refTableName
			fk.RefColumn = refColumnName
			table.Fk[columnName] = fk
		}
	}
}

// GetColumns for PostgreSQL
func (postgresDB *PostgresDB) GetColumns(db *sql.DB, table *Table, blackList map[string]bool) {
	// retrieve columns
	colDefRows, err := db.Query(
		`SELECT
			column_name,
			data_type,
			data_type ||
			CASE
				WHEN data_type = 'character' THEN '('||character_maximum_length||')'
				WHEN data_type = 'numeric' THEN '(' || numeric_precision || ',' || numeric_scale ||')'
				ELSE ''
			END AS column_type,
			is_nullable,
			column_default,
			'' AS extra
		FROM
			information_schema.columns
		WHERE
			table_catalog = current_database() AND table_schema NOT IN ('pg_catalog', 'information_schema')
			 AND table_name = $1`,
		table.Name)
	if err != nil {
		beeLogger.Log.Fatalf("Could not query INFORMATION_SCHEMA for column information: %s", err)
	}
	defer colDefRows.Close()

	for colDefRows.Next() {
		// datatype as bytes so that SQL <null> values can be retrieved
		var colNameBytes, dataTypeBytes, columnTypeBytes, isNullableBytes, columnDefaultBytes, extraBytes []byte
		if err := colDefRows.Scan(&colNameBytes, &dataTypeBytes, &columnTypeBytes, &isNullableBytes, &columnDefaultBytes, &extraBytes); err != nil {
			beeLogger.Log.Fatalf("Could not query INFORMATION_SCHEMA for column information: %s", err)
		}
		colName, dataType, columnType, isNullable, columnDefault, extra :=
			string(colNameBytes), string(dataTypeBytes), string(columnTypeBytes), string(isNullableBytes), string(columnDefaultBytes), string(extraBytes)
		// Create a column
		col := new(Column)
		col.Name = utils.CamelCase(colName)
		col.Type, err = postgresDB.GetGoDataType(dataType)
		if err != nil {
			beeLogger.Log.Fatalf("%s", err)
		}

		// Tag info
		tag := new(OrmTag)
		tag.Column = colName
		if table.Pk == colName {
			col.Name = "Id"
			col.Type = "int"
			if extra == "auto_increment" {
				tag.Auto = true
			} else {
				tag.Pk = true
			}
		} else {
			fkCol, isFk := table.Fk[colName]
			isBl := false
			if isFk {
				_, isBl = blackList[fkCol.RefTable]
			}
			// check if the current column is a foreign key
			if isFk && !isBl {
				tag.RelFk = true
				refStructName := fkCol.RefTable
				col.Name = utils.CamelCase(colName)
				col.Type = "*" + utils.CamelCase(refStructName)
			} else {
				// if the name of column is Id, and it's not primary key
				if colName == "id" {
					col.Name = "Id_RENAME"
				}
				if isNullable == "YES" {
					tag.Null = true
				}
				if isSQLStringType(dataType) {
					tag.Size = extractColSize(columnType)
				}
				if isSQLTemporalType(dataType) || strings.HasPrefix(dataType, "timestamp") {
					tag.Type = dataType
					//check auto_now, auto_now_add
					if columnDefault == "CURRENT_TIMESTAMP" && extra == "on update CURRENT_TIMESTAMP" {
						tag.AutoNow = true
					} else if columnDefault == "CURRENT_TIMESTAMP" {
						tag.AutoNowAdd = true
					}
					// need to import time package
					table.ImportTimePkg = true
				}
				if isSQLDecimal(dataType) {
					tag.Digits, tag.Decimals = extractDecimal(columnType)
				}
				if isSQLBinaryType(dataType) {
					tag.Size = extractColSize(columnType)
				}
				if isSQLStrangeType(dataType) {
					tag.Type = dataType
				}
			}
		}
		col.Tag = tag
		table.Columns = append(table.Columns, col)
	}
}

// GetGoDataType returns the Go type from the mapped Postgres type
func (*PostgresDB) GetGoDataType(sqlType string) (string, error) {
	if v, ok := typeMappingPostgres[sqlType]; ok {
		return v, nil
	}
	return "", fmt.Errorf("data type '%s' not found", sqlType)
}

// deleteAndRecreatePaths removes several directories completely
func createPaths(mode byte, paths *MvcPath) {
	if (mode & OModel) == OModel {
		os.Mkdir(paths.ModelPath, 0777)
	}
	if (mode & OController) == OController {
		os.Mkdir(paths.ControllerPath, 0777)
		os.Mkdir(paths.FilterPath, 0777)
	}
	if (mode & ORouter) == ORouter {
		os.Mkdir(paths.RouterPath, 0777)
	}
	if (mode & OVue) == OVue {
		os.Mkdir(paths.VuePath, 0777)
	}
}

var notirceMsgArr []string

// writeSourceFiles generates source files for model/controller/router
// It will wipe the following directories and recreate them:./models, ./controllers, ./routers
// Newly geneated files will be inside these folders.
func writeSourceFiles(pkgPath string, tables []*Table, mode byte, paths *MvcPath) {
	if (OModel & mode) == OModel {
		beeLogger.Log.Info("Creating model files...")
		writeModelFiles(tables, paths.ModelPath)
	}
	if (OController & mode) == OController {
		beeLogger.Log.Info("Creating controller files...")
		writeControllerFiles(tables, paths.ControllerPath, pkgPath)

		beeLogger.Log.Info("Creating controller files...")
		writeFilterFiles(tables, paths.FilterPath, pkgPath)

	}
	if (ORouter & mode) == ORouter {
		beeLogger.Log.Info("Creating router files...")
		writeRouterFile(tables, paths.RouterPath, pkgPath)
	}
	if (OVue & mode) == OVue {
		beeLogger.Log.Info("Creating router files...")
		writeVueControllerIndex(tables, paths.VuePath, pkgPath)
	}

	if len(notirceMsgArr) > 0 {
		beeLogger.Log.Warnf("add to file this route \n %s\n", strings.Join(notirceMsgArr, "\n"))
	}
}

// writeModelFiles generates model files
func writeModelFiles(tables []*Table, mPath string) {
	w := colors.NewColorWriter(os.Stdout)

	for _, tb := range tables {
		filename := getFileName(tb.Name)
		fpath := path.Join(mPath, filename+"Model.go")
		var f *os.File
		var err error
		if utils.IsExist(fpath) {
			beeLogger.Log.Warnf("'%s' already exists. Do you want to overwrite it? [Yes|No] ", fpath)
			if utils.AskForConfirmation() {
				f, err = os.OpenFile(fpath, os.O_RDWR|os.O_TRUNC, 0666)
				if err != nil {
					beeLogger.Log.Warnf("%s", err)
					continue
				}
			} else {
				beeLogger.Log.Warnf("Skipped create file '%s'", fpath)
				continue
			}
		} else {
			f, err = os.OpenFile(fpath, os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				beeLogger.Log.Warnf("%s", err)
				continue
			}
		}
		var template string
		if tb.Pk == "" {
			template = StructModelTPL
		} else {
			template = ModelTPL
		}
		fileStr := strings.Replace(template, "{{modelStruct}}", tb.String(), 1)
		fileStr = strings.Replace(fileStr, "{{modelName}}", utils.CamelCase(tb.Name), -1)
		fileStr = strings.Replace(fileStr, "{{tableName}}", tb.Name, -1)

		// If table contains time field, import time.Time package
		timePkg := ""
		importTimePkg := ""
		if tb.ImportTimePkg {
			timePkg = "\"time\"\n"
			importTimePkg = "import \"time\"\n"
		}
		fileStr = strings.Replace(fileStr, "{{timePkg}}", timePkg, -1)
		fileStr = strings.Replace(fileStr, "{{importTimePkg}}", importTimePkg, -1)
		if _, err := f.WriteString(fileStr); err != nil {
			beeLogger.Log.Fatalf("Could not write model file to '%s': %s", fpath, err)
		}
		utils.CloseFile(f)
		fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", fpath, "\x1b[0m")
		utils.FormatSourceCode(fpath)
	}
}

// writeControllerFiles generates controller files
func writeControllerFiles(tables []*Table, cPath string, pkgPath string) {
	w := colors.NewColorWriter(os.Stdout)

	var operateListArr []string
	for _, tb := range tables {
		if tb.Pk == "" {
			continue
		}
		filename := getFileName(tb.Name)
		fpath := path.Join(cPath, filename+"Controller.go")
		var f *os.File
		var err error
		if utils.IsExist(fpath) {
			beeLogger.Log.Warnf("'%s' already exists. Do you want to overwrite it? [Yes|No] ", fpath)
			if utils.AskForConfirmation() {
				f, err = os.OpenFile(fpath, os.O_RDWR|os.O_TRUNC, 0666)
				if err != nil {
					beeLogger.Log.Warnf("%s", err)
					continue
				}
			} else {
				beeLogger.Log.Warnf("Skipped create file '%s'", fpath)
				continue
			}
		} else {
			f, err = os.OpenFile(fpath, os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				beeLogger.Log.Warnf("%s", err)
				continue
			}
		}

		var createAutoArr []string
		var updateAutoArr []string
		var pkgListArr []string
		var isUserTime bool = false
		var isUserStrconv bool = false
		for _, col := range tb.Columns {
			if col.Name == "CreatedAt" {
				if col.Type == "datetime" {
					createAutoArr = append(createAutoArr, "v.CreatedAt = time.Now()\n")
				} else if col.Type == "int" {
					createAutoArr = append(createAutoArr, "v.CreatedAt = int(time.Now().Unix())\n")
				}
				isUserTime = true
			}
			if col.Name == "CreatedBy" {
				if col.Type == "int" {
					createAutoArr = append(createAutoArr, "v.CreatedBy,_ = strconv.Atoi(c.User.GetId())\n")
					isUserStrconv = true
				} else {
					createAutoArr = append(createAutoArr, "v.CreatedBy = c.User.GetId()\n")
				}
			}
			if col.Name == "UpdatedAt" {
				if col.Type == "datetime" {
					updateAutoArr = append(updateAutoArr, "v.UpdatedAt = time.Now()\n")
					createAutoArr = append(createAutoArr, "v.UpdatedAt = time.Now()\n")
				} else if col.Type == "int" {
					updateAutoArr = append(updateAutoArr, "v.UpdatedAt = int(time.Now().Unix())\n")
					createAutoArr = append(createAutoArr, "v.UpdatedAt = int(time.Now().Unix())\n")
				}
				isUserTime = true
			}
			if col.Name == "UpdatedBy" {
				if col.Type == "int" {
					updateAutoArr = append(updateAutoArr, "v.UpdatedBy,_ = strconv.Atoi(c.User.GetId())\n")
					isUserStrconv = true
				} else {
					updateAutoArr = append(updateAutoArr, "v.UpdatedBy = c.User.GetId()\n")
				}
			}
		}

		if isUserTime {
			pkgListArr = append(pkgListArr, "\"time\"\n")
		}
		if isUserStrconv {
			pkgListArr = append(pkgListArr, "\"strconv\"\n")
		}

		createAuto := strings.Join(createAutoArr, "")
		updateAuto := strings.Join(updateAutoArr, "")
		pkgList := strings.Join(pkgListArr, "")

		fileStr := strings.Replace(CtrlTPL, "{{ctrlName}}", utils.CamelCase(tb.Name), -1)
		fileStr = strings.Replace(fileStr, "{{pkgPath}}", pkgPath, -1)
		fileStr = strings.Replace(fileStr, "{{createAuto}}", createAuto, -1)
		fileStr = strings.Replace(fileStr, "{{updateAuto}}", updateAuto, -1)
		fileStr = strings.Replace(fileStr, "{{pkg}}", pkgList, -1)

		if _, err := f.WriteString(fileStr); err != nil {
			beeLogger.Log.Fatalf("Could not write controller file to '%s': %s", fpath, err)
		}
		utils.CloseFile(f)
		fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", fpath, "\x1b[0m")
		utils.FormatSourceCode(fpath)

		fileStr = strings.Replace(operateListTPL, "{{ctrlName}}", utils.CamelCase(tb.Name), -1)
		fileStr = strings.Replace(fileStr, "{{pageUrl}}", strings2.UrlStyleString(tb.Name), -1)

		operateListArr = append(operateListArr, fileStr)

	}
	notirceMsgArr = append(notirceMsgArr, "add to operate list:\n"+strings.Join(operateListArr, ""))
}

// writeControllerFiles generates controller files
func writeFilterFiles(tables []*Table, cPath string, pkgPath string) {
	w := colors.NewColorWriter(os.Stdout)

	for _, tb := range tables {
		if tb.Pk == "" {
			continue
		}
		filename := getFileName(tb.Name)
		fpath := path.Join(cPath, filename+"Filter.go")
		var f *os.File
		var err error
		if utils.IsExist(fpath) {
			beeLogger.Log.Warnf("'%s' already exists. Do you want to overwrite it? [Yes|No] ", fpath)
			if utils.AskForConfirmation() {
				f, err = os.OpenFile(fpath, os.O_RDWR|os.O_TRUNC, 0666)
				if err != nil {
					beeLogger.Log.Warnf("%s", err)
					continue
				}
			} else {
				beeLogger.Log.Warnf("Skipped create file '%s'", fpath)
				continue
			}
		} else {
			f, err = os.OpenFile(fpath, os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				beeLogger.Log.Warnf("%s", err)
				continue
			}
		}

		var pkgListArr []string
		var validArr []string
		var inpuTfieldListArr []string
		var isUserTime bool = false
		for _, col := range tb.Columns {
			if col.Tag.Null == false && col.Tag.Column != tb.Pk {
				fileStr := strings.Replace(FilterValidRuleTPL, "{{validFunc}}", "Required", -1)
				fileStr = strings.Replace(fileStr, "{{colName}}", col.Name, -1)
				fileStr = strings.Replace(fileStr, "{{colColumn}}", col.Tag.Column, -1)
				fileStr = strings.Replace(fileStr, "{{msg}}", col.Tag.Column+" is required", -1)
				validArr = append(validArr, fileStr)
			}

			if tb.Pk != col.Tag.Column && col.Name != "CreatedAt" && col.Name != "CreatedBy" && col.Name != "UpdatedAt" && col.Name != "UpdatedBy" {

				structfield := fmt.Sprintf("%s %s %s", col.Name, col.Type, col.Tag.String())
				inpuTfieldListArr = append(inpuTfieldListArr, structfield)
			}
		}

		if isUserTime {
			pkgListArr = append(pkgListArr, "\"time\"\n")
		}

		pkgList := strings.Join(pkgListArr, "")
		validStr := strings.Join(validArr, "")
		inpuTfieldList := strings.Join(inpuTfieldListArr, "\n	")

		fileStr := strings.Replace(FilterTPL, "{{modelName}}", utils.CamelCase(tb.Name), -1)
		fileStr = strings.Replace(fileStr, "{{pkgPath}}", pkgPath, -1)
		fileStr = strings.Replace(fileStr, "{{pkg}}", pkgList, -1)
		fileStr = strings.Replace(fileStr, "{{ValidRuleList}}", validStr, -1)
		fileStr = strings.Replace(fileStr, "{{inpuTfieldList}}", inpuTfieldList, -1)

		if _, err := f.WriteString(fileStr); err != nil {
			beeLogger.Log.Fatalf("Could not write filter file to '%s': %s", fpath, err)
		}
		utils.CloseFile(f)
		fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", fpath, "\x1b[0m")
		utils.FormatSourceCode(fpath)
	}
}

// writeRouterFile generates router file
func writeRouterFile(tables []*Table, rPath string, pkgPath string) {
	w := colors.NewColorWriter(os.Stdout)

	var nameSpaces []string
	for _, tb := range tables {
		if tb.Pk == "" {
			continue
		}
		// Add namespaces
		nameSpace := strings.Replace(NamespaceTPL, "{{nameSpace}}", strings2.UrlStyleString(tb.Name), -1)
		nameSpace = strings.Replace(nameSpace, "{{ctrlName}}", utils.CamelCase(tb.Name), -1)
		nameSpaces = append(nameSpaces, nameSpace)
	}
	// Add export controller
	fpath := filepath.Join(rPath, "router.go")
	routerStr := strings.Replace(RouterTPL, "{{nameSpaces}}", strings.Join(nameSpaces, ""), 1)
	routerStr = strings.Replace(routerStr, "{{pkgPath}}", pkgPath, 1)
	var f *os.File
	var err error
	if utils.IsExist(fpath) {
		//beeLogger.Log.Warnf("'%s' already exists. Do you want to overwrite it? [Yes|No] ", fpath)
		//if utils.AskForConfirmation() {
		//	f, err = os.OpenFile(fpath, os.O_RDWR|os.O_TRUNC, 0666)
		//	if err != nil {
		//		beeLogger.Log.Warnf("%s", err)
		//		return
		//	}
		//} else {
		beeLogger.Log.Warnf("Skipped create file '%s'", fpath)
		notirceMsgArr = append(notirceMsgArr, "add to routers/router.go \n"+strings.Join(nameSpaces, ""))
		return
		//}
	} else {
		f, err = os.OpenFile(fpath, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			beeLogger.Log.Warnf("%s", err)
			return
		}
	}
	if _, err := f.WriteString(routerStr); err != nil {
		beeLogger.Log.Fatalf("Could not write router file to '%s': %s", fpath, err)
	}
	utils.CloseFile(f)
	fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", fpath, "\x1b[0m")
	utils.FormatSourceCode(fpath)
}

// writeControllerFiles generates controller files
func writeVueControllerIndex(tables []*Table, cPath string, pkgPath string) {
	w := colors.NewColorWriter(os.Stdout)

	var vueRuleArr []string
	var vueMenuArr []string

	for _, tb := range tables {
		if tb.Pk == "" {
			continue
		}
		vueComponentPath := strings2.LowerCamelCase(tb.Name)
		pageUrl := strings2.UrlStyleString(tb.Name)

		cBase := cPath + string(os.PathSeparator) + vueComponentPath
		os.Mkdir(cBase, 0777)

		//列表
		fpathIndex := path.Join(cBase, "Index.vue")
		var f *os.File
		var err error
		if utils.IsExist(fpathIndex) {
			beeLogger.Log.Warnf("'%s' already exists. Do you want to overwrite it? [Yes|No] ", fpathIndex)
			if utils.AskForConfirmation() {
				f, err = os.OpenFile(fpathIndex, os.O_RDWR|os.O_TRUNC, 0666)
				if err != nil {
					beeLogger.Log.Warnf("%s", err)
					continue
				}
			} else {
				beeLogger.Log.Warnf("Skipped create file '%s'", fpathIndex)
				continue
			}
		} else {
			f, err = os.OpenFile(fpathIndex, os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				beeLogger.Log.Warnf("%s", err)
				continue
			}
		}

		var listColumnsArr []string
		var listColumnShowArr []string
		var selectOptionsArr []string
		var createFromFieldArr []string
		var customFieldCreateArr []string
		var customFieldEditArr []string
		var customRulesCreateArr []string
		var customRulesEditArr []string
		var editfromFieldArr []string
		var editSubmitItemsArr []string
		var createSubmitDataFixArr []string
		var indexModifyRowsArr []string

		var index int32 = 0
		for _, col := range tb.Columns {
			fieldComment := col.Tag.Comment
			if fieldComment == "" {
				fieldComment = col.Name
			}
			// Add index page list column
			tlpstr := strings.Replace(VueIndexListColumnTPL, "{{fieldName}}", col.Tag.Column, -1)
			tlpstr = strings.Replace(tlpstr, "{{fieldComment}}", fieldComment, -1)
			if index > 6 {
				tlpstr = strings.Replace(tlpstr, "show: true", "show: false", -1)
			} else {
				listColumnShowArr = append(listColumnShowArr, tlpstr)
			}
			listColumnsArr = append(listColumnsArr, tlpstr)

			// Add index page select column
			tlpstr = strings.Replace(VueIndexSelectOptionTPL, "{{fieldName}}", col.Tag.Column, -1)
			tlpstr = strings.Replace(tlpstr, "{{fieldComment}}", fieldComment, -1)
			selectOptionsArr = append(selectOptionsArr, tlpstr)

			// Add index page list column
			if tb.Pk != col.Tag.Column && col.Name != "CreatedAt" && col.Name != "CreatedBy" && col.Name != "UpdatedAt" && col.Name != "UpdatedBy" {
				tlpstr = strings.Replace(VueCreateFieldComponentTPL, "{{fieldName}}", col.Tag.Column, -1)
				tlpstr = strings.Replace(tlpstr, "{{fieldComment}}", fieldComment, -1)
				createFromFieldArr = append(createFromFieldArr, tlpstr)
			}

			// add index row modify
			if (col.Name == "CreatedAt" || col.Name == "UpdatedAt") && col.Type == "int" {
				indexModifyRowsArr = append(indexModifyRowsArr, "\n                            list[i]."+col.Tag.Column+" = this.format(list[i]."+col.Tag.Column+");")
			}

			// Add index page customFieldEdit
			tlpstrField := strings.Replace(VueCreateCustomFormComponentTPL, "{{fieldName}}", col.Tag.Column, -1)
			tlpstrField = strings.Replace(tlpstrField, "{{fieldDefault}}", col.Tag.Default, -1)
			customFieldEditArr = append(customFieldEditArr, tlpstrField)

			if tb.Pk != col.Tag.Column {

				// Add index page customRules
				tlpstr = strings.Replace(VueCreateCustomRulesComponentTPL, "{{fieldName}}", col.Tag.Column, -1)
				tlpstr = strings.Replace(tlpstr, "{{fieldComment}}", fieldComment, -1)
				if col.Tag.Null != true {
					tlpstr = strings.Replace(tlpstr, "{{required}}", "true", -1)
				} else {
					tlpstr = strings.Replace(tlpstr, "{{required}}", "false", -1)
				}
				if col.Tag.Size != "" {
					tlpstr = strings.Replace(tlpstr, "{{length}}", col.Tag.Size, -1)
				} else {
					tlpstr = strings.Replace(tlpstr, "length: {{length}},", "", -1)
				}

				jsValidatorType := ""
				if col.Type == "int" {
					jsValidatorType = "number"
				} else if col.Type == "int8" {
					jsValidatorType = "integer"
				}
				tlpstr = strings.Replace(tlpstr, "{{type}}", jsValidatorType, -1)

				customRulesEditArr = append(customRulesEditArr, tlpstr)

				if col.Name != "CreatedAt" && col.Name != "CreatedBy" && col.Name != "UpdatedAt" && col.Name != "UpdatedBy" {
					// Add index page customFieldCreate
					customFieldCreateArr = append(customFieldCreateArr, tlpstrField)

					// Add index page customRulesCreate
					customRulesCreateArr = append(customRulesCreateArr, tlpstr)
				}
			}

			// Add index page list column
			tlpstr = strings.Replace(vueEditComponentFromItemTPL, "{{fieldName}}", col.Tag.Column, -1)
			tlpstr = strings.Replace(tlpstr, "{{fieldComment}}", fieldComment, -1)
			if tb.Pk == col.Tag.Column || col.Name == "CreatedAt" || col.Name == "CreatedBy" || col.Name == "UpdatedAt" || col.Name == "UpdatedBy" {
				tlpstr = strings.Replace(tlpstr, "{{disabled}}", "disabled", -1)
			} else {
				tlpstr = strings.Replace(tlpstr, "{{disabled}}", "", -1)
			}
			editfromFieldArr = append(editfromFieldArr, tlpstr)

			// Add eidt page submit column
			if tb.Pk != col.Tag.Column && col.Name != "CreatedAt" && col.Name != "CreatedBy" && col.Name != "UpdatedAt" && col.Name != "UpdatedBy" {
				tlpstr = strings.Replace(vueEditComponentSubmitItemTPL, "{{fieldName}}", col.Tag.Column, -1)
				if col.Type == "int" || col.Type == "int8" {
					tlpstr = strings.Replace(tlpstr, "this.customForm."+col.Tag.Column, "parseInt(this.customForm."+col.Tag.Column+")", -1)
					createSubmitDataFixArr = append(createSubmitDataFixArr, "\n			      	  params['"+col.Tag.Column+"'] = parseInt(params['"+col.Tag.Column+"']);")
				} else if col.Type == "float" {
					tlpstr = strings.Replace(tlpstr, "this.customForm."+col.Tag.Column, "parseFloat(this.customForm."+col.Tag.Column+")", -1)
					createSubmitDataFixArr = append(createSubmitDataFixArr, "	\n				  	  params['"+col.Tag.Column+"'] = parseFloat(params['"+col.Tag.Column+"']);")
				}
				editSubmitItemsArr = append(editSubmitItemsArr, tlpstr)
			}

			index++
		}

		listColumns := strings.Join(listColumnsArr, "")
		listColumnShow := strings.Join(listColumnShowArr, "")
		selectOptions := strings.Join(selectOptionsArr, "")
		createFromField := strings.Join(createFromFieldArr, "")
		customFieldCreate := strings.Join(customFieldCreateArr, "")
		customFieldEdit := strings.Join(customFieldEditArr, "")
		customRulesCreate := strings.Join(customRulesCreateArr, "")
		customRulesEdit := strings.Join(customRulesEditArr, "")
		editfromField := strings.Join(editfromFieldArr, "")
		editSubmitItems := strings.Join(editSubmitItemsArr, "")
		createSubmitDataFix := strings.Join(createSubmitDataFixArr, "")
		indexModifyRows := strings.Join(indexModifyRowsArr, "")

		fileStr := strings.Replace(VueIndexTPL, "{{ctrlName}}", utils.CamelCase(tb.Name), -1)
		fileStr = strings.Replace(fileStr, "{{tbName}}", tb.Name, -1)
		fileStr = strings.Replace(fileStr, "{{tbPk}}", tb.Pk, -1)
		fileStr = strings.Replace(fileStr, "{{pkgPath}}", pkgPath, -1)
		fileStr = strings.Replace(fileStr, "{{listColumn}}", listColumns, -1)
		fileStr = strings.Replace(fileStr, "{{listColumnShow}}", listColumnShow, -1)
		fileStr = strings.Replace(fileStr, "{{selectOptions}}", selectOptions, -1)
		fileStr = strings.Replace(fileStr, "{{pageUrl}}", pageUrl, -1)
		if len(indexModifyRowsArr) > 0 {
			colModifyStr := strings.Replace(VueIndexColModifyTPL, "{{indexModifyRows}}", indexModifyRows, -1)
			fileStr = strings.Replace(fileStr, "{{colModifyStr}}", colModifyStr, -1)
		} else {
			fileStr = strings.Replace(fileStr, "{{colModifyStr}}", "", -1)
		}

		if _, err := f.WriteString(fileStr); err != nil {
			beeLogger.Log.Fatalf("Could not write controller file to '%s': %s", fpathIndex, err)
		}
		utils.CloseFile(f)
		fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", fpathIndex, "\x1b[0m")
		utils.FormatSourceCode(fpathIndex)

		//添加组件
		fpathIndex = path.Join(cBase, "CreateComponent.vue")
		if utils.IsExist(fpathIndex) {
			beeLogger.Log.Warnf("'%s' already exists. Do you want to overwrite it? [Yes|No] ", fpathIndex)
			if utils.AskForConfirmation() {
				f, err = os.OpenFile(fpathIndex, os.O_RDWR|os.O_TRUNC, 0666)
				if err != nil {
					beeLogger.Log.Warnf("%s", err)
					continue
				}
			} else {
				beeLogger.Log.Warnf("Skipped create file '%s'", fpathIndex)
				continue
			}
		} else {
			f, err = os.OpenFile(fpathIndex, os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				beeLogger.Log.Warnf("%s", err)
				continue
			}
		}
		fileStr = strings.Replace(VueCreateComponentTPL, "{{ctrlName}}", utils.CamelCase(tb.Name), -1)
		fileStr = strings.Replace(fileStr, "{{tbName}}", tb.Name, -1)
		fileStr = strings.Replace(fileStr, "{{tbPk}}", tb.Pk, -1)
		fileStr = strings.Replace(fileStr, "{{pkgPath}}", pkgPath, -1)
		fileStr = strings.Replace(fileStr, "{{fromField}}", createFromField, -1)
		fileStr = strings.Replace(fileStr, "{{customField}}", customFieldCreate, -1)
		fileStr = strings.Replace(fileStr, "{{customRules}}", customRulesCreate, -1)
		fileStr = strings.Replace(fileStr, "{{pageUrl}}", pageUrl, -1)
		fileStr = strings.Replace(fileStr, "{{createSubmitDataFix}}", createSubmitDataFix, -1)
		if _, err := f.WriteString(fileStr); err != nil {
			beeLogger.Log.Fatalf("Could not write controller file to '%s': %s", fpathIndex, err)
		}
		utils.CloseFile(f)
		fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", fpathIndex, "\x1b[0m")
		utils.FormatSourceCode(fpathIndex)

		//编辑组件
		fpathIndex = path.Join(cBase, "EditComponent.vue")
		if utils.IsExist(fpathIndex) {
			beeLogger.Log.Warnf("'%s' already exists. Do you want to overwrite it? [Yes|No] ", fpathIndex)
			if utils.AskForConfirmation() {
				f, err = os.OpenFile(fpathIndex, os.O_RDWR|os.O_TRUNC, 0666)
				if err != nil {
					beeLogger.Log.Warnf("%s", err)
					continue
				}
			} else {
				beeLogger.Log.Warnf("Skipped create file '%s'", fpathIndex)
				continue
			}
		} else {
			f, err = os.OpenFile(fpathIndex, os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				beeLogger.Log.Warnf("%s", err)
				continue
			}
		}
		fileStr = strings.Replace(vueEditComponentTPL, "{{ctrlName}}", utils.CamelCase(tb.Name), -1)
		fileStr = strings.Replace(fileStr, "{{fromField}}", editfromField, -1)
		fileStr = strings.Replace(fileStr, "{{tbName}}", tb.Name, -1)
		fileStr = strings.Replace(fileStr, "{{tbPk}}", tb.Pk, -1)
		fileStr = strings.Replace(fileStr, "{{pkgPath}}", pkgPath, -1)
		fileStr = strings.Replace(fileStr, "{{customField}}", customFieldEdit, -1)
		fileStr = strings.Replace(fileStr, "{{customRules}}", customRulesEdit, -1)
		fileStr = strings.Replace(fileStr, "{{editSubmitItems}}", editSubmitItems, -1)
		fileStr = strings.Replace(fileStr, "{{pageUrl}}", pageUrl, -1)

		if _, err := f.WriteString(fileStr); err != nil {
			beeLogger.Log.Fatalf("Could not write controller file to '%s': %s", fpathIndex, err)
		}
		utils.CloseFile(f)
		fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", fpathIndex, "\x1b[0m")
		utils.FormatSourceCode(fpathIndex)

		//列显示设置组件
		fpathIndex = path.Join(cBase, "ColSettingComponent.vue")
		if utils.IsExist(fpathIndex) {
			beeLogger.Log.Warnf("'%s' already exists. Do you want to overwrite it? [Yes|No] ", fpathIndex)
			if utils.AskForConfirmation() {
				f, err = os.OpenFile(fpathIndex, os.O_RDWR|os.O_TRUNC, 0666)
				if err != nil {
					beeLogger.Log.Warnf("%s", err)
					continue
				}
			} else {
				beeLogger.Log.Warnf("Skipped create file '%s'", fpathIndex)
				continue
			}
		} else {
			f, err = os.OpenFile(fpathIndex, os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				beeLogger.Log.Warnf("%s", err)
				continue
			}
		}
		fileStr = strings.Replace(vueColSettingComponentTPL, "{{ctrlName}}", utils.CamelCase(tb.Name), -1)
		if _, err := f.WriteString(fileStr); err != nil {
			beeLogger.Log.Fatalf("Could not write controller file to '%s': %s", fpathIndex, err)
		}
		utils.CloseFile(f)
		fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", fpathIndex, "\x1b[0m")
		utils.FormatSourceCode(fpathIndex)

		//vue 路由规则
		fileStr = strings.Replace(vueRuleTPL, "{{pageUrl}}", pageUrl, -1)
		fileStr = strings.Replace(fileStr, "{{filPath}}", vueComponentPath, -1)
		vueRuleArr = append(vueRuleArr, fileStr)

		//vue 菜单
		fileStr = strings.Replace(menuListTPL, "{{pageUrl}}", pageUrl, -1)
		fileStr = strings.Replace(fileStr, "{{ctrlName}}", utils.CamelCase(tb.Name), -1)
		vueMenuArr = append(vueMenuArr, fileStr)
	}
	notirceMsgArr = append(notirceMsgArr, "add to vue vue/src/router/index.js \n"+strings.Join(vueRuleArr, ""))
	notirceMsgArr = append(notirceMsgArr, "add to vue menu \n"+strings.Join(vueMenuArr, ""))

}

func isSQLTemporalType(t string) bool {
	return t == "date" || t == "datetime" || t == "timestamp" || t == "time"
}

func isSQLStringType(t string) bool {
	return t == "char" || t == "varchar"
}

func isSQLSignedIntType(t string) bool {
	return t == "int" || t == "tinyint" || t == "smallint" || t == "mediumint" || t == "bigint"
}

func isSQLDecimal(t string) bool {
	return t == "decimal"
}

func isSQLBinaryType(t string) bool {
	return t == "binary" || t == "varbinary"
}

func isSQLBitType(t string) bool {
	return t == "bit"
}
func isSQLStrangeType(t string) bool {
	return t == "interval" || t == "uuid" || t == "json"
}

// extractColSize extracts field size: e.g. varchar(255) => 255
func extractColSize(colType string) string {
	regex := regexp.MustCompile(`^[a-z]+\(([0-9]+)\)$`)
	size := regex.FindStringSubmatch(colType)
	return size[1]
}

func extractIntSignness(colType string) string {
	regex := regexp.MustCompile(`(int|smallint|mediumint|bigint)\([0-9]+\)(.*)`)
	signRegex := regex.FindStringSubmatch(colType)
	return strings.Trim(signRegex[2], " ")
}

func extractDecimal(colType string) (digits string, decimals string) {
	decimalRegex := regexp.MustCompile(`decimal\(([0-9]+),([0-9]+)\)`)
	decimal := decimalRegex.FindStringSubmatch(colType)
	digits, decimals = decimal[1], decimal[2]
	return
}

func getFileName(tbName string) (filename string) {
	// avoid test file
	filename = utils.CamelCase(tbName)
	for strings.HasSuffix(filename, "_test") {
		pos := strings.LastIndex(filename, "_")
		filename = filename[:pos] + filename[pos+1:]
	}
	return
}

func getPackagePath(curpath string) (packpath string) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		beeLogger.Log.Fatal("GOPATH environment variable is not set or empty")
	}

	beeLogger.Log.Debugf("GOPATH: %s", utils.FILE(), utils.LINE(), gopath)

	appsrcpath := ""
	haspath := false
	wgopath := filepath.SplitList(gopath)

	for _, wg := range wgopath {
		wg, _ = filepath.EvalSymlinks(filepath.Join(wg, "src"))
		if strings.HasPrefix(strings.ToLower(curpath), strings.ToLower(wg)) {
			haspath = true
			appsrcpath = wg
			break
		}
	}

	if !haspath {
		beeLogger.Log.Fatalf("Cannot generate application code outside of GOPATH '%s' compare with CWD '%s'", gopath, curpath)
	}

	if curpath == appsrcpath {
		beeLogger.Log.Fatal("Cannot generate application code outside of application path")
	}

	packpath = strings.Join(strings.Split(curpath[len(appsrcpath)+1:], string(filepath.Separator)), "/")
	return
}

const (
	StructModelTPL = `package models
{{importTimePkg}}
{{modelStruct}}
`

	ModelTPL string = `package models

import (
	"strings"
	{{timePkg}}
    "github.com/yimishiji/bee/pkg/db"
)

{{modelStruct}}

func (t *{{modelName}}) TableName() string {
	return "{{tableName}}"
}


// Add{{modelName}} insert a new {{modelName}} into database and returns
// last inserted Id on success.
func Add{{modelName}}(m *{{modelName}}) (err error) {
    res := db.Conn.Create(m)
    return res.Error
}

// Get{{modelName}}ById retrieves {{modelName}} by Id. Returns error if
// Id doesn't exist
func Get{{modelName}}ById(id int) (v {{modelName}}, err error) {
    res := db.Conn.Where(id).First(&v)
    return v, res.Error
}

// GetAll{{modelName}} retrieves all {{modelName}} matches certain condition. Returns empty list if
// no records exist
func GetAll{{modelName}}(query map[string]string, fields []string, sortFields []string, offset int64, limit int64) (ml []{{modelName}}, total int64, err error) {
    //过虑条件
    gormQuery := db.NewGormQuery(query)

    //排序
    for _, v := range sortFields {
        gormQuery = gormQuery.Order(v)
    }

    //获取总页数
    var itemCount int64
    gormQuery.Model({{modelName}}{}).Count(&itemCount)

    //select
    if len(fields) > 0 {
        gormQuery = gormQuery.Select(strings.Join(fields, ","))
    }

    //查询
	var l []{{modelName}}
    err = gormQuery.Limit(limit).Offset(offset).Find(&l).Error
    if err != nil {
        return nil, itemCount, err
    }

    // 如果需要精简返回值，将返回列表类型设为 []interface{}，再调用以下函数
    //ml = db.SelectField(l, fields)

	return l, itemCount, err
}

// Update{{modelName}} updates {{modelName}} by Id and returns error if
// the record to be updated doesn't exist
func Update{{modelName}}(m *{{modelName}}) (err error) {
	return db.Conn.Save(m).Error
}

// Delete{{modelName}} deletes {{modelName}} by Id and returns error if
// the record to be deleted doesn't exist
func Delete{{modelName}}(id int) (err error) {
	v := new({{modelName}})

	// ascertain id exists in the database
	err = db.Conn.Where(id).First(&v).Error
	if err != nil {
		return err
	}

	return  db.Conn.Delete(&v).Error
}

// BeforeCreate hook
//func (user *SdbB2cCustomerLog) BeforeCreate(scope *gorm.Scope) error {
//    //scope.SetColumn("ID", uuid.New())
//    return nil
//}
`
	CtrlTPL = `package controllers

import (
	"{{pkgPath}}/models"
    "{{pkgPath}}/filters"
	{{pkg}}	

    "github.com/yimishiji/bee/pkg/base"
    "github.com/yimishiji/bee/pkg/structs"
)

// {{ctrlName}}Controller operations for {{ctrlName}}
type {{ctrlName}}Controller struct {
	base.Controller
    filter *filters.{{ctrlName}}Filter
}

// URLMapping ...
func (c *{{ctrlName}}Controller) URLMapping() {
	c.Mapping("Post", c.Post)
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("Put", c.Put)
	c.Mapping("Delete", c.Delete)
}

// init inputFilter
func (c *{{ctrlName}}Controller) Prepare() {
    c.filter = filters.New{{ctrlName}}Filter(c.Ctx.Input)
}

// Post ...
// @Title Post
// @Description create {{ctrlName}}
// @Param	body		body 	models.{{ctrlName}}	true		"body for {{ctrlName}} content"
// @Success 201 {int} models.{{ctrlName}}
// @Failure 403 body is empty
// @router / [post]
func (c *{{ctrlName}}Controller) Post() {
    var v models.{{ctrlName}}
    if f, err := c.filter.Get{{ctrlName}}Post(); err == nil {
        structs.StructMerge(&v, f)
		{{createAuto}}
		if err := models.Add{{ctrlName}}(&v); err == nil {
			c.Ctx.Output.SetStatus(201)
			c.Data["json"] = c.Resp(base.ApiCode_SUCC, "ok", v)
		} else {
			c.Data["json"] = c.Resp(base.ApiCode_SYS_ERROR, "system error", err.Error())
		}
	} else {
		c.Data["json"] = c.Resp(base.ApiCode_VALIDATE_ERROR, "invalid:"+err.Error(), err.Error())
	}
	c.ServeJSON()
}

// GetOne ...
// @Title Get One
// @Description get {{ctrlName}} by id
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 :id is empty
// @router /:id [get]
func (c *{{ctrlName}}Controller) GetOne() {
    id := c.filter.GetId(":id")
	v, err := models.Get{{ctrlName}}ById(id)
	if err != nil {
		c.Data["json"] = c.Resp(base.ApiCode_VALIDATE_ERROR, "not find", err.Error())
	} else {
		c.Data["json"] = c.Resp(base.ApiCode_SUCC, "ok", v)
	}
	c.ServeJSON()
}

// GetAll ...
// @Title Get All
// @Description get {{ctrlName}}
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2,col-isnull:,col:>50,col:like-adc,col:between-10-20 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403
// @router / [get]
func (c *{{ctrlName}}Controller) GetAll() {
    pageParams, err := c.filter.GetListPrams()
	if err != nil {
		c.Data["json"] = c.Resp(base.ApiCode_ILLEGAL_ERROR, "illegal operation", err)
		c.ServeJSON()
	}

    l, itemCount, err := models.GetAll{{ctrlName}}(pageParams.Querys, pageParams.Field, pageParams.SortFields, pageParams.Offsets, pageParams.Limits)
	if err != nil {
		c.Data["json"] = c.Resp(base.ApiCode_ILLEGAL_ERROR, "not find", err.Error())
	} else {
        list := base.NewListPageData(pageParams.Limits, pageParams.Offsets, itemCount, l)
        c.Data["json"] = c.Resp(base.ApiCode_SUCC, "ok", list)
	}
	c.ServeJSON()
}

// Put ...
// @Title Put
// @Description update the {{ctrlName}}
// @Param	id		path 	string	true		"The id you want to update"
// @Param	body		body 	models.{{ctrlName}}	true		"body for {{ctrlName}} content"
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 :id is not int
// @router /:id [put]
func (c *{{ctrlName}}Controller) Put() {
    id := c.filter.GetId(":id")
    v, err := models.Get{{ctrlName}}ById(id)
    if err != nil {
        c.Data["json"] = c.Resp(base.ApiCode_VALIDATE_ERROR, "invalid:"+err.Error(), err.Error())
    }

    if f, err := c.filter.Get{{ctrlName}}Put(); err == nil {
        structs.StructMerge(&v, f)
		{{updateAuto}}
		if err := models.Update{{ctrlName}}(&v); err == nil {
			c.Data["json"] = c.Resp(base.ApiCode_SUCC, "ok")
		} else {
			c.Data["json"] = c.Resp(base.ApiCode_SYS_ERROR, "system error", err.Error())
		}
	} else {
		c.Data["json"] = c.Resp(base.ApiCode_VALIDATE_ERROR, "invalid:"+err.Error(), err.Error())
	}
	c.ServeJSON()
}

// Delete ...
// @Title Delete
// @Description delete the {{ctrlName}}
// @Param	id		path 	string	true		"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *{{ctrlName}}Controller) Delete() {
    id := c.filter.GetId(":id")
	if err := models.Delete{{ctrlName}}(id); err == nil {
		c.Data["json"] = c.Resp(base.ApiCode_SUCC, "ok")
	} else {
		c.Data["json"] = c.Resp(base.ApiCode_ILLEGAL_ERROR, "illegal operation", err.Error())
	}
	c.ServeJSON()
}
`
	FilterTPL = `
package filters

import (
	"encoding/json"
	"errors"

	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/validation"
	"github.com/yimishiji/bee/pkg/filters"
)

type {{modelName}}Filter struct {
	filters.InputFilter
}

func New{{modelName}}Filter(r *context.BeegoInput) *{{modelName}}Filter {
	return &{{modelName}}Filter{
		filters.InputFilter{
			Input: r,
		},
	}
}

//post提交 数据格式
type {{modelName}}Post struct {
    {{inpuTfieldList}}
}

//获取Post接交数据
func (this *{{modelName}}Filter) Get{{modelName}}Post() (v {{modelName}}Post, err error) {
	if err := json.Unmarshal(this.Input.RequestBody, &v); err == nil {
		//验证器
		valid := validation.Validation{}{{ValidRuleList}}
		if valid.HasErrors() {
			err = errors.New(valid.Errors[0].String())
			return v, err
		}
		//自定义验证方法
		//if filters.InStingArr(v.Type, []string{"orders", "goods", "users"}) == false {
		//	return v, errors.New("type is not enable")
		//}
		return v, nil
	} else {
		return v, err
	}
}

//Put提交 数据格式, 每个表单提交需针对性定义一份结构体,
type {{modelName}}Put struct {
     {{inpuTfieldList}}
}

//获取put提交数据
func (this *{{modelName}}Filter) Get{{modelName}}Put() (v {{modelName}}Put, err error) {
	if err := json.Unmarshal(this.Input.RequestBody, &v); err == nil {
		//验证器
		valid := validation.Validation{}{{ValidRuleList}}
		if valid.HasErrors() {
			err = errors.New(valid.Errors[0].String())
			return v, err
		}
		return v, nil
	} else {
		return v, err
	}
}

//分页参数
func (this *{{modelName}}Filter) GetListPrams() (params *filters.PageCommonParams, err error) {
	if params, err := this.GetPagePublicParams(); err == nil {

		//验证筛选的条件合法性
		//if t, ok := params.Querys["type"]; ok {
		//	if filters.InStingArr(t, []string{"orders", "goods", "users"}) == false {
		//		return params, errors.New("type is not enable")
		//	}
		//}

		return params, nil
	} else {
		return params, err
	}
}
`
	FilterValidRuleTPL = `
		valid.{{validFunc}}(v.{{colName}}, "{{colColumn}}").Message("{{msg}}")`
	RouterTPL = `// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"{{pkgPath}}/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/v1",
		{{nameSpaces}}
	)
	beego.AddNamespace(ns)
}
`
	NamespaceTPL = `
		beego.NSNamespace("/{{nameSpace}}",
			beego.NSInclude(
				&controllers.{{ctrlName}}Controller{},
			),
		),`
	VueIndexTPL = `
<template>
    <div class="main">
        <div class="top">
            <v-select placeholder="请选择类型" size="lg" class="change" v-model="searchType" :data="options"></v-select>
            <div class="search">
                <v-input placeholder="请输入" size="large" v-model="searchText" class="distance-up10" @keyup.enter.native="refreshTable()"></v-input>
                <v-button type="primary" class="search-button" @click="refreshTable()" size="large">搜&nbsp;索</v-button>
            </div>
            <v-button type="primary" class="add-button" size="large" @click="create" >
                <v-icon type="plus"></v-icon>
                &nbsp;&nbsp;添加
            </v-button>
            <span  @click="settingCol()" class="pull-right"><v-icon type="setting" class="colsetting"></v-icon></span>
        </div>
        <div class="table goods">
            <v-data-table :data='loadData' :columns='columns' ref="xtable" stripe bordered >
                <template slot="td" slot-scope="props" attrs='width="502px"'>

                    <div v-if="props.column.field=='action'" class="operate">
                        <span @click="view(props.item)">
                            <v-tooltip content="查看" :placement="props.index==0 ? 'bottom' : 'top'" ><v-icon type="eye-o"></v-icon></v-tooltip>
                        </span>
                        <span @click="edit(props.item)">
                            <v-tooltip content="编辑" :placement="props.index==0 ? 'bottom' : 'top'" ><v-icon type="edit"></v-icon></v-tooltip>
                        </span>
                        <v-popconfirm :placement="props.index==0 ? 'bottom' : 'top'" title=" 确定删除吗?" @confirm="del(props.item)">
                            <v-tooltip content="删除" :placement="props.index==0 ? 'bottom' : 'top'" ><i class="fa fa-trash-o "></i></v-tooltip>
                        </v-popconfirm>
                    </div>

                    <span v-else v-html="props.content"></span>
                </template>
            </v-data-table>
        </div>
        <create-item @refreshList="refreshTable" ref="createRef"></create-item>
        <edit-item @refreshList="refreshTable" ref="editRef"></edit-item>
        <col-setting :columnsSetting="columnsSetting" @setCols="setColomns" ref="colSettingRef"></col-setting>
    </div>
</template>

<script>
    import {domainHost,  hostName} from '../../config/api'
    import { formatDate } from '../../common/date'
    import createVue from './CreateComponent.vue'
    import editVue from './EditComponent.vue'
    import colSetting from './ColSettingComponent.vue'

    const IndexApi = hostName+"v1/{{pageUrl}}";
    const DeleteAPI = hostName+"v1/{{pageUrl}}";

    export default {
        data () {
            return {
                //下拉搜索选择
                options    : [ {{selectOptions}}
                ],
                //下拉选中
                searchType : '',
                //搜索框
                searchText : '',
                //商品列表---状态
                checkYes   : 1,
                checkNo    : 1,
                checkStatus: '',
                //商品列表头部
                columns    : [{{listColumnShow}}
                    {title: "operate", field: 'action', show: true},
                ],
                columnsSetting : [{{listColumn}}
                    {title: "operate", field: 'action', show: true},
                ],
            }
        },
        created: function () {
            this.$store.state.BreadShow = true;
            this.$store.state.oneValue = {value: "Index", url: ""};
            this.$store.state.twoValue = {value: '{{tbName}}', url: ""};
            this.$store.state.threeValue = {value: "List", url: ""};
        },
        components:{
            'create-item': createVue,
            'edit-item': editVue,
            'col-setting': colSetting,
        },
        methods: {
            refreshTable:function () {
              this.$refs.xtable.refresh();
            },
            loadData(pramas) {
                let params = {
                    'limit': pramas.pageSize,
                    'page': pramas.pageNo,
                    'order':'desc',
                    'sortby':"{{tbPk}}",
                    //'query':{}
                };

                if(this.searchType && this.searchText.trim()){
                    params[ "query"] = this.searchType+":"+this.searchText.trim();
                }

                return this.$http.get(IndexApi,{ params }).then(resp =>{
                    if (resp.data.status == 1){
                        var result = resp.data.results[0];
                        let list = result.list? result.list : [];
                        {{colModifyStr}}
                        let listdata = {};
                        listdata['result'] = list;
                        listdata['totalCount'] = Number(result.totalCount);
                        return listdata;
                    }
                });
            },
            create: function () {
                this.$refs.createRef.show = true;
            },
            view: function (item) {
                this.$refs.editRef.{{tbPk}} = item.{{tbPk}};
                for (let key in item) {
                    this.$refs.editRef.customForm[key] = item[key];
                }
                this.$refs.editRef.updateMode = false
                this.$refs.editRef.show = true;
            },
            edit: function (item) {
                this.$refs.editRef.{{tbPk}} = item.{{tbPk}};
                for (let key in item) {
                    this.$refs.editRef.customForm[key] = item[key];
                }
                this.$refs.editRef.updateMode = true;
                this.$refs.editRef.show = true;
            },
            del:function (item) {
                this.$store.state.loading     = true;
                this.$http.delete(DeleteAPI+"/"+item.{{tbPk}}).then(resp=> {
                    this.$store.state.loading = false;
                    if (resp.data.status == 1) {
                        this.$notification.success({
                            message    : '提示',
                            duration   : 2,
                            description: "删除成功"
                        });
                        this.refreshTable();
                    }
                });
            },
            format:function(time){
                let date = new Date(parseInt(time));
                return formatDate(date,'yyyy-MM-dd hh:mm:ss');
            },
            settingCol:function(){
                this.$refs.colSettingRef.show = true;
            },
            setColomns:function(cols){
                this.columns = cols;
            },
        }
    }
</script>

<style scoped>
   .main{
       padding: 24px;
       width: 100%;
       min-height: calc(100% - 52px);
       overflow: auto;
       background: #fff;
   }
   .top {
       min-width: 540px;
       height: 38px;
       text-align: left;
   }
   /* 下拉 */
   .top .change {
       margin-right: 10px;
       width: 100px;
       float: left;
   }
   /* 搜索 */
   .top .add-button {
       margin-left: 10px;
       float: left;
   }
   .top .search {
       width: 330px;
       height: 32px;
       float: left;
       position: relative;
   }
   .top .search input {
       padding: 6px 8px 6px 6px;
       width: 260px;
       float: left;
       box-shadow:none !important;
   }
   .top .search .search-button {
       width: 70px;
       height: 32px;
       position: absolute;
       top: 0px;
       right: 4px;
       z-index: 10;
       border-bottom-left-radius: 0px;
       border-top-left-radius: 0px;
   }
   /* 列表 */
   .table {
       margin-top: 10px;
       width: 100%;
   }
    .table .new{
        width: 20px;
        height:20px;
        position:absolute;
        top:0px;
        left:0px;
        z-index: 10;
        border-width:20px 20px 0px 0px;
        border-style:solid;
        border-color:#fbc900 transparent transparent;
    }
    .table .new-zi{
        color: #ffffff;
        position: absolute;
        top: -2px;
        left: 1px;
        z-index: 11;
    }
    .colsetting{
        font-size: 20px;
        padding-top: 15px;
        padding-right: 10px;
    }
</style>
`
	VueIndexColModifyTPL = `
                        for (let i in list) {{{indexModifyRows}}
                        }
`
	VueIndexListColumnTPL = `
                    {title: "{{fieldComment}}", field: '{{fieldName}}', show: true},`
	VueIndexSelectOptionTPL = `
                    { value: '{{fieldName}}', label: '{{fieldComment}}' },`
	VueCreateComponentTPL = `<template>
    <v-modal class="model" title="{{tbName}}" :width='540' :visible="show" @cancel="ruleCancel">
        <v-form direction="horizontal" :model="customForm" :rules="customRules" ref="customRuleForm"  @keyup.enter.native="submitForm('customRuleForm')">
            {{fromField}}

            <div class="layer-button">
                <v-button type="primary" @click="submitForm('customRuleForm')" :loading="this.$store.state.loading">{{
                    this.$store.state.loading ?
                    "正在发送中" : "确认" }}
                </v-button>
                <v-button @click="ruleCancel">取消</v-button>
            </div>
        </v-form>
    </v-modal>
</template>

<script>
  import {hostName} from '../../config/api';

  const createApi = hostName + "v1/{{pageUrl}}";

  export default {
      data() {
          return {
              customForm: {{{customField}}
              },
              show: false,
              customRules:{{{customRules}}
              },
              labelCol: {
                  span: 6
              },
              wrapperCol: {
                  span: 14
              }
          }
      },
      methods: {
		  submitForm: function (formName) {
              this.$refs[formName].validate((valid) => {
                  if(valid) {
					  let params = this.customForm;{{createSubmitDataFix}}
					  this.$store.state.loading     = true;
					  this.$http.post(createApi, this.$qs.parse(params)).then(resp=> {
						  this.$store.state.loading = false;
						  if (resp.data.status == 1) {
							  this.$notification.success({
								  message    : '提示',
								  duration   : 2,
								  description: "创建成功"
							  });
							  this.ruleCancel();
							  this.$emit('refreshList');
						  }
					  });
                  }
              });
          },
          //取消
          ruleCancel: function () {
              for(var i in this.customForm) {
                  this.customForm[i] = '';
              }
              this.show = false;
          }
      }
  }
</script>

<style scoped>
    .text-input {
        margin: 10px auto;
        width: 80%;
        text-align: center;
    }
</style>
`
	VueCreateFieldComponentTPL = ` 
			<v-form-item label="{{fieldComment}}" :label-col="labelCol" :wrapper-col="wrapperCol" prop="{{fieldName}}" has-feedback>
                <v-input v-model="customForm.{{fieldName}}" size="large"></v-input>
            </v-form-item>
`
	VueCreateCustomFormComponentTPL = `
                          {{fieldName}}  : '{{fieldDefault}}',`
	VueCreateCustomRulesComponentTPL = `
                  {{fieldName}}:[
                      {required: {{required}}, message: '请输入{{fieldComment}}', length: {{length}}, type: "{{type}}"}
                  ],`

	vueEditComponentTPL = `<template>
    <v-modal class="add-user"  :title=" updateMode ? '编辑' : '详情' " :width='640' :visible="show" @cancel="ruleCancel">
        <v-form direction="horizontal"  v-bind:class="{ 'view-mode': !updateMode }" :model="customForm" :rules="customRules" ref="customRuleForm" @keyup.enter.native="submitForm('customRuleForm')">
            {{fromField}}
            <div class="layer-button">
                <v-button v-if="updateMode" type="primary" style="margin-right:10px" @click.prevent="submitForm('customRuleForm')" :loading="loading">{{ loading ? "正在修改中" : "马上修改" }}</v-button>
                <v-button type="ghost" @click.prevent="ruleCancel()">{{ updateMode ? "取消" : "关闭" }}</v-button>
            </div>
        </v-form>
    </v-modal>
</template>

<script>
  import {hostName} from '../../config/api'

  const UpdateAPI = hostName + "v1/{{pageUrl}}";

  export default {
      data() {
          return {
              {{tbPk}}  : '',
              show      : false,
              loading   : false,
              updateMode: false,
              customForm: {{{customField}}
              },
              customRules:{{{customRules}}
              },
              labelCol: {
                  span: 6
              },
              wrapperCol: {
                  span: 14
              }
          }
      },
      methods: {
          submitForm: function (formName) {
              this.$refs[formName].validate((valid) => {
                  if(valid) {
                      let params = {{{editSubmitItems}}
                      };

                      this.loading     = true;
                      this.$http.put(UpdateAPI + "/" + this.customForm.{{tbPk}}, this.$qs.parse(params)).then(resp => {
                          this.loading = false;
                          if (resp.data.status == 1) {
                              this.$notification.success({
                                  message    : '提示',
                                  duration   : 2,
                                  description: resp.data.status_txt
                              });
                              this.ruleCancel();
                              this.$emit('refreshList');
                          }
                      });
                  }else{
                      alert(valid);
                  }
              });
          },
          //取消
          ruleCancel: function () {
              this.show = false;
              this.loading = false;
              this.customForm = {};
          }
      }
  }
</script>

<style scoped>
    .text-input {
        margin: 10px auto;
        width: 80%;
        text-align: center;
    }
    .view-mode .ant-form-item{
        margin-bottom: 0px;
    }
    .view-mode .ant-form-item-required:before{
        display:none;
    }
</style>
`
	vueEditComponentFromItemTPL = `
            <v-form-item label="{{fieldComment}}" :label-col="labelCol" :wrapper-col="wrapperCol" prop="{{fieldName}}" has-feedback>
                <v-input v-if="updateMode"  v-model="customForm.{{fieldName}}" size="large" {{disabled}}></v-input>
                <span v-if="!updateMode"  class="ant-form-text">{{customForm.{{fieldName}}}}</span>
            </v-form-item>
`
	vueEditComponentSubmitItemTPL = `
                          {{fieldName}}  : this.customForm.{{fieldName}},`
	vueColSettingComponentTPL = `
<template>
    <v-modal class="model" title="显示" :width='220' :visible="show" @cancel="ruleCancel">
        <v-form direction="horizontal">
            <ul>
                <template v-for="item in columnsSetting">
                   <li> <v-checkbox v-model="item.show" :true-value="true" :false-value="false">{{item.title}}</v-checkbox></li>
                </template>
            </ul>


            <div class="layer-button">
                <v-button @click="ruleCancel">完成</v-button>
            </div>
        </v-form>
    </v-modal>
</template>

<script>
  export default {
      data() {
          return {
              show: false,
              labelCol: {
                  span: 6
              },
              wrapperCol: {
                  span: 14
              }
          }
      },
      props: ['columnsSetting'],
      methods: {
          //取消
          ruleCancel: function () {
              var cols = [];
              for (var i=0;i<this.columnsSetting.length;i++)
              {
                  if(this.columnsSetting[i].show){
                      cols.push(this.columnsSetting[i]);
                  }
              }
              this.$emit('setCols', cols);
              this.show = false;
          }
      }
  }
</script>

<style scoped>
    li {
        margin-bottom: 5px;
    }
</style>

`
	operateListTPL = `
	operateList = append(operateList, RoleRight{
		RightName:   "{{ctrlName}}-list",
		RightAction: "[GET]/{{pageUrl}}",
		FrontURL:    "{{pageUrl}}",
	})
	operateList = append(operateList, RoleRight{
		RightName:   "{{ctrlName}}-create",
		RightAction: "[POST]/{{pageUrl}}",
		FrontURL:    "{{pageUrl}}",
	})
	operateList = append(operateList, RoleRight{
		RightName:   "{{ctrlName}}-update",
		RightAction: "[PUT]/{{pageUrl}}",
		FrontURL:    "{{pageUrl}}",
	})
	operateList = append(operateList, RoleRight{
		RightName:   "{{ctrlName}}-delete",
		RightAction: "[DELETE]/{{pageUrl}}",
		FrontURL:    "{{pageUrl}}",
	})
`
	vueRuleTPL = `
              {
                  path: '/{{pageUrl}}/index',
                  component: name => require(['../components/{{filPath}}/Index'], name),
              },`
	menuListTPL = `
                    {"name":"{{ctrlName}}","url":"/{{pageUrl}}/index","icon":"bars"},`
)

package mysqldump_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/jamf/go-mysqldump"
)

const expected = `-- Go SQL Dump ` + mysqldump.Version + `
--
-- ------------------------------------------------------
-- Server version	test_version

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES UTF8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table ~Test_Table~
--

DROP TABLE IF EXISTS ~Test_Table~;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE 'Test_Table' (
	~id~ int(11) NOT NULL AUTO_INCREMENT,
	~email~ varchar(255) NOT NULL,
	~given_name~ varchar(127) NOT NULL DEFAULT '',
	~surname~ varchar(127) NOT NULL DEFAULT '',
	~name~ varchar(255) GENERATED ALWAYS AS (CONCAT(given_name,' ',surname)),
	~int8~ TINYINT NOT NULL,
	~NullInt8~ TINYINT,
	~uint8~ TINYINT UNSIGNED NOT NULL,
	~NullUint8~ TINYINT UNSIGNED,
	~int16~ SMALLINT NOT NULL,
	~NullInt16~ SMALLINT,
	~uint16~ SMALLINT UNSIGNED NOT NULL,
	~NullUint16~ SMALLINT UNSIGNED,
	~int32~ INT(11) NOT NULL,
	~NullInt32~ INT(11),
	~uint32~ INT(11) UNSIGNED NOT NULL,
	~NullUint32~ INT(11) UNSIGNED,
	~int64~ BIGINT NOT NULL,
	~NullInt64~ BIGINT,
	~uint64~ BIGINT UNSIGNED NOT NULL,
	~float32~ FLOAT NOT NULL,
	~NullFloat32~ FLOAT,
	~float64~ DOUBLE NOT NULL,
	~NullFloat64~ DOUBLE,
	~bool~ TINYINT(1) NOT NULL,
	~NullBool~ TINYINT(1),
	~time~ DATETIME NOT NULL,
	~NullTime~ DATETIME,
	~varbinary~ VARBINARY,
	~rawbytes~ BLOB,
	PRIMARY KEY (~id~)
)ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table ~Test_Table~
--

LOCK TABLES ~Test_Table~ WRITE;
/*!40000 ALTER TABLE ~Test_Table~ DISABLE KEYS */;
INSERT INTO ~Test_Table~ (~id~, ~email~, ~given_name~, ~surname~, ~int8~, ~NullInt8~, ~uint8~, ~NullUint8~, ~int16~, ~NullInt16~, ~uint16~, ~NullUint16~, ~int32~, ~NullInt32~, ~uint32~, ~NullUint32~, ~int64~, ~NullInt64~, ~uint64~, ~float32~, ~NullFloat32~, ~float64~, ~NullFloat64~, ~bool~, ~NullBool~, ~time~, ~NullTime~, ~varbinary~, ~rawbytes~) VALUES (1,'test1@test.de','Test','Name 1',1,NULL,1,NULL,1,NULL,1,NULL,1,NULL,1,NULL,1,NULL,1,1.000000,NULL,1.000000,NULL,1,NULL,'1970-01-01 00:00:00',NULL,NULL,NULL),(2,'test2@test.de',NULL,'Test Name 2',2,NULL,2,NULL,2,NULL,2,NULL,2,NULL,2,NULL,2,NULL,2,2.000000,NULL,2.000000,NULL,1,NULL,'1970-01-01 00:00:00',NULL,NULL,NULL);
/*!40000 ALTER TABLE ~Test_Table~ ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

`

func mockColumnRows() *sqlmock.Rows {
	col1 := sqlmock.NewColumn("Field").OfType("VARCHAR", "").Nullable(true)
	col2 := sqlmock.NewColumn("Type").OfType("TEXT", "").Nullable(true)
	col3 := sqlmock.NewColumn("Null").OfType("VARCHAR", "").Nullable(true)
	col4 := sqlmock.NewColumn("Key").OfType("ENUM", "").Nullable(true)
	col5 := sqlmock.NewColumn("Default").OfType("TEXT", "").Nullable(true)
	col6 := sqlmock.NewColumn("Extra").OfType("VARCHAR", "").Nullable(true)
	return sqlmock.NewRowsWithColumnDefinition(col1, col2, col3, col4, col5, col6).
		AddRow("id", "int(11)", "NO", "PRI", nil, "auto_increment").
		AddRow("email", "varchar(255)", "NO", "", nil, "").
		AddRow("given_name", "varchar(127)", "NO", "", "", "").
		AddRow("surname", "varchar(127)", "NO", "", "", "").
		AddRow("name", "varchar(255)", "YES", "", nil, "VIRTUAL GENERATED").
		AddRow("int8", "TINYINT", "NO", "", nil, "").
		AddRow("NullInt8", "TINYINT", "YES", "", nil, "").
		AddRow("uint8", "TINYINT UNSIGNED", "NO", "", nil, "").
		AddRow("NullUint8", "TINYINT UNSIGNED", "YES", "", nil, "").
		AddRow("int16", "SMALLINT", "NO", "", nil, "").
		AddRow("NullInt16", "SMALLINT", "YES", "", nil, "").
		AddRow("uint16", "SMALLINT UNSIGNED", "NO", "", nil, "").
		AddRow("NullUint16", "SMALLINT UNSIGNED", "YES", "", nil, "").
		AddRow("int32", "INT(11)", "NO", "", nil, "").
		AddRow("NullInt32", "INT(11)", "YES", "", nil, "").
		AddRow("uint32", "INT(11) UNSIGNED", "NO", "", nil, "").
		AddRow("NullUint32", "INT(11) UNSIGNED", "YES", "", nil, "").
		AddRow("int64", "BIGINT", "NO", "", nil, "").
		AddRow("NullInt64", "BIGINT", "YES", "", nil, "").
		AddRow("uint64", "BIGINT UNSIGNED", "NO", "", nil, "").
		AddRow("float32", "FLOAT", "NO", "", nil, "").
		AddRow("NullFloat32", "FLOAT", "YES", "", nil, "").
		AddRow("float64", "DOUBLE", "NO", "", nil, "").
		AddRow("NullFloat64", "DOUBLE", "YES", "", nil, "").
		AddRow("bool", "BOOL", "NO", "", nil, "").
		AddRow("NullBool", "BOOL", "YES", "", nil, "").
		AddRow("time", "DATETIME", "NO", "", nil, "").
		AddRow("NullTime", "DATETIME", "YES", "", nil, "").
		AddRow("varbinary", "VARBINARY", "YES", "", nil, "").
		AddRow("rawbytes", "BLOB", "YES", "", nil, "")
}

func c(name string, v interface{}) *sqlmock.Column {
	var t string
	var nullable bool
	switch v.(type) {
	case string:
		t = "VARCHAR"
	case sql.NullString:
		nullable = true
		t = "VARCHAR"
	case int8:
		t = "TINYINT"
	case int16:
		t = "SMALLINT"
	case sql.NullInt16:
		nullable = true
		t = "SMALLINT"
	case int32:
		t = "INT(11)"
	case sql.NullInt32:
		nullable = true
		t = "INT(11)"
	case int64:
		t = "BIGINT"
	case sql.NullInt64:
		nullable = true
		t = "BIGINT"
	case int:
		t = "BIGINT"
	case uint8:
		t = "TINYINT UNSIGNED"
	case uint16:
		t = "SMALLINT UNSIGNED"
	case uint32:
		t = "INT UNSIGNED"
	case uint64:
		t = "BIGINT UNSIGNED"
	case uint:
		t = "BIGINT UNSIGNED"
	case float32:
		t = "FLOAT"
	case float64:
		t = "DOUBLE"
	case sql.NullFloat64:
		nullable = true
		t = "DOUBLE"
	case bool:
		t = "BOOL"
	case sql.NullBool:
		nullable = true
		t = "BOOL"
	case time.Time:
		t = "DATETIME"
	case sql.NullTime:
		nullable = true
		t = "DATETIME"
	case []byte:
		nullable = true
		t = "VARBINARY"
	case sql.RawBytes:
		nullable = true
		t = "BLOB"
	default:
		panic(fmt.Errorf("unknown value type: %T", v))
	}
	return sqlmock.NewColumn(name).OfType(t, v).Nullable(nullable)
}

func RunDump(t testing.TB, data *mysqldump.Data) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer db.Close()

	data.Connection = db
	showTablesRows := sqlmock.NewRowsWithColumnDefinition(c("Tables_in_Testdb", "")).
		AddRow("Test_Table")

	showColumnsRows := mockColumnRows()

	serverVersionRows := sqlmock.NewRowsWithColumnDefinition(c("Version()", "")).
		AddRow("test_version")

	createTableRows := sqlmock.NewRowsWithColumnDefinition(c("Table", ""), c("Create Table", "")).
		AddRow("Test_Table", strings.ReplaceAll(`CREATE TABLE 'Test_Table' (
	~id~ int(11) NOT NULL AUTO_INCREMENT,
	~email~ varchar(255) NOT NULL,
	~given_name~ varchar(127) NOT NULL DEFAULT '',
	~surname~ varchar(127) NOT NULL DEFAULT '',
	~name~ varchar(255) GENERATED ALWAYS AS (CONCAT(given_name,' ',surname)),
	~int8~ TINYINT NOT NULL,
	~NullInt8~ TINYINT,
	~uint8~ TINYINT UNSIGNED NOT NULL,
	~NullUint8~ TINYINT UNSIGNED,
	~int16~ SMALLINT NOT NULL,
	~NullInt16~ SMALLINT,
	~uint16~ SMALLINT UNSIGNED NOT NULL,
	~NullUint16~ SMALLINT UNSIGNED,
	~int32~ INT(11) NOT NULL,
	~NullInt32~ INT(11),
	~uint32~ INT(11) UNSIGNED NOT NULL,
	~NullUint32~ INT(11) UNSIGNED,
	~int64~ BIGINT NOT NULL,
	~NullInt64~ BIGINT,
	~uint64~ BIGINT UNSIGNED NOT NULL,
	~float32~ FLOAT NOT NULL,
	~NullFloat32~ FLOAT,
	~float64~ DOUBLE NOT NULL,
	~NullFloat64~ DOUBLE,
	~bool~ TINYINT(1) NOT NULL,
	~NullBool~ TINYINT(1),
	~time~ DATETIME NOT NULL,
	~NullTime~ DATETIME,
	~varbinary~ VARBINARY,
	~rawbytes~ BLOB,
	PRIMARY KEY (~id~)
)ENGINE=InnoDB DEFAULT CHARSET=latin1`, "~", "`"))

	createTableValueRows := sqlmock.NewRowsWithColumnDefinition(
		c("id", int32(0)),
		c("email", ""),
		c("given_name", sql.NullString{}),
		c("surname", sql.NullString{}),
		c("int8", int8(0)),
		c("NullInt8", sql.NullInt16{}),
		c("uint8", uint8(0)),
		c("NullUint8", sql.NullInt16{}),
		c("int16", int16(0)),
		c("NullInt16", sql.NullInt16{}),
		c("uint16", uint16(0)),
		c("NullUint16", sql.NullInt32{}),
		c("int32", int32(0)),
		c("NullInt32", sql.NullInt32{}),
		c("uint32", uint32(0)),
		c("NullUint32", sql.NullInt64{}),
		c("int64", int64(0)),
		c("NullInt64", sql.NullInt64{}),
		c("uint64", uint64(0)),
		c("float32", float32(0)),
		c("NullFloat32", sql.NullFloat64{}),
		c("float64", float64(0)),
		c("NullFloat64", sql.NullFloat64{}),
		c("bool", false),
		c("NullBool", sql.NullBool{}),
		c("time", time.Time{}),
		c("NullTime", sql.NullTime{}),
		c("varbinary", []byte{}),
		c("rawbytes", sql.RawBytes{}),
	).
		AddRow(
			int32(1),
			"test1@test.de",
			"Test",
			"Name 1",
			int8(1),
			sql.NullInt16{},
			uint8(1),
			sql.NullInt16{},
			int16(1),
			sql.NullInt16{},
			uint16(1),
			sql.NullInt32{},
			int32(1),
			sql.NullInt32{},
			uint32(1),
			sql.NullInt64{},
			int64(1),
			sql.NullInt64{},
			uint64(1),
			float32(1),
			sql.NullFloat64{},
			float64(1),
			sql.NullFloat64{},
			true,
			sql.NullBool{},
			time.Unix(0, 0),
			sql.NullTime{},
			[]byte{},
			sql.RawBytes{},
		).
		AddRow(
			int32(2),
			"test2@test.de",
			nil,
			"Test Name 2",
			int8(2),
			sql.NullInt16{},
			uint8(2),
			sql.NullInt16{},
			int16(2),
			sql.NullInt16{},
			uint16(2),
			sql.NullInt32{},
			int32(2),
			sql.NullInt32{},
			uint32(2),
			sql.NullInt64{},
			int64(2),
			sql.NullInt64{},
			uint64(2),
			float32(2),
			sql.NullFloat64{},
			float64(2),
			sql.NullFloat64{},
			true,
			sql.NullBool{},
			time.Unix(0, 0),
			sql.NullTime{},
			[]byte{},
			sql.RawBytes{},
		)

	mock.ExpectBegin()
	mock.ExpectQuery(`^SELECT version\(\)$`).WillReturnRows(serverVersionRows)
	mock.ExpectQuery(`^SHOW TABLES$`).WillReturnRows(showTablesRows)
	mock.ExpectExec("^LOCK TABLES `Test_Table` READ /\\*!32311 LOCAL \\*/$").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("^SHOW CREATE TABLE `Test_Table`$").WillReturnRows(createTableRows)
	mock.ExpectQuery("^SHOW COLUMNS FROM `Test_Table`$").WillReturnRows(showColumnsRows)
	mock.ExpectQuery("^SELECT (.+) FROM `Test_Table`$").WillReturnRows(createTableValueRows)
	mock.ExpectExec("UNLOCK TABLES")
	mock.ExpectRollback()

	assert.NoError(t, data.Dump(), "an error was not expected when dumping a stub database connection")

	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")
}

func TestDumpOk(t *testing.T) {
	var buf bytes.Buffer

	RunDump(t, &mysqldump.Data{
		Out:        &buf,
		LockTables: true,
	})

	result := strings.Replace(strings.Split(buf.String(), "-- Dump completed")[0], "`", "~", -1)

	assert.Equal(t, expected, result)
}

func TestNoLockOk(t *testing.T) {
	var buf bytes.Buffer

	data := &mysqldump.Data{
		Out:        &buf,
		LockTables: false,
	}

	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer db.Close()

	data.Connection = db
	showTablesRows := sqlmock.NewRowsWithColumnDefinition(c("Tables_in_Testdb", "")).
		AddRow("Test_Table")

	showColumnsRows := mockColumnRows()

	serverVersionRows := sqlmock.NewRowsWithColumnDefinition(c("Version()", "")).
		AddRow("test_version")

	createTableRows := sqlmock.NewRowsWithColumnDefinition(c("Table", ""), c("Create Table", "")).
		AddRow("Test_Table", strings.ReplaceAll(`CREATE TABLE 'Test_Table' (
	~id~ int(11) NOT NULL AUTO_INCREMENT,
	~email~ varchar(255) NOT NULL,
	~given_name~ varchar(127) NOT NULL DEFAULT '',
	~surname~ varchar(127) NOT NULL DEFAULT '',
	~name~ varchar(255) GENERATED ALWAYS AS (CONCAT(given_name,' ',surname)),
	~int8~ TINYINT NOT NULL,
	~NullInt8~ TINYINT,
	~uint8~ TINYINT UNSIGNED NOT NULL,
	~NullUint8~ TINYINT UNSIGNED,
	~int16~ SMALLINT NOT NULL,
	~NullInt16~ SMALLINT,
	~uint16~ SMALLINT UNSIGNED NOT NULL,
	~NullUint16~ SMALLINT UNSIGNED,
	~int32~ INT(11) NOT NULL,
	~NullInt32~ INT(11),
	~uint32~ INT(11) UNSIGNED NOT NULL,
	~NullUint32~ INT(11) UNSIGNED,
	~int64~ BIGINT NOT NULL,
	~NullInt64~ BIGINT,
	~uint64~ BIGINT UNSIGNED NOT NULL,
	~float32~ FLOAT NOT NULL,
	~NullFloat32~ FLOAT,
	~float64~ DOUBLE NOT NULL,
	~NullFloat64~ DOUBLE,
	~bool~ TINYINT(1) NOT NULL,
	~NullBool~ TINYINT(1),
	~time~ DATETIME NOT NULL,
	~NullTime~ DATETIME,
	~varbinary~ VARBINARY,
	~rawbytes~ BLOB,
	PRIMARY KEY (~id~)
)ENGINE=InnoDB DEFAULT CHARSET=latin1`, "~", "`"))

	createTableValueRows := sqlmock.NewRowsWithColumnDefinition(
		c("id", int32(0)),
		c("email", ""),
		c("given_name", sql.NullString{}),
		c("surname", sql.NullString{}),
		c("int8", int8(0)),
		c("NullInt8", sql.NullInt16{}),
		c("uint8", uint8(0)),
		c("NullUint8", sql.NullInt16{}),
		c("int16", int16(0)),
		c("NullInt16", sql.NullInt16{}),
		c("uint16", uint16(0)),
		c("NullUint16", sql.NullInt32{}),
		c("int32", int32(0)),
		c("NullInt32", sql.NullInt32{}),
		c("uint32", uint32(0)),
		c("NullUint32", sql.NullInt64{}),
		c("int64", int64(0)),
		c("NullInt64", sql.NullInt64{}),
		c("uint64", uint64(0)),
		c("float32", float32(0)),
		c("NullFloat32", sql.NullFloat64{}),
		c("float64", float64(0)),
		c("NullFloat64", sql.NullFloat64{}),
		c("bool", false),
		c("NullBool", sql.NullBool{}),
		c("time", time.Time{}),
		c("NullTime", sql.NullTime{}),
		c("varbinary", []byte{}),
		c("rawbytes", sql.RawBytes{}),
	).
		AddRow(
			int32(1),
			"test1@test.de",
			"Test",
			"Name 1",
			int8(1),
			sql.NullInt16{},
			uint8(1),
			sql.NullInt16{},
			int16(1),
			sql.NullInt16{},
			uint16(1),
			sql.NullInt32{},
			int32(1),
			sql.NullInt32{},
			uint32(1),
			sql.NullInt64{},
			int64(1),
			sql.NullInt64{},
			uint64(1),
			float32(1),
			sql.NullFloat64{},
			float64(1),
			sql.NullFloat64{},
			true,
			sql.NullBool{},
			time.Unix(0, 0),
			sql.NullTime{},
			[]byte{},
			sql.RawBytes{},
		).
		AddRow(
			int32(2),
			"test2@test.de",
			nil,
			"Test Name 2",
			int8(2),
			sql.NullInt16{},
			uint8(2),
			sql.NullInt16{},
			int16(2),
			sql.NullInt16{},
			uint16(2),
			sql.NullInt32{},
			int32(2),
			sql.NullInt32{},
			uint32(2),
			sql.NullInt64{},
			int64(2),
			sql.NullInt64{},
			uint64(2),
			float32(2),
			sql.NullFloat64{},
			float64(2),
			sql.NullFloat64{},
			true,
			sql.NullBool{},
			time.Unix(0, 0),
			sql.NullTime{},
			[]byte{},
			sql.RawBytes{},
		)

	mock.ExpectBegin()
	mock.ExpectQuery(`^SELECT version\(\)$`).WillReturnRows(serverVersionRows)
	mock.ExpectQuery(`^SHOW TABLES$`).WillReturnRows(showTablesRows)
	mock.ExpectQuery("^SHOW CREATE TABLE `Test_Table`$").WillReturnRows(createTableRows)
	mock.ExpectQuery("^SHOW COLUMNS FROM `Test_Table`$").WillReturnRows(showColumnsRows)
	mock.ExpectQuery("^SELECT (.+) FROM `Test_Table`$").WillReturnRows(createTableValueRows)
	mock.ExpectRollback()

	assert.NoError(t, data.Dump(), "an error was not expected when dumping a stub database connection")

	result := strings.Replace(strings.Split(buf.String(), "-- Dump completed")[0], "`", "~", -1)

	assert.Equal(t, expected, result)

	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")
}

func BenchmarkDump(b *testing.B) {
	data := &mysqldump.Data{
		Out:        io.Discard,
		LockTables: true,
	}
	for i := 0; i < b.N; i++ {
		RunDump(b, data)
	}
}
